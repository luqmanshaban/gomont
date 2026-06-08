package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/luqmanshaban/gomont/internals/api/utils"
	"github.com/luqmanshaban/gomont/internals/core"
	"github.com/luqmanshaban/gomont/internals/store"
)

type URLHandler struct {
	Store *store.URLStore
}

// POST /urls
func (h *URLHandler) AddNewEndpoint(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsContextKey).(*utils.CustomClaims)
	if !ok {
		slog.Error("failed to assert context claims to *utils.CustomClaims")
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal authentication error"})
		return
	}

	var payload struct {
		Endpoint string `json:"endpoint"`
		Interval int    `json:"interval"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("failed to decode request body", "error", err)
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}

	if payload.Endpoint == "" || payload.Interval <= 0 {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "endpoint and valid positive interval are required"})
		return
	}

	newURL, err := h.Store.AddNewEndpoint(claims.ID, payload.Endpoint, payload.Interval)
	if err != nil {
		slog.Error("failed to save new monitor endpoint", "user_id", claims.ID, "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to add endpoint"})
		return
	}

	utils.WriteJson(w, http.StatusCreated, map[string]any{
		"id":          newURL.ID,
		"user_id":     newURL.UserID,
		"endpoint":    newURL.Endpoint,
		"isHealthy":   newURL.IsHealthy,
		"interval":    newURL.Interval,
		"max_retries": newURL.MaxRetries,
		"created_at":  newURL.CreatedAt,
	})
}

// PATCH /urls/{id}
func (h *URLHandler) UpdateURL(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsContextKey).(*utils.CustomClaims)
	if !ok {
		slog.Error("failed to assert context claims to *utils.CustomClaims")
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal authentication error"})
		return
	}

	idStr := r.PathValue("url_id")
	urlID, err := strconv.Atoi(idStr)
	if err != nil || urlID <= 0 {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid or missing URL ID path parameter"})
		return
	}

	var payload struct {
		Endpoint *string `json:"endpoint"`
		Interval *int    `json:"interval"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("failed to decode patch request body", "error", err)
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}

	existing, err := h.Store.GetURLByID(urlID, claims.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteJson(w, http.StatusNotFound, map[string]string{"message": "monitor not found"})
			return
		}
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal server error"})
		return
	}

	// Apply patches conditionally if provided in the JSON body
	if payload.Endpoint != nil {
		existing.Endpoint = *payload.Endpoint
	}
	if payload.Interval != nil {
		existing.Interval = *payload.Interval
	}

	updatedURL, err := h.Store.UpdateURL(urlID, claims.ID, existing.Endpoint, existing.Interval)
	if err != nil {
		slog.Error("failed updating endpoint record", "id", urlID, "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to update monitor"})
		return
	}

	utils.WriteJson(w, http.StatusOK, map[string]any{
		"id":          updatedURL.ID,
		"user_id":     updatedURL.UserID,
		"endpoint":    updatedURL.Endpoint,
		"isHealthy":   updatedURL.IsHealthy,
		"interval": updatedURL.Interval,
		"max_retries": updatedURL.MaxRetries,
		"updated_at":  updatedURL.UpdatedAt,
		"created_at":  updatedURL.CreatedAt,
	})
}

// GET /urls
func (h *URLHandler) GetURLsByUserId(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsContextKey).(*utils.CustomClaims)
	if !ok {
		slog.Error("failed to assert context claims to *utils.CustomClaims")
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal authentication error"})
		return
	}

	urls, err := h.Store.GetURLsByUserID(claims.ID)
	if err != nil {
		slog.Error("failed to fetch user endpoints list", "user_id", claims.ID, "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to load monitors"})
		return
	}

	// Return empty slice explicitly instead of null if user has no links yet
	if urls == nil {
		urls = []core.URL{}
	}

	utils.WriteJson(w, http.StatusOK, urls)
}

// GET /urls/{id}
func (h *URLHandler) GetURLById(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsContextKey).(*utils.CustomClaims)
	if !ok {
		slog.Error("failed to assert context claims to *utils.CustomClaims")
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal authentication error"})
		return
	}

	idStr := r.PathValue("url_id")
	urlID, err := strconv.Atoi(idStr)
	if err != nil || urlID <= 0 {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid URL ID format"})
		return
	}

	urlItem, err := h.Store.GetURLByID(urlID, claims.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteJson(w, http.StatusNotFound, map[string]string{"message": "monitor endpoint not found"})
			return
		}
		slog.Error("failed to retrieve endpoint metadata", "id", urlID, "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to fetch monitor details"})
		return
	}

	utils.WriteJson(w, http.StatusOK, map[string]any{
		"id":          urlItem.ID,
		"user_id":     urlItem.UserID,
		"endpoint":    urlItem.Endpoint,
		"isHealthy":   urlItem.IsHealthy,
		"interval":   urlItem.Interval,
		"max_retries": urlItem.MaxRetries,
		"updated_at":  urlItem.UpdatedAt,
		"created_at":  urlItem.CreatedAt,
	})
}

// DELETE /urls/{id}
func (h *URLHandler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(UserClaimsContextKey).(*utils.CustomClaims)
	if !ok {
		slog.Error("failed to assert context claims to *utils.CustomClaims")
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal authentication error"})
		return
	}

	idStr := r.PathValue("url_id")
	urlID, err := strconv.Atoi(idStr)
	if err != nil || urlID <= 0 {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid URL ID format"})
		return
	}

	err = h.Store.DeleteURL(urlID, claims.ID)
	if err != nil {
		if err.Error() == "url record not found or unauthorized" {
			utils.WriteJson(w, http.StatusNotFound, map[string]string{"message": "monitor not found or unauthorized"})
			return
		}
		slog.Error("failed to destroy endpoint row", "id", urlID, "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to delete monitor"})
		return
	}

	utils.WriteJson(w, http.StatusOK, map[string]string{"message": "monitor deleted successfully"})
}
