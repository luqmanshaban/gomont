package store

import (
	"database/sql"
	"math/rand/v2"
	"time"
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
func (u *UserStore) CreateUser(email string) (int, error) {
	code := generateVerificationCode()
	expt := time.Now().Add(15 * time.Minute)

	_,err := u.DB.Exec("INSERT INTO users (email, verification_code, verification_expiry) VALUES ($1, $2, $3)", email, code,expt);
	if err != nil {
		return 0, err
	}
	return code, nil
}