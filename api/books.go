package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"home-library/pkg/auth"
	"home-library/pkg/db"
	"home-library/pkg/models"
)

func Books(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sub, accessToken, err := auth.ValidateRequest(r)
	if err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	pool, err := db.GetDB()
	if err != nil {
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	email, _ := auth.FetchUserEmail(accessToken)
	user, err := models.GetOrCreateUser(pool, sub, email)
	if err != nil {
		http.Error(w, `{"error":"failed to get user"}`, http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Single book by ID
		id := r.URL.Query().Get("id")
		if id != "" {
			book, err := models.GetBookByID(pool, id, user.ID)
			if err != nil {
				http.Error(w, `{"error":"book not found"}`, http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(book)
			return
		}

		title := r.URL.Query().Get("title")
		authorQ := r.URL.Query().Get("author")
		genre := r.URL.Query().Get("genre")
		books, err := models.GetBooks(pool, user.ID, title, authorQ, genre)
		if err != nil {
			http.Error(w, `{"error":"failed to get books"}`, http.StatusInternalServerError)
			return
		}
		if books == nil {
			books = []models.Book{}
		}
		json.NewEncoder(w).Encode(books)

	case http.MethodPost:
		var body struct {
			ISBN        string `json:"isbn"`
			Title       string `json:"title"`
			Author      string `json:"author"`
			Genre       string `json:"genre"`
			Publisher   string `json:"publisher"`
			PublishYear *int   `json:"publish_year"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if body.ISBN == "" || body.Title == "" {
			http.Error(w, `{"error":"isbn and title are required"}`, http.StatusBadRequest)
			return
		}
		book, err := models.CreateBook(pool, user.ID, body.ISBN, body.Title, body.Author, body.Genre, body.Publisher, body.PublishYear)
		if err != nil {
			if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
				http.Error(w, `{"error":"book already in library"}`, http.StatusConflict)
				return
			}
			http.Error(w, `{"error":"failed to add book"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(book)

	case http.MethodDelete:
		bookID := r.URL.Query().Get("id")
		if bookID == "" {
			http.Error(w, `{"error":"id parameter required"}`, http.StatusBadRequest)
			return
		}
		if err := models.DeleteBook(pool, bookID, user.ID); err != nil {
			http.Error(w, `{"error":"failed to delete book"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
