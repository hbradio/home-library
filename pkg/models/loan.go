package models

import (
	"database/sql"
	"time"
)

type LoanEvent struct {
	ID         string    `json:"id"`
	BookID     string    `json:"book_id"`
	PatronID   *string   `json:"patron_id"`
	EventType  string    `json:"event_type"`
	CreatedAt  time.Time `json:"created_at"`
	PatronName *string   `json:"patron_name,omitempty"`
	BookTitle  *string   `json:"book_title,omitempty"`
}

func CreateLoanEvent(db *sql.DB, bookID string, patronID *string, eventType string) (*LoanEvent, error) {
	var e LoanEvent
	err := db.QueryRow(
		`INSERT INTO loan_events (book_id, patron_id, event_type)
		 VALUES ($1, $2, $3)
		 RETURNING id, book_id, patron_id, event_type, created_at`,
		bookID, patronID, eventType,
	).Scan(&e.ID, &e.BookID, &e.PatronID, &e.EventType, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func GetCurrentLoanStatus(db *sql.DB, bookID string) (string, *string, error) {
	var eventType string
	var patronID sql.NullString
	err := db.QueryRow(
		`SELECT event_type, patron_id FROM loan_events WHERE book_id = $1 ORDER BY created_at DESC LIMIT 1`,
		bookID,
	).Scan(&eventType, &patronID)
	if err == sql.ErrNoRows {
		return "available", nil, nil
	}
	if err != nil {
		return "", nil, err
	}
	if eventType == "checkout" {
		pid := patronID.String
		return "checked_out", &pid, nil
	}
	return "available", nil, nil
}

func GetLoanHistoryForBook(db *sql.DB, bookID string) ([]LoanEvent, error) {
	rows, err := db.Query(
		`SELECT le.id, le.book_id, le.patron_id, le.event_type, le.created_at,
		 p.first_name || ' ' || p.last_name as patron_name
		 FROM loan_events le
		 LEFT JOIN patrons p ON p.id = le.patron_id
		 WHERE le.book_id = $1
		 ORDER BY le.created_at DESC`,
		bookID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []LoanEvent
	for rows.Next() {
		var e LoanEvent
		var patronName sql.NullString
		if err := rows.Scan(&e.ID, &e.BookID, &e.PatronID, &e.EventType, &e.CreatedAt, &patronName); err != nil {
			return nil, err
		}
		if patronName.Valid {
			e.PatronName = &patronName.String
		}
		events = append(events, e)
	}
	return events, nil
}

func GetLoansByPatron(db *sql.DB, patronID string) ([]LoanEvent, error) {
	rows, err := db.Query(
		`SELECT le.id, le.book_id, le.patron_id, le.event_type, le.created_at,
		 b.title as book_title
		 FROM loan_events le
		 JOIN books b ON b.id = le.book_id
		 WHERE le.patron_id = $1
		 ORDER BY le.created_at DESC`,
		patronID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []LoanEvent
	for rows.Next() {
		var e LoanEvent
		var bookTitle sql.NullString
		if err := rows.Scan(&e.ID, &e.BookID, &e.PatronID, &e.EventType, &e.CreatedAt, &bookTitle); err != nil {
			return nil, err
		}
		if bookTitle.Valid {
			e.BookTitle = &bookTitle.String
		}
		events = append(events, e)
	}
	return events, nil
}
