package store

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/luqmanshaban/gomont/internals/core"
)

type NotificationStore struct {
	DB *sql.DB
}

func NewNotificationStore(db *sql.DB) *NotificationStore {
	return &NotificationStore{DB: db}
}

func (s *NotificationStore) CreatNotification(notification core.Notification) (int, error) {
	var notfId int
	query := `
	INSERT INTO notifications (notification_channels_id, url_id, user_id, error_message)
	VALUES ($1, $2, $3, $4)
	RETURNING id
	`
	err := s.DB.QueryRow(query, notification.NotificationChannelID, notification.URLID, notification.UserID, notification.ErrorMsg).Scan(&notfId)
	if err != nil {
		slog.Error("failed to save notifcation", "error", err)
		return 0, err
	}

	return notfId, nil
}

func (s *NotificationStore) GetNotificationById(notificationId, userId int) (core.Notification, error) {
	var notification core.Notification
	query := `
	SELECT FROM notifications
	id, user_id, notification_channels_id, url_id, error_message, created_at
    WHERE id = $1 AND user_id = $2
	`
	err := s.DB.QueryRow(query, notificationId, userId).Scan(
		&notification.ID,
		&notification.NotificationChannelID,
		&notification.UserID,
		&notification.URLID,
		&notification.ErrorMsg,
		&notification.CreatedAt,
	)
	if err != nil {
		slog.Error("failed to fetch for notifcation", "error", err)
		return core.Notification{}, err
	}

	return notification, nil
}

func (s *NotificationStore) GetNotificationByUserId(userId int) (core.Notification, error) {
	var notification core.Notification
	query := `
	SELECT FROM notifications
	id, user_id, notification_channels_id, url_id, error_message, created_at
    WHERE user_id = $2
	`
	err := s.DB.QueryRow(query, userId).Scan(
		&notification.ID,
		&notification.NotificationChannelID,
		&notification.UserID,
		&notification.URLID,
		&notification.ErrorMsg,
		&notification.CreatedAt,
	)
	if err != nil {
		slog.Error("failed to fetch for notifcation", "error", err)
		return core.Notification{}, err
	}

	return notification, nil
}

func (u *NotificationStore) DeleteNotification(id, userId int) (bool, error) {

	query := "DELETE FROM notifications WHERE id = $1 AND user_id = $2"
	res, err := u.DB.Exec(query, id, userId)
	if err != nil {
		slog.Error("failed to delete notification", "id", id, "error", err)
		return false, err
	}

	affectRows, err := res.RowsAffected()
	if err != nil {
		slog.Error("failed to confirm notification deletion", "id", id, "error", err)
		return false, err
	}

	if affectRows == 0 {
		slog.Error("failed to delete user", "id", id, "error", "No Rows affected")
		return false, fmt.Errorf("No Rows affected")
	}

	return true, nil
}
