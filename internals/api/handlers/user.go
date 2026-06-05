package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/luqmanshaban/gomont/internals/api/utils"
	"github.com/luqmanshaban/gomont/internals/config"
	"github.com/luqmanshaban/gomont/internals/store"
)


type UserHandler struct {
	Store *store.UserStore
	Cfg   *config.Config
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("failed to request body", "error", err)
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}

	if payload.Email == "" {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "email field is required"})
		return
	}

	exists, err := h.Store.IsUserExist(payload.Email)
	if err != nil {
		slog.Error("failed to check if user exists", "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to signup user, try again later"})
		return
	}

	if exists {
		code, err := h.Store.LoginUser(payload.Email)
		if err != nil {
			slog.Error("failed to Login existing user", "email", payload.Email, "error", err)
			utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "Failed to login user, please try again later"})
			return
		}
		err = utils.SendEmail(h.Cfg, payload.Email, code)
		if err != nil {
			slog.Error("failed to send login verification email", "email", payload.Email, "error", err)
			utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to send verification code", "error": err.Error()})
			return
		}
		utils.WriteJson(w, http.StatusOK, map[string]any{"message": "check your inbox for next steps"})
		return
	}

	code, err := h.Store.CreateUser(payload.Email)
	if err != nil {
		slog.Error("failed to signup with email", "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to create user", "error": err.Error()})
		return
	}

	err = utils.SendEmail(h.Cfg, payload.Email, code)
	if err != nil {
		slog.Error("failed to send email email", "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to send verification code to email, please try again later"})
		return
	}

	utils.WriteJson(w, http.StatusOK, map[string]any{"message": "check your inbox for next steps"})
}

// internals/api/handlers/user.go

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// 1. Extract the claims from the context injected by AuthMiddleware
	claims, ok := r.Context().Value(UserClaimsContextKey).(*utils.CustomClaims)
	if !ok {
		slog.Error("failed to assert context claims to *utils.CustomClaims")
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "internal authentication error"})
		return
	}

	// 2. Fetch the latest user details from the database using the ID from the token
	user, err := h.Store.GetUserById(claims.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteJson(w, http.StatusNotFound, map[string]string{"message": "user profile not found"})
			return
		}
		slog.Error("failed to retrieve user profile", "user_id", claims.ID, "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to fetch profile"})
		return
	}

	// 3. Scrub or format any sensitive data if needed before returning 
	// For example, returning a safe response structure:
	response := map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"names":      user.Names,
		"created_at": user.CreatedAt,
	}

	// 4. Return the data to the client
	utils.WriteJson(w, http.StatusOK, response)
}


// VerifyLogin handles checking the login OTP and returning a JWT token
func (h *UserHandler) VerifyLogin(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
		Code  int    `json:"code"`
	}

	// 1. Decode payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("failed to decode verify body", "error", err)
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}

	// 2. Validate input
	if payload.Email == "" || payload.Code == 0 {
		utils.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "email and code fields are required"})
		return
	}

	// 3. Call the refactored store method that returns (*core.User, error)
	user, err := h.Store.IsUserLoginCodeValid(payload.Email, payload.Code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteJson(w, http.StatusUnauthorized, map[string]string{"message": "invalid email or verification code"})
			return
		}
		
		// If you explicitly returned a "verification code expired" error from the store:
		if err.Error() == "verification code expired" {
			utils.WriteJson(w, http.StatusUnauthorized, map[string]string{"message": "your verification code has expired"})
			return
		}

		slog.Error("unexpected error verifying login code", "email", payload.Email, "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to verify login code"})
		return
	}

	// 4. Generate JWT using the user details retrieved from the database check
	token, err := utils.GenerateJWT(h.Cfg.JWT_SECRET, user.Email, user.ID) // Adjust this if your GenerateToken function takes user.ID or user.Names
	if err != nil {
		slog.Error("failed to generate auth token", "email", user.Email, "error", err)
		utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "failed to complete authentication"})
		return
	}

	// 5. Send token back to client
	utils.WriteJson(w, http.StatusOK, map[string]string{
		"token":   token,
		"message": "login successful",
	})
}