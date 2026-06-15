package worker

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/luqmanshaban/gomont/internals/core"
	"github.com/luqmanshaban/gomont/internals/store"
)

type Worker struct {
	Store *store.URLStore
}

func (w Worker) Worker(workerId int, jobs <-chan core.URL) {

	for job := range jobs {
		err := Ping(job.Endpoint)
		if err != nil {
			slog.Error("failed to ping url", "worker_id", workerId, "url", job.Endpoint, "error", err)
			// TODO: Add exponential backoff / mark as unhealthy in DB here
			continue
		}
		// update url status to healthy
		err = w.Store.UpdateURLHealthStatusTrue(job.ID, job.UserID, job.Interval)
		if err != nil {
			slog.Error("Error on worker component","worker_id", workerId, "error", err.Error())
			panic(err)
		}
		slog.Info("successfully updated url status to healthy", "url", job.Endpoint)
	}
}

func Ping(url string) error {
	client := http.Client{Timeout: 10 * time.Second}

	res, err := client.Get(url)
	if err != nil {
		slog.Error("failed to ping url", "url", url, "error", err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("The endpoint returned an error status: ", res.StatusCode)
}
