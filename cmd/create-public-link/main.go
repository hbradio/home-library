package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"home-library/pkg/db"
	"home-library/pkg/models"

	"github.com/joho/godotenv"
)

func main() {
	email := flag.String("email", "", "Email address of the user")
	slug := flag.String("slug", "", "Custom slug for the public link (e.g. 'my-library')")
	flag.Parse()

	if *email == "" || *slug == "" {
		fmt.Fprintf(os.Stderr, "Usage: create-public-link --email user@example.com --slug my-library\n")
		os.Exit(1)
	}

	*slug = strings.TrimSpace(strings.ToLower(*slug))
	if err := models.ValidateSlug(*slug); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid slug: %v\n", err)
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
	fmt.Printf("Found user: %s (ID: %s)\n", *email, userID)

	// Check if user already has a public link
	existing, err := models.GetPublicLinkByUser(database, userID)
	if err == nil {
		// User already has a link -- update the slug
		fmt.Printf("User already has a public link with slug: %s\n", existing.Slug)
		fmt.Printf("Updating slug to: %s\n", *slug)
		updated, err := models.UpdatePublicLinkSlug(database, userID, *slug)
		if err != nil {
			if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
				fmt.Fprintf(os.Stderr, "Slug %q is already taken by another user.\n", *slug)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Failed to update slug: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Done. Public link updated: /lib/%s\n", updated.Slug)
		return
	}

	// Create new public link
	link, err := models.CreatePublicLink(database, userID, *slug)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			fmt.Fprintf(os.Stderr, "Slug %q is already taken by another user.\n", *slug)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Failed to create public link: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Done. Public link created: /lib/%s\n", link.Slug)
}
