package store

import (
	"database/sql"
	"errors"
	"log/slog"
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
	runs_at := time.Now().Add(time.Duration(interval) * time.Minute)
	query := `
	INSERT INTO urls (user_id, endpoint, interval, runs_at)
	VALUES ($1, $2, $3, $4)
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
	runs_at := time.Now().Add(time.Duration(interval) * time.Minute)
	query := `
	    UPDATE urls
		SET
		runs_at = $1,
		updated_at = NOW(),
		is_healthy = true
		WHERE id = $2 AND user_id = $3
	`

	_,err := s.DB.Exec(query, runs_at, id, userId)
	if err != nil {
		slog.Error("failed to update url status to 'healthy'", "id", id, "error", err)
		return err
	}

	slog.Info("url status updated to 'healthy'", "id", id)
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
		SELECT id, user_id, endpoint, is_healthy, max_retries, interval, created_at
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
	// 1. Begin transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	// Always rollback if function exits before commit
	defer tx.Rollback()

	// 2. Query with locking
	// FOR UPDATE SKIP LOCKED ensures other workers skip rows currently being processed
	query := `
		SELECT id, user_id, endpoint, is_healthy, max_retries, interval, created_at
		FROM urls
		WHERE runs_at <= NOW()
		ORDER BY created_at DESC
		LIMIT 50
		FOR UPDATE SKIP LOCKED
	`

	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}

	var urls []core.URL
	var ids []int // Keep track to update their status to 'processing'

	for rows.Next() {
		var u core.URL
		if err := rows.Scan(
			&u.ID, &u.UserID, &u.Endpoint, &u.IsHealthy, 
			&u.MaxRetries, &u.Interval, &u.CreatedAt,
		); err != nil {
			rows.Close()
			return nil, err
		}
		urls = append(urls, u)
		ids = append(ids, u.ID)
	}
	rows.Close()

	if len(ids) == 0 {
		return urls, tx.Commit() // Nothing to update
	}

	// 3. Mark as 'processing' (or similar) so others don't pick it up
	// This prevents race conditions where two workers ping the same URL simultaneously
	_, err = tx.Exec(`
		UPDATE urls 
		SET last_checked = NOW() 
		WHERE id = ANY($1)
	`, pq.Array(ids))

	if err != nil {
		return nil, err
	}

	// 4. Commit transaction
	return urls, tx.Commit()
}