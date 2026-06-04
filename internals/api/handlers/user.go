package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/luqmanshaban/gomont/internals/config"
	"github.com/luqmanshaban/gomont/internals/store"
)

type UserHandler struct {
	Store *store.UserStore
	Cfg *config.Config
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("failed to request body", "error", err)
		WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}
	
	if payload.Email == "" {
		WriteJson(w, http.StatusBadRequest, map[string]string{"message": "email field is required"})
		return
	}

	code, err := h.Store.CreateUser(payload.Email)
	if err != nil {
		slog.Error("failed to signup with email", "error", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to create user", "error": err.Error()})
		return
	}

	err = SendEmail(h.Cfg, payload.Email, code)
	if err != nil {
		slog.Error("failed to send email email", "error", err)
		WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to send verification code to email, please try again later"})
		return
	}

	WriteJson(w, http.StatusOK, map[string]any{"message": "check your inbox for next steps"})
}