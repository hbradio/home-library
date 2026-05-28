package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"home-library/pkg/db"

	"github.com/joho/godotenv"
)

func main() {
	email := flag.String("email", "", "Email address of the user")
	flag.Parse()

	if *email == "" {
		fmt.Fprintf(os.Stderr, "Usage: count-books --email user@example.com\n")
		os.Exit(1)
	}

	godotenv.Load()

	database, err := db.GetDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	// Look up user by email
	var userID string
	err = database.QueryRow(`SELECT id FROM users WHERE email = $1`, *email).Scan(&userID)
	if err == sql.ErrNoRows {
		fmt.Fprintf(os.Stderr, "No user found with email: %s\n", *email)
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to query user: %v\n", err)
		os.Exit(1)
	}

	var total, manual int
	err = database.QueryRow(
		`SELECT COUNT(*), COUNT(CASE WHEN isbn LIKE 'MANUAL-%' THEN 1 END) FROM books WHERE user_id = $1`,
		userID,
	).Scan(&total, &manual)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to count books: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User: %s\n", *email)
	fmt.Printf("Total books: %d\n", total)
	fmt.Printf("  Scanned (with ISBN): %d\n", total-manual)
	fmt.Printf("  Manual (no ISBN):    %d\n", manual)
}
