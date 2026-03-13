package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"home-library/pkg/auth"
	"home-library/pkg/db"
	"home-library/pkg/models"
)

func SharedLibraries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

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

	ownerID := r.URL.Query().Get("owner_id")

	if ownerID != "" {
		// Return books from a specific owner's library
		hasAccess, err := models.HasShareAccess(pool, ownerID, strings.ToLower(user.Email))
		if err != nil {
			http.Error(w, `{"error":"access check failed"}`, http.StatusInternalServerError)
			return
		}
		if !hasAccess {
			http.Error(w, `{"error":"access denied"}`, http.StatusForbidden)
			return
		}

		bookID := r.URL.Query().Get("book_id")
		if bookID != "" {
			book, err := models.GetBookByID(pool, bookID, ownerID)
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
		books, err := models.GetBooks(pool, ownerID, title, authorQ, genre)
		if err != nil {
			http.Error(w, `{"error":"failed to get books"}`, http.StatusInternalServerError)
			return
		}
		if books == nil {
			books = []models.Book{}
		}
		json.NewEncoder(w).Encode(books)
		return
	}

	// Return list of libraries shared with this user
	libs, err := models.GetLibrariesSharedWithEmail(pool, strings.ToLower(user.Email))
	if err != nil {
		http.Error(w, `{"error":"failed to get shared libraries"}`, http.StatusInternalServerError)
		return
	}
	if libs == nil {
		libs = []models.SharedLibrary{}
	}
	json.NewEncoder(w).Encode(libs)
}
