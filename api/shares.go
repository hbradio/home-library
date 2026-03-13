package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"home-library/pkg/auth"
	"home-library/pkg/db"
	"home-library/pkg/models"
)

func Shares(w http.ResponseWriter, r *http.Request) {
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
		shares, err := models.GetSharesByOwner(pool, user.ID)
		if err != nil {
			http.Error(w, `{"error":"failed to get shares"}`, http.StatusInternalServerError)
			return
		}
		if shares == nil {
			shares = []models.Share{}
		}
		json.NewEncoder(w).Encode(shares)

	case http.MethodPost:
		var body struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		body.Email = strings.TrimSpace(strings.ToLower(body.Email))
		if body.Email == "" {
			http.Error(w, `{"error":"email is required"}`, http.StatusBadRequest)
			return
		}
		if body.Email == strings.ToLower(user.Email) {
			http.Error(w, `{"error":"cannot share with yourself"}`, http.StatusBadRequest)
			return
		}
		share, err := models.CreateShare(pool, user.ID, body.Email)
		if err != nil {
			if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
				http.Error(w, `{"error":"already shared with this email"}`, http.StatusConflict)
				return
			}
			http.Error(w, `{"error":"failed to create share"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(share)

	case http.MethodDelete:
		shareID := r.URL.Query().Get("id")
		if shareID == "" {
			http.Error(w, `{"error":"id parameter required"}`, http.StatusBadRequest)
			return
		}
		if err := models.DeleteShare(pool, shareID, user.ID); err != nil {
			http.Error(w, `{"error":"failed to delete share"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
