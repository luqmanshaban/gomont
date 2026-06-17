package store

import (
	"database/sql"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	"github.com/lib/pq"
	"github.com/luqmanshaban/gomont/internals/core"
)

type URLStore struct {
	DB *sql.DB
}

func NewURLStore(db *sql.DB) *URLStore {
	return &URLStore{DB: db}
}

func (s *URLStore) AddNewEndpoint(userId int, endpoint string, interval int) (core.URL, error) {
	var newUrl core.URL
	runs_at := time.Now().UTC().Add(1 * time.Second)
	query := `
	INSERT INTO urls (user_id, endpoint, interval, runs_at, status, is_healthy)
	VALUES ($1, $2, $3, $4, 'pending', true)
	RETURNING id, user_id, endpoint, is_healthy, max_retries, interval, created_at
	`
	err := s.DB.QueryRow(query, userId, endpoint, interval, runs_at).Scan(
		&newUrl.ID,
		&newUrl.UserID,
		&newUrl.Endpoint,
		&newUrl.IsHealthy,
		&newUrl.MaxRetries,
		&newUrl.Interval,
		&newUrl.CreatedAt,
	)
	if err != nil {
		slog.Error("failed to add new url", "error", err)
		return core.URL{}, err
	}
	return newUrl, nil
}

func (s *URLStore) UpdateURL(id, userId int, endpoint string, interval int) (core.URL, error) {
	var updatedURL core.URL
	query := `
	    UPDATE urls
		SET
		endpoint = COALESCE($1, endpoint),
		interval = COALESCE($2, interval),
		updated_at = NOW()
		WHERE id = $3 AND user_id = $4
		RETURNING id, user_id, endpoint, is_healthy, max_retries, interval, created_at, updated_at
	`

	err := s.DB.QueryRow(query, endpoint, interval, id, userId).Scan(
		&updatedURL.ID,
		&updatedURL.UserID,
		&updatedURL.Endpoint,
		&updatedURL.IsHealthy,
		&updatedURL.MaxRetries,
		&updatedURL.Interval,
		&updatedURL.CreatedAt,
		&updatedURL.UpdatedAt,
	)
	if err != nil {
		slog.Error("failed to update url record", "id", id, "error", err)
		return core.URL{}, err
	}

	return updatedURL, nil
}

// updating url's health status
func (s *URLStore) UpdateURLHealthStatusTrue(id, userId int, interval int) error {
	runs_at := time.Now().UTC().Add(time.Duration(interval) * time.Minute)
	query := `
	    UPDATE urls
		SET
		runs_at = $1,
		updated_at = NOW(),
		status = 'pending',
		is_healthy = true,
		notification_sent = false,
		retries = 0
		WHERE id = $2 AND user_id = $3
	`

	_, err := s.DB.Exec(query, runs_at, id, userId)
	if err != nil {
		slog.Error("failed to update url status to 'healthy'", "id", id, "error", err)
		return err
	}

	slog.Info("url status updated to 'healthy'", "id", id)
	return nil
}

func (s *URLStore) UpdateURLHealthStatusFalse(id, userId int) error {
	runs_at := time.Now().UTC().UTC().Add(30 * time.Minute)
	query := `
	    UPDATE urls
		SET
		runs_at = $1,
		updated_at = NOW(),
		status = 'pending',
		is_healthy = false,
		notification_sent = true
		WHERE id = $2 AND user_id = $3
	`

	_, err := s.DB.Exec(query, runs_at, id, userId)
	if err != nil {
		slog.Error("failed to update url status to 'unhealthy'", "id", id, "error", err)
		return err
	}

	slog.Info("url marked as permanently unhealthy, rescheduled for background check in 30 mins", "id", id)
	return nil
}

