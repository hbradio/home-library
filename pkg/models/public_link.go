package models

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"regexp"
	"time"
)

type PublicLink struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

func GenerateSlug() (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		b[i] = chars[n.Int64()]
	}
	return string(b), nil
}

func ValidateSlug(slug string) error {
	if len(slug) < 3 || len(slug) > 30 {
		return fmt.Errorf("slug must be 3-30 characters")
	}
	if !slugPattern.MatchString(slug) {
		return fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens, and cannot start or end with a hyphen")
	}
	return nil
}

func CreatePublicLink(db *sql.DB, userID, slug string) (*PublicLink, error) {
	var link PublicLink
	err := db.QueryRow(
		`INSERT INTO public_links (user_id, slug)
		 VALUES ($1, $2)
		 RETURNING id, user_id, slug, created_at`,
		userID, slug,
	).Scan(&link.ID, &link.UserID, &link.Slug, &link.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func GetPublicLinkByUser(db *sql.DB, userID string) (*PublicLink, error) {
	var link PublicLink
	err := db.QueryRow(
		`SELECT id, user_id, slug, created_at FROM public_links WHERE user_id = $1`,
		userID,
	).Scan(&link.ID, &link.UserID, &link.Slug, &link.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func GetPublicLinkBySlug(db *sql.DB, slug string) (*PublicLink, error) {
	var link PublicLink
	err := db.QueryRow(
		`SELECT id, user_id, slug, created_at FROM public_links WHERE slug = $1`,
		slug,
	).Scan(&link.ID, &link.UserID, &link.Slug, &link.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func UpdatePublicLinkSlug(db *sql.DB, userID, newSlug string) (*PublicLink, error) {
	var link PublicLink
	err := db.QueryRow(
		`UPDATE public_links SET slug = $2 WHERE user_id = $1
		 RETURNING id, user_id, slug, created_at`,
		userID, newSlug,
	).Scan(&link.ID, &link.UserID, &link.Slug, &link.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func DeletePublicLink(db *sql.DB, userID string) error {
	_, err := db.Exec(`DELETE FROM public_links WHERE user_id = $1`, userID)
	return err
}
