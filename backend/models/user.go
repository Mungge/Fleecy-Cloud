package models

import (
	"database/sql"
	"time"

	"github.com/Mungge/Fleecy-Cloud/config"
)

var db *sql.DB

func init() {
	var err error
	db, err = config.Connect()
	if err != nil {
		panic(err)
	}
}

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func CreateUser(user User) error {
	query := `
		INSERT INTO users (name, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	return db.QueryRow(
		query,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)
}

func GetUserByEmail(email string) (User, error) {
	var user User
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = $1`

	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return User{}, err
	}

	return user, err
}

