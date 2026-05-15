package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"home-library/pkg/auth"
	"home-library/pkg/db"
	"home-library/pkg/models"
)

func PublicLink(w http.ResponseWriter, r *http.Request) {
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
		link, err := models.GetPublicLinkByUser(pool, user.ID)
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"no public link"}`, http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, `{"error":"failed to get public link"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(link)

	case http.MethodPost:
		// Check if one already exists
		_, err := models.GetPublicLinkByUser(pool, user.ID)
		if err == nil {
			http.Error(w, `{"error":"public link already exists"}`, http.StatusConflict)
			return
		}

		slug, err := models.GenerateSlug()
		if err != nil {
			http.Error(w, `{"error":"failed to generate slug"}`, http.StatusInternalServerError)
			return
		}

		link, err := models.CreatePublicLink(pool, user.ID, slug)
		if err != nil {
			http.Error(w, `{"error":"failed to create public link"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(link)

	case http.MethodPut:
		var body struct {
			Slug string `json:"slug"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		body.Slug = strings.TrimSpace(strings.ToLower(body.Slug))

		if err := models.ValidateSlug(body.Slug); err != nil {
			json.NewEncoder(w)
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		link, err := models.UpdatePublicLinkSlug(pool, user.ID, body.Slug)
		if err != nil {
			if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
				http.Error(w, `{"error":"this slug is already taken"}`, http.StatusConflict)
				return
			}
			if err == sql.ErrNoRows {
				http.Error(w, `{"error":"no public link to update"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error":"failed to update slug"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(link)

	case http.MethodDelete:
		if err := models.DeletePublicLink(pool, user.ID); err != nil {
			http.Error(w, `{"error":"failed to delete public link"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
