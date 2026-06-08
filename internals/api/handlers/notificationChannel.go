package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/luqmanshaban/gomont/internals/api/utils"
	"github.com/luqmanshaban/gomont/internals/store"
)

type NotificationChannelHandler struct {
	Store *store.NotificationStore
}

func (h *NotificationChannelHandler) GetNotificationChannelsById(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsContextKey).(*utils.CustomClaims)
	if !ok {
		slog.Error("failed to assert context claims to *utils.CustomClaims")
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal authentication error"})
		return
	}
	rowIdStr := r.PathValue("row_id")
	if rowIdStr == "" {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "row id is not provided in the path"})
		return
	}

	rowId, err := strconv.Atoi(rowIdStr)
	if err != nil {
		slog.Error("Invalid ID format: Must be an integer", "error", err)
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "Invalid ID format: Must be an integer"})
		return
	}

	channels, err := h.Store.GetEmailsByID(rowId, claims.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteJson(w, http.StatusNotFound, map[string]string{"message": "No records found for the user"})
			return
		}
		slog.Error("failed to fetch for notification channels", "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal server error, please try again later"})
		return
	}


	utils.WriteJson(w, http.StatusOK, map[string]any {
		"id": channels.ID,
		"emails": channels.Emails,
		"created_at": channels.CreatedAt,
	})
}

func (h *NotificationChannelHandler) GetNotificationChannelsByUserId(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsContextKey).(*utils.CustomClaims)
	if !ok {
		slog.Error("failed to assert context claims to *utils.CustomClaims")
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal authentication error"})
		return
	}

	channels, err := h.Store.GetEmailsByUserID(claims.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteJson(w, http.StatusNotFound, map[string]string{"message": "No records found for the user"})
			return
		}
		slog.Error("failed to fetch for notification channels", "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal server error, please try again later"})
		return
	}


	utils.WriteJson(w, http.StatusOK, map[string]any {
		"id": channels.ID,
		"emails": channels.Emails,
		"created_at": channels.CreatedAt,
	})
}

func (h *NotificationChannelHandler) AddNotificationChannels(w http.ResponseWriter, r *http.Request) {
	// 1. Extract context claims
	claims, ok := r.Context().Value(UserClaimsContextKey).(*utils.CustomClaims)
	if !ok {
		slog.Error("failed to assert context claims to *utils.CustomClaims")
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal authentication error"})
		return
	}

	// 2. Extract and parse row_id from the path
	rowIdStr := r.PathValue("row_id")
	if rowIdStr == "" {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "row id is not provided in the path"})
		return
	}

	rowId, err := strconv.Atoi(rowIdStr)
	if err != nil {
		slog.Error("Invalid ID format: Must be an integer", "error", err)
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "Invalid ID format: Must be an integer"})
		return
	}

	// 3. Define the request payload 
	// If you want to append multiple emails at once, use []string. If just one, use string.
	var payload struct {
		Emails []string `json:"emails"` 
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("failed to decode request body", "error", err)
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}

	// 4. Validate input payload
	if len(payload.Emails) == 0 {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "emails field is required and cannot be empty"})
		return
	}

	channels, err := h.Store.AddEmail(payload.Emails, rowId, claims.ID)
	if err != nil {
		slog.Error("failed to append emails to notification channel", "row_id", rowId, "user_id", claims.ID, "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to add notification channels"})
		return
	}

	// 6. Return response
	utils.WriteJson(w, http.StatusOK, map[string]any{
		"id":         channels.ID,
		"emails":     channels.Emails,
		"created_at": channels.CreatedAt,
	})
}

func (h *NotificationChannelHandler) UpdateNotificationChannels(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsContextKey).(*utils.CustomClaims)
	if !ok {
		slog.Error("failed to assert context claims to *utils.CustomClaims")
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal authentication error"})
		return
	}

	rowIdStr := r.PathValue("row_id")
	if rowIdStr == "" {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "row id is not provided in the path"})
		return
	}

	rowId, err := strconv.Atoi(rowIdStr); 
	if err != nil {
		slog.Error("Invalid ID format: Must be an integer", "error", err)
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message":"Invalid ID format: Must be an integer"})
		return
	}

	var payload struct {
		OldEmail string `json:"old_email"`
		NewEmail string `json:"new_email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("failed to request body", "error", err)
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}

	if payload.OldEmail == "" || payload.NewEmail == "" {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "old_email and payload_email fields are required"})
		return
	}

	channels, err := h.Store.UpdateEmail(rowId, payload.OldEmail, payload.NewEmail, claims.ID)

	utils.WriteJson(w, http.StatusOK, map[string]any {
		"id": channels.ID,
		"emails": channels.Emails,
		"created_at": channels.CreatedAt,
	})
}

func (h *NotificationChannelHandler) DeleteNotificationChannels(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsContextKey).(*utils.CustomClaims)
	if !ok {
		slog.Error("failed to assert context claims to *utils.CustomClaims")
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal authentication error"})
		return
	}

	rowIdStr := r.PathValue("row_id")
	if rowIdStr == "" {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "row id is not provided in the path"})
		return
	}

	rowId, err := strconv.Atoi(rowIdStr); 
	if err != nil {
		slog.Error("Invalid ID format: Must be an integer", "error", err)
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message":"Invalid ID format: Must be an integer"})
		return
	}

	var payload struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("failed to request body", "error", err)
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}

	if payload.Email == ""  {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "old_email and payload_email fields are required"})
		return
	}

	channels, err := h.Store.DeleteEmailFromChannel(rowId, claims.ID, payload.Email)

	utils.WriteJson(w, http.StatusAccepted, map[string]any {
		"id": channels.ID,
		"emails": channels.Emails,
		"created_at": channels.CreatedAt,
	})
}