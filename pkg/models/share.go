package models

import (
	"database/sql"
	"time"
)

type Share struct {
	ID              string    `json:"id"`
	OwnerID         string    `json:"owner_id"`
	SharedWithEmail string    `json:"shared_with_email"`
	CreatedAt       time.Time `json:"created_at"`
}

type SharedLibrary struct {
	ShareID    string `json:"share_id"`
	OwnerID    string `json:"owner_id"`
	OwnerEmail string `json:"owner_email"`
}

func CreateShare(db *sql.DB, ownerID, email string) (*Share, error) {
	var s Share
	err := db.QueryRow(
		`INSERT INTO library_shares (owner_id, shared_with_email)
		 VALUES ($1, $2)
		 RETURNING id, owner_id, shared_with_email, created_at`,
		ownerID, email,
	).Scan(&s.ID, &s.OwnerID, &s.SharedWithEmail, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func GetSharesByOwner(db *sql.DB, ownerID string) ([]Share, error) {
	rows, err := db.Query(
		`SELECT id, owner_id, shared_with_email, created_at
		 FROM library_shares WHERE owner_id = $1 ORDER BY created_at DESC`,
		ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shares []Share
	for rows.Next() {
		var s Share
		if err := rows.Scan(&s.ID, &s.OwnerID, &s.SharedWithEmail, &s.CreatedAt); err != nil {
			return nil, err
		}
		shares = append(shares, s)
	}
	return shares, nil
}

func DeleteShare(db *sql.DB, shareID, ownerID string) error {
	_, err := db.Exec(
		`DELETE FROM library_shares WHERE id = $1 AND owner_id = $2`,
		shareID, ownerID,
	)
	return err
}

func GetLibrariesSharedWithEmail(db *sql.DB, email string) ([]SharedLibrary, error) {
	rows, err := db.Query(
		`SELECT ls.id, ls.owner_id, u.email
		 FROM library_shares ls
		 JOIN users u ON u.id = ls.owner_id
		 WHERE ls.shared_with_email = $1
		 ORDER BY ls.created_at DESC`,
		email,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var libs []SharedLibrary
	for rows.Next() {
		var l SharedLibrary
		if err := rows.Scan(&l.ShareID, &l.OwnerID, &l.OwnerEmail); err != nil {
			return nil, err
		}
		libs = append(libs, l)
	}
	return libs, nil
}

func HasShareAccess(db *sql.DB, ownerID, viewerEmail string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM library_shares WHERE owner_id = $1 AND shared_with_email = $2)`,
		ownerID, viewerEmail,
	).Scan(&exists)
	return exists, err
}
