package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"home-library/pkg/db"
	"home-library/pkg/models"
)

func PublicLibrary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, `{"error":"slug parameter required"}`, http.StatusBadRequest)
		return
	}

	pool, err := db.GetDB()
	if err != nil {
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	link, err := models.GetPublicLinkBySlug(pool, slug)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"library not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error":"failed to look up library"}`, http.StatusInternalServerError)
		return
	}

	title := r.URL.Query().Get("title")
	author := r.URL.Query().Get("author")
	genre := r.URL.Query().Get("genre")

	books, err := models.GetBooks(pool, link.UserID, title, author, genre)
	if err != nil {
		http.Error(w, `{"error":"failed to get books"}`, http.StatusInternalServerError)
		return
	}
	if books == nil {
		books = []models.Book{}
	}
	json.NewEncoder(w).Encode(books)
}