// exponential backoff with jitter
func (r *URLStore) RetryURLPinging(id int, retries int) error {

	backoff := time.Duration(10<<retries) * time.Second
	jitter := time.Duration(rand.Intn(5)) * time.Second

	nextRunAt := time.Now().UTC().Add(backoff + jitter)

	_, err := r.DB.Exec(`
		UPDATE urls
		SET status='pending',
		retries = retries + 1,
		runs_at = $1,
		updated_at = NOW()
		where id = $2
		`, nextRunAt, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *URLStore) DeleteURL(id, userId int) error {
	query := `DELETE FROM urls WHERE id = $1 AND user_id = $2`

	res, err := s.DB.Exec(query, id, userId)
	if err != nil {
		slog.Error("failed to execute delete query on url", "id", id, "error", err)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		slog.Warn("no url rows deleted", "id", id, "user_id", userId)
		return errors.New("url record not found or unauthorized")
	}

	return nil
}

func (s *URLStore) GetURLsByUserID(userID int) ([]core.URL, error) {
	query := `
		SELECT id, user_id, endpoint, is_healthy, max_retries, interval,updated_at, created_at
		FROM urls
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.DB.Query(query, userID)
	if err != nil {
		slog.Error("failed to query user urls", "user_id", userID, "error", err)
		return nil, err
	}
	defer rows.Close() // Essential to prevent database connection leaks

	var urls []core.URL
	for rows.Next() {
		var u core.URL
		err := rows.Scan(
			&u.ID,
			&u.UserID,
			&u.Endpoint,
			&u.IsHealthy,
			&u.MaxRetries,
			&u.Interval,
			&u.UpdatedAt,
			&u.CreatedAt,
		)
		if err != nil {
			slog.Error("failed to scan url row", "error", err)
			return nil, err
		}
		urls = append(urls, u)
	}

	// Check for any execution errors encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (s *URLStore) GetURLByID(id, userID int) (core.URL, error) {
	var u core.URL
	query := `
		SELECT id, user_id, endpoint, is_healthy, max_retries, interval, created_at
		FROM urls
		WHERE id = $1 AND user_id = $2
	`

	err := s.DB.QueryRow(query, id, userID).Scan(
		&u.ID,
		&u.UserID,
		&u.Endpoint,
		&u.IsHealthy,
		&u.MaxRetries,
		&u.Interval,
		&u.CreatedAt,
	)
	if err != nil {
		slog.Error("failed to fetch url by id", "id", id, "error", err)
		return core.URL{}, err
	}

	return u, nil
}

func (s *URLStore) FetchURLsToPing() ([]core.URL, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// FIXED: Added u.email selection and INNER JOIN on users table
	query := `
		SELECT 
			urls.id, urls.user_id, users.email, urls.endpoint, 
			urls.is_healthy, urls.max_retries, urls.interval, 
			urls.created_at, urls.retries, urls.notification_sent
		FROM urls
		INNER JOIN users ON users.id = urls.user_id
		WHERE urls.runs_at <= NOW()
		AND urls.status = 'pending'
		ORDER BY urls.created_at DESC
		LIMIT 50
		FOR UPDATE SKIP LOCKED
	`

	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []core.URL
	var ids []int 

	for rows.Next() {
		var u core.URL
		// FIXED: Added &u.UserEmail to the scanner mapping array
		if err := rows.Scan(
			&u.ID, &u.UserID, &u.UserEmail, &u.Endpoint, 
			&u.IsHealthy, &u.MaxRetries, &u.Interval, 
			&u.CreatedAt, &u.Retries, &u.NotifcationSent,
		); err != nil {
			return nil, err
		}
		urls = append(urls, u)
		ids = append(ids, u.ID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return urls, tx.Commit() 
	}

	_, err = tx.Exec(`UPDATE urls SET status = 'processing' WHERE id = ANY($1)`, pq.Array(ids))
	if err != nil {
		return nil, err
	}

	return urls, tx.Commit()
}

func (s *URLStore) ResetStaleURLs() error {
	_, err := s.DB.Exec(`
		UPDATE urls
		SET status = 'pending'
		WHERE status = 'processing'
	`)
	return err
}

func (s *URLStore) ManuallyRetryURL(id, userID int) error {
	query := `
		UPDATE urls
		SET 
			status = 'pending',
			retries = 0,
			runs_at = NOW(),
			last_manual_retry_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`
	res, err := s.DB.Exec(query, id, userID)
	if err != nil {
		slog.Error("failed to execute manual retry query", "url_id", id, "error", err)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("url record not found or unauthorized")
	}

	return nil
}