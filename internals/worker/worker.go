package worker

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/luqmanshaban/gomont/internals/api/utils"
	"github.com/luqmanshaban/gomont/internals/config"
	"github.com/luqmanshaban/gomont/internals/core"
	"github.com/luqmanshaban/gomont/internals/sse"
	"github.com/luqmanshaban/gomont/internals/store"
)

type Worker struct {
	URLStore  *store.URLStore
	NotificationChannelStore *store.NotificationChannelStore
	Cfg    *config.Config
	Broker *sse.Broker
}

func (w Worker) Worker(workerId int, jobs <-chan core.URL) {
	for job := range jobs {
		err := Ping(job.Endpoint)
		if err != nil {
			slog.Error("failed to ping url", "worker_id", workerId, "url", job.Endpoint, "error", err)
			_ = w.Retry(job, err)
			continue
		}

		// Update url status to healthy
		err = w.URLStore.UpdateURLHealthStatusTrue(job.ID, job.UserID, job.Interval)
		if err != nil {
			slog.Error("Error updating health status in worker", "worker_id", workerId, "error", err.Error())
			panic(err) // Panic if DB is entirely unreachable
		}
		slog.Info("successfully updated url status to healthy", "url", job.Endpoint)

		// job.IsHealthy reflects status as of the producer's fetch, before
		// this check ran. Only publish when the check actually changed the
		// status (down -> healthy), not on every successful ping while
		// already healthy — keeps the dashboard stream meaningful instead
		// of noisy.
		if !job.IsHealthy && w.Broker != nil {
			w.Broker.Publish(sse.Event{
				MonitorID: job.ID,
				Name:      "status_change",
				Data: map[string]any{
					"id":         job.ID,
					"is_healthy": true,
					"updated_at": time.Now(),
				},
			})
		}
	}
}

func (w Worker) Retry(job core.URL, pingErr error) error {
	if job.Retries >= job.MaxRetries {
		// 1. Permanently update down status and stagger the next check time out by 30 mins
		err := w.URLStore.UpdateURLHealthStatusFalse(job.ID, job.UserID)
		if err != nil {
			slog.Error("failed to update failure state", "url_id", job.ID, "error", err)
			return err
		}

		// job.IsHealthy reflects status before this terminal failure was
		// recorded. Only publish on the actual healthy -> down transition,
		// not on every exhausted-retries call (which would already be
		// false on subsequent ping cycles while still down).
		if job.IsHealthy && w.Broker != nil {
			w.Broker.Publish(sse.Event{
				MonitorID: job.ID,
				Name:      "status_change",
				Data: map[string]any{
					"id":         job.ID,
					"is_healthy": false,
					"updated_at": time.Now(),
				},
			})
		}

		// 2. Strict Checklist Rule: Only dispatch notification if one hasn't been sent yet
		if !job.NotifcationSent {
			receipients, err := w.NotificationChannelStore.GetEmailsByUserID(job.UserID)
			if err != nil {
				slog.Error("failed to retrieve email recepients", "url_id", job.ID, "error", err)
				return err
			}
			if err = utils.SendNotificationEmail(w.Cfg, receipients.Emails, job.Endpoint, pingErr.Error(), time.Now()); err != nil {
				slog.Error("failed to send notification email", "url_id", job.ID, "error", err)
				return err
			}
			slog.Warn("Outage alert notification dispatched to user", "url_id", job.ID)
		} else {
			slog.Info("Alert suppressed; notification already delivered for this outage cycle", "url_id", job.ID)
		}

		return nil
	} else {
		// Increments retry values smoothly based on real database positions
		if err := w.URLStore.RetryURLPinging(job.ID, job.Retries); err != nil {
			slog.Error("failed to execute retry backoff sequence", "url_id", job.ID, "error", err)
			return err
		}
		slog.Info("url scheduled for retry", "url_id", job.ID, "next_attempt", job.Retries+1, "max", job.MaxRetries)
		return nil
	}
}

func Ping(url string) error {
	client := http.Client{Timeout: 10 * time.Second}

	res, err := client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("the endpoint returned an error status: %d", res.StatusCode)
}
