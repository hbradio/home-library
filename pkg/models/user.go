package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Auth0ID   string    `json:"auth0_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetOrCreateUser(db *sql.DB, auth0ID, email string) (*User, error) {
	var u User
	err := db.QueryRow(
		`SELECT id, auth0_id, email, created_at, updated_at FROM users WHERE auth0_id = $1`,
		auth0ID,
	).Scan(&u.ID, &u.Auth0ID, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err == nil {
		return &u, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	err = db.QueryRow(
		`INSERT INTO users (auth0_id, email) VALUES ($1, $2)
		 RETURNING id, auth0_id, email, created_at, updated_at`,
		auth0ID, email,
	).Scan(&u.ID, &u.Auth0ID, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
