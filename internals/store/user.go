package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/luqmanshaban/gomont/internals/core"
)

type UserStore struct {
	DB *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{DB: db}
}

func generateVerificationCode() int {
	return rand.IntN(900_000) + 100_000
}

func (u *UserStore) GetUserByEmail(email string) (core.User, error) {
	var user core.User
	if err := u.DB.QueryRow("SELECT FROM users (id, email, COALESCE(names, ''), created_at) WHERE email = $1", email).Scan(
		&user.ID,
		&user.Email,
		&user.Names,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("user not found", "email", email, "error", err)
			return core.User{}, sql.ErrNoRows
		}
		slog.Error("failed to fetch for user", "email", email, "error", err)
		return core.User{}, err
	}
	return user, nil
}

func (u *UserStore) GetUserById(id int) (core.User, error) {
	var user core.User
	if err := u.DB.QueryRow("SELECT id, email, COALESCE(names, ''), created_at FROM users WHERE id = $1", id).Scan(
		&user.ID,
		&user.Email,
		&user.Names,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("user not found", "id", id, "error", err)
			return core.User{}, sql.ErrNoRows
		}
		slog.Error("failed to fetch for user", "id", id, "error", err)
		return core.User{}, err
	}
	return user, nil
}

func (u *UserStore) UpdateUserById(id int, names string) (core.User, error) {
	var user core.User

	query := "UPDATE users SET names = $1 WHERE id = $2 RETURNING id, email, names, created_at"
	err := u.DB.QueryRow(query, names, id).Scan(
		&user.ID,
		&user.Email,
		&user.Names,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("user does not exist", "id", id, "error", err)
			return core.User{}, err
		}
		slog.Error("failed to update user names", "id", id, "error", err)
		return core.User{}, err
	}

	return user, nil
}

func (u *UserStore) DeleteUserById(id int) (bool, error) {

	query := "DELETE FROM users WHERE id = $1"
	res, err := u.DB.Exec(query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("user does not exist", "id", id, "error", err)
			return false, err
		}
		slog.Error("failed to delete user", "id", id, "error", err)
		return false, err
	}

	affectRows, err := res.RowsAffected()
	if err != nil {
		slog.Error("failed to delete user", "id", id, "error", err)
		return false, err
	}

	if affectRows == 0 {
		slog.Error("failed to delete user", "id", id, "error", errors.New("No Rows affected"))
		return false, errors.New("No Rows affected")
	}

	return true, nil
}

func (u *UserStore) CreateUser(email string) (int, error) {
	code := generateVerificationCode()
	expt := time.Now().Add(15 * time.Minute)

	_, err := u.DB.Exec("INSERT INTO users (email, verification_code, verification_expiry) VALUES ($1, $2, $3)", email, code, expt)
	if err != nil {
		return 0, err
	}
	return code, nil
}

func (u *UserStore) IsUserExist(email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	err := u.DB.QueryRow(query, email).Scan(&exists)
	if err != nil {
		slog.Error("failed to check if user exists", "email", email, "error", err)
		return false, err
	}

	return exists, nil
}

func (u *UserStore) LoginUser(email string) (int, error) {
	code := generateVerificationCode()
	expt := time.Now().Add(15 * time.Minute)

	res, err := u.DB.Exec("UPDATE users SET verification_code = $1, verification_expiry = $2 WHERE email = $3", code, expt, email)
	if err != nil {
		slog.Error("failed to save generated verification code for user", "email", email, "error", err)
		return 0, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		slog.Error("failed to insert verification code for user", "email", email, "error", err)
		return 0, err
	}

	if rowsAffected == 0 {
		slog.Error("failed to insert verification code for user", "email", email, "error", errors.New("No Rows affected"))
		return 0, errors.New("No Rows affected")
	}

	return code, nil
}

func (u *UserStore) IsUserLoginCodeValid(email string, code int) (*core.User, error) {
	var user core.User

	// 1. Fetch the user details to verify the code and expiry
	query := "SELECT id, email, verification_expiry, created_at FROM users WHERE email = $1 AND verification_code = $2"

	if err := u.DB.QueryRow(query, email, code).Scan(
		&user.ID,
		&user.Email,
		&user.VerificationExpiry,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("user not found or code incorrect", "email", email, "error", err)
			return nil, sql.ErrNoRows
		}
		slog.Error("failed to fetch for user", "email", email, "error", err)
		return nil, err
	}

	// 2. Check if the code has expired
	if user.VerificationExpiry.Before(time.Now()) {
		slog.Warn("user login code has expired", "email", email, "expiry", user.VerificationExpiry)
		return nil, fmt.Errorf("verification code expired")
	}

	// 3. SUCCESS! Clear the verification code and expiry from the DB immediately
	// Setting the code to NULL (or 0) and expiry to NULL (or a zero time) prevents reuse
	clearQuery := "UPDATE users SET verification_code = NULL, verification_expiry = NULL WHERE id = $1"
	if _, err := u.DB.Exec(clearQuery, user.ID); err != nil {
		// Log the error, but consider whether you want to block the login over a cleanup failure.
		// Usually, blocking is safer to prevent exploit loops.
		slog.Error("failed to clear verification code", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to finalize login sequence: %w", err)
	}

	// 4. Return the user info to the handler for JWT generation
	return &user, nil
}
