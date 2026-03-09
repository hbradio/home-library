package models

import (
	"database/sql"
	"strings"
	"time"
)

type Book struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	ISBN        string     `json:"isbn"`
	Title       string     `json:"title"`
	Author      string     `json:"author"`
	Genre       string     `json:"genre"`
	Publisher   string     `json:"publisher"`
	PublishYear *int       `json:"publish_year"`
	CreatedAt   time.Time  `json:"created_at"`
	LoanStatus  string     `json:"loan_status"`
	PatronName  *string    `json:"patron_name,omitempty"`
}

func CreateBook(db *sql.DB, userID, isbn, title, author, genre, publisher string, publishYear *int) (*Book, error) {
	var b Book
	var py sql.NullInt64
	if publishYear != nil {
		py = sql.NullInt64{Int64: int64(*publishYear), Valid: true}
	}
	err := db.QueryRow(
		`INSERT INTO books (user_id, isbn, title, author, genre, publisher, publish_year)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, user_id, isbn, title, author, genre, publisher, publish_year, created_at`,
		userID, isbn, title, author, genre, publisher, py,
	).Scan(&b.ID, &b.UserID, &b.ISBN, &b.Title, &b.Author, &b.Genre, &b.Publisher, &b.PublishYear, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	b.LoanStatus = "available"
	return &b, nil
}

func GetBooks(db *sql.DB, userID, title, author, genre string) ([]Book, error) {
	query := `SELECT b.id, b.user_id, b.isbn, b.title, b.author, b.genre, COALESCE(b.publisher, ''), b.publish_year, b.created_at,
		COALESCE(
			(SELECT le.event_type FROM loan_events le WHERE le.book_id = b.id ORDER BY le.created_at DESC LIMIT 1),
			'available'
		) as loan_status,
		(SELECT p.first_name || ' ' || p.last_name FROM loan_events le
		 JOIN patrons p ON p.id = le.patron_id
		 WHERE le.book_id = b.id ORDER BY le.created_at DESC LIMIT 1) as patron_name
		FROM books b WHERE b.user_id = $1`
	args := []interface{}{userID}
	argN := 2

	if title != "" {
		query += ` AND LOWER(b.title) LIKE $` + itoa(argN)
		args = append(args, "%"+strings.ToLower(title)+"%")
		argN++
	}
	if author != "" {
		query += ` AND LOWER(b.author) LIKE $` + itoa(argN)
		args = append(args, "%"+strings.ToLower(author)+"%")
		argN++
	}
	if genre != "" {
		query += ` AND LOWER(b.genre) LIKE $` + itoa(argN)
		args = append(args, "%"+strings.ToLower(genre)+"%")
		argN++
	}
	query += ` ORDER BY b.title ASC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		var patronName sql.NullString
		err := rows.Scan(&b.ID, &b.UserID, &b.ISBN, &b.Title, &b.Author, &b.Genre, &b.Publisher, &b.PublishYear, &b.CreatedAt, &b.LoanStatus, &patronName)
		if err != nil {
			return nil, err
		}
		if b.LoanStatus == "checkout" {
			b.LoanStatus = "checked_out"
			if patronName.Valid {
				b.PatronName = &patronName.String
			}
		} else {
			b.LoanStatus = "available"
		}
		books = append(books, b)
	}
	return books, nil
}

func GetBookByID(db *sql.DB, bookID, userID string) (*Book, error) {
	var b Book
	var patronName sql.NullString
	err := db.QueryRow(
		`SELECT b.id, b.user_id, b.isbn, b.title, b.author, b.genre, COALESCE(b.publisher, ''), b.publish_year, b.created_at,
		COALESCE(
			(SELECT le.event_type FROM loan_events le WHERE le.book_id = b.id ORDER BY le.created_at DESC LIMIT 1),
			'available'
		) as loan_status,
		(SELECT p.first_name || ' ' || p.last_name FROM loan_events le
		 JOIN patrons p ON p.id = le.patron_id
		 WHERE le.book_id = b.id ORDER BY le.created_at DESC LIMIT 1) as patron_name
		FROM books b WHERE b.id = $1 AND b.user_id = $2`,
		bookID, userID,
	).Scan(&b.ID, &b.UserID, &b.ISBN, &b.Title, &b.Author, &b.Genre, &b.Publisher, &b.PublishYear, &b.CreatedAt, &b.LoanStatus, &patronName)
	if err != nil {
		return nil, err
	}
	if b.LoanStatus == "checkout" {
		b.LoanStatus = "checked_out"
		if patronName.Valid {
			b.PatronName = &patronName.String
		}
	} else {
		b.LoanStatus = "available"
	}
	return &b, nil
}

func GetBookByISBN(db *sql.DB, isbn, userID string) (*Book, error) {
	var b Book
	var patronName sql.NullString
	err := db.QueryRow(
		`SELECT b.id, b.user_id, b.isbn, b.title, b.author, b.genre, COALESCE(b.publisher, ''), b.publish_year, b.created_at,
		COALESCE(
			(SELECT le.event_type FROM loan_events le WHERE le.book_id = b.id ORDER BY le.created_at DESC LIMIT 1),
			'available'
		) as loan_status,
		(SELECT p.first_name || ' ' || p.last_name FROM loan_events le
		 JOIN patrons p ON p.id = le.patron_id
		 WHERE le.book_id = b.id ORDER BY le.created_at DESC LIMIT 1) as patron_name
		FROM books b WHERE b.isbn = $1 AND b.user_id = $2`,
		isbn, userID,
	).Scan(&b.ID, &b.UserID, &b.ISBN, &b.Title, &b.Author, &b.Genre, &b.Publisher, &b.PublishYear, &b.CreatedAt, &b.LoanStatus, &patronName)
	if err != nil {
		return nil, err
	}
	if b.LoanStatus == "checkout" {
		b.LoanStatus = "checked_out"
		if patronName.Valid {
			b.PatronName = &patronName.String
		}
	} else {
		b.LoanStatus = "available"
	}
	return &b, nil
}

func DeleteBook(db *sql.DB, bookID, userID string) error {
	_, err := db.Exec(`DELETE FROM loan_events WHERE book_id = $1`, bookID)
	if err != nil {
		return err
	}
	_, err = db.Exec(`DELETE FROM books WHERE id = $1 AND user_id = $2`, bookID, userID)
	return err
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
