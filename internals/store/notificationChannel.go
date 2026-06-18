package store

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/lib/pq" // Required for pq.Array
	"github.com/luqmanshaban/gomont/internals/core"
)

type NotificationChannelStore struct {
	DB *sql.DB
}

func NewNotificationChannelStore(db *sql.DB) *NotificationChannelStore {
	return &NotificationChannelStore{DB: db}
}

// AddEmail appends unique emails to the existing array
func (s *NotificationChannelStore) AddEmail(newEmails []string, rowID, userID int) (core.NotificationChannel, error) {
	var emailList core.NotificationChannel
	// Using array_cat and set-logic to ensure uniqueness (requires pg 9.5+)
	// Or simpler: array_append in a loop, but here is a batch approach:
	query := `
		UPDATE notification_channels
		SET emails = (SELECT array_agg(distinct e) FROM unnest(emails || $1) e)
		WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, emails, created_at`

	err := s.DB.QueryRow(query, pq.Array(newEmails), rowID, userID).Scan(
		&emailList.ID,
		&emailList.UserID,
		pq.Array(&emailList.Emails),
		&emailList.CreatedAt,
	)
	if err != nil {
		slog.Error("failed to add emails", "error", err)
		return core.NotificationChannel{}, err
	}
	return emailList, nil
}

// UpdateEmail searches for an old email in the array and replaces it with a new one
func (s *NotificationChannelStore) UpdateEmail(rowID int, oldEmail, newEmail string, userID int) (core.NotificationChannel, error) {
	var emailList core.NotificationChannel
	query := `
		UPDATE notification_channels
		SET emails = array_replace(emails, $1, $2)
		WHERE id = $3 AND user_id = $4
		RETURNING id, user_id, emails, created_at
		`

	err := s.DB.QueryRow(query, oldEmail, newEmail, rowID, userID).Scan(
		&emailList.ID,
		&emailList.UserID,
		pq.Array(&emailList.Emails),
		&emailList.CreatedAt,
	)
	if err != nil {
		slog.Error("failed to update email in array", "error", err)
		return core.NotificationChannel{},err
	}
	return emailList, nil
}

// DeleteEmail removes a specific string from the array
func (s *NotificationChannelStore) DeleteEmailFromChannel(rowID, userID int, emailToRemove string) (core.NotificationChannel,error) {
	var emailList core.NotificationChannel
	// array_remove removes all occurrences of the element
	query := `
	UPDATE notification_channels 
	SET emails = array_remove(emails, $1) 
	WHERE id = $2 AND user_id = $3
	RETURNING id, user_id, emails, created_at
	`

	err := s.DB.QueryRow(query, emailToRemove, rowID, userID).Scan(
		&emailList.ID,
		&emailList.UserID,
		pq.Array(&emailList.Emails),
		&emailList.CreatedAt,
	)
	if err != nil {
		slog.Error("failed to remove email from array", "error", err)
		return core.NotificationChannel{},err
	}
	return emailList,nil
}

// GetEmailsByUserID retrieves the email record for a specific user
func (s *NotificationChannelStore) GetEmailsByID(rowId, userId int) (core.NotificationChannel, error) {
	var emailList core.NotificationChannel
	query := `SELECT id, user_id, emails, created_at FROM notification_channels WHERE id = $1 AND user_id = $2`

	err := s.DB.QueryRow(query, rowId, userId).Scan(
		&emailList.ID,
		&emailList.UserID,
		pq.Array(&emailList.Emails),
		&emailList.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("email list is empty", "error", err)
			return core.NotificationChannel{}, err // Or handle as error depending on your logic
		}
		return core.NotificationChannel{}, err
	}
	return emailList, nil
}

func (s *NotificationChannelStore) GetEmailsByUserID(userID int) (core.NotificationChannel, error) {
	var emailList core.NotificationChannel
	query := `SELECT id, user_id, emails, created_at FROM notification_channels WHERE user_id = $1`

	err := s.DB.QueryRow(query, userID).Scan(
		&emailList.ID,
		&emailList.UserID,
		pq.Array(&emailList.Emails),
		&emailList.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("email list is empty", "error", err)
			return core.NotificationChannel{}, err // Or handle as error depending on your logic
		}
		return core.NotificationChannel{}, err
	}
	return emailList, nil
}
