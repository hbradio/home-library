package handler

import (
	"encoding/json"
	"net/http"

	"home-library/pkg/auth"
	"home-library/pkg/db"
	"home-library/pkg/models"
)

func PatronsHandler(w http.ResponseWriter, r *http.Request) {
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
		// Single patron detail
		id := r.URL.Query().Get("id")
		if id != "" {
			patron, err := models.GetPatronByID(pool, id, user.ID)
			if err != nil {
				http.Error(w, `{"error":"patron not found"}`, http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(patron)
			return
		}

		// List with optional search
		q := r.URL.Query().Get("q")
		patrons, err := models.GetPatrons(pool, user.ID, q)
		if err != nil {
			http.Error(w, `{"error":"failed to get patrons"}`, http.StatusInternalServerError)
			return
		}
		if patrons == nil {
			patrons = []models.Patron{}
		}
		json.NewEncoder(w).Encode(patrons)

	case http.MethodPost:
		var body struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if body.FirstName == "" || body.LastName == "" {
			http.Error(w, `{"error":"first_name and last_name are required"}`, http.StatusBadRequest)
			return
		}
		patron, err := models.CreatePatron(pool, user.ID, body.FirstName, body.LastName)
		if err != nil {
			http.Error(w, `{"error":"failed to create patron"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(patron)

	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, `{"error":"id parameter required"}`, http.StatusBadRequest)
			return
		}
		if err := models.DeletePatron(pool, id, user.ID); err != nil {
			http.Error(w, `{"error":"failed to delete patron"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
