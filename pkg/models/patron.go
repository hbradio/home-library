package models

import (
	"database/sql"
	"strings"
	"time"
)

type Patron struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
}

type PatronDetail struct {
	Patron
	CheckedOutBooks []Book `json:"checked_out_books"`
}

func CreatePatron(db *sql.DB, userID, firstName, lastName string) (*Patron, error) {
	var p Patron
	err := db.QueryRow(
		`INSERT INTO patrons (user_id, first_name, last_name)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, first_name, last_name, created_at`,
		userID, firstName, lastName,
	).Scan(&p.ID, &p.UserID, &p.FirstName, &p.LastName, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func GetPatrons(db *sql.DB, userID, search string) ([]Patron, error) {
	query := `SELECT id, user_id, first_name, last_name, created_at FROM patrons WHERE user_id = $1`
	args := []interface{}{userID}

	if search != "" {
		query += ` AND (LOWER(first_name) LIKE $2 OR LOWER(last_name) LIKE $2 OR LOWER(first_name || ' ' || last_name) LIKE $2)`
		args = append(args, "%"+strings.ToLower(search)+"%")
	}
	query += ` ORDER BY first_name ASC, last_name ASC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patrons []Patron
	for rows.Next() {
		var p Patron
		if err := rows.Scan(&p.ID, &p.UserID, &p.FirstName, &p.LastName, &p.CreatedAt); err != nil {
			return nil, err
		}
		patrons = append(patrons, p)
	}
	return patrons, nil
}

func GetPatronByID(db *sql.DB, patronID, userID string) (*PatronDetail, error) {
	var p PatronDetail
	err := db.QueryRow(
		`SELECT id, user_id, first_name, last_name, created_at FROM patrons WHERE id = $1 AND user_id = $2`,
		patronID, userID,
	).Scan(&p.ID, &p.UserID, &p.FirstName, &p.LastName, &p.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Get checked out books (latest event is 'checkout')
	rows, err := db.Query(
		`SELECT b.id, b.user_id, b.isbn, b.title, b.author, b.genre, b.publish_year, b.created_at
		 FROM books b
		 WHERE b.user_id = $1
		 AND (SELECT le.event_type FROM loan_events le WHERE le.book_id = b.id ORDER BY le.created_at DESC LIMIT 1) = 'checkout'
		 AND (SELECT le.patron_id FROM loan_events le WHERE le.book_id = b.id ORDER BY le.created_at DESC LIMIT 1) = $2
		 ORDER BY b.title ASC`,
		userID, patronID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.UserID, &b.ISBN, &b.Title, &b.Author, &b.Genre, &b.PublishYear, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.LoanStatus = "checked_out"
		p.CheckedOutBooks = append(p.CheckedOutBooks, b)
	}
	return &p, nil
}

func DeletePatron(db *sql.DB, patronID, userID string) error {
	_, err := db.Exec(`DELETE FROM patrons WHERE id = $1 AND user_id = $2`, patronID, userID)
	return err
}
