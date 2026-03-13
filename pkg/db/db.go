package db

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	pool *sql.DB
	once sync.Once
	initErr error
)

func GetDB() (*sql.DB, error) {
	once.Do(func() {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			initErr = fmt.Errorf("DATABASE_URL not set")
			return
		}
		pool, initErr = sql.Open("pgx", dsn)
		if initErr != nil {
			return
		}
		initErr = pool.Ping()
		if initErr != nil {
			return
		}
		initErr = migrate(pool)
	})
	return pool, initErr
}

func migrate(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
			auth0_id TEXT UNIQUE NOT NULL,
			email TEXT,
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS books (
			id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id),
			isbn TEXT NOT NULL,
			title TEXT NOT NULL,
			author TEXT,
			genre TEXT,
			publish_year INT,
			created_at TIMESTAMPTZ DEFAULT now(),
			UNIQUE(user_id, isbn)
		)`,
		`CREATE TABLE IF NOT EXISTS patrons (
			id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id),
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS loan_events (
			id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
			book_id UUID NOT NULL REFERENCES books(id),
			patron_id UUID REFERENCES patrons(id),
			event_type TEXT NOT NULL CHECK (event_type IN ('checkout', 'return')),
			created_at TIMESTAMPTZ DEFAULT now()
		)`,
		`ALTER TABLE books ADD COLUMN IF NOT EXISTS publisher TEXT`,
		`ALTER TABLE books ADD COLUMN IF NOT EXISTS dewey_decimal TEXT`,
		`ALTER TABLE books ADD COLUMN IF NOT EXISTS lc_classification TEXT`,
		`CREATE INDEX IF NOT EXISTS idx_books_user_id ON books(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_patrons_user_id ON patrons(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_loan_events_book_id ON loan_events(book_id)`,
		`CREATE TABLE IF NOT EXISTS library_shares (
			id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
			owner_id UUID NOT NULL REFERENCES users(id),
			shared_with_email TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT now(),
			UNIQUE(owner_id, shared_with_email)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_library_shares_email ON library_shares(shared_with_email)`,
	}
	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}
