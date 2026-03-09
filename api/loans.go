package handler

import (
	"encoding/json"
	"net/http"

	"home-library/pkg/auth"
	"home-library/pkg/db"
	"home-library/pkg/models"
)

type LoanResponse struct {
	Action     string            `json:"action"`
	Book       *models.Book      `json:"book,omitempty"`
	LoanEvent  *models.LoanEvent `json:"loan_event,omitempty"`
	PatronName *string           `json:"patron_name,omitempty"`
}

func LoansHandler(w http.ResponseWriter, r *http.Request) {
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
		// Get loan history for a book
		bookID := r.URL.Query().Get("book_id")
		patronID := r.URL.Query().Get("patron_id")
		if bookID != "" {
			events, err := models.GetLoanHistoryForBook(pool, bookID)
			if err != nil {
				http.Error(w, `{"error":"failed to get loan history"}`, http.StatusInternalServerError)
				return
			}
			if events == nil {
				events = []models.LoanEvent{}
			}
			json.NewEncoder(w).Encode(events)
		} else if patronID != "" {
			events, err := models.GetLoansByPatron(pool, patronID)
			if err != nil {
				http.Error(w, `{"error":"failed to get patron loans"}`, http.StatusInternalServerError)
				return
			}
			if events == nil {
				events = []models.LoanEvent{}
			}
			json.NewEncoder(w).Encode(events)
		} else {
			http.Error(w, `{"error":"book_id or patron_id parameter required"}`, http.StatusBadRequest)
		}

	case http.MethodPost:
		var body struct {
			ISBN     string  `json:"isbn"`
			PatronID *string `json:"patron_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if body.ISBN == "" {
			http.Error(w, `{"error":"isbn is required"}`, http.StatusBadRequest)
			return
		}

		// Look up book by ISBN for this user
		book, err := models.GetBookByISBN(pool, body.ISBN, user.ID)
		if err != nil {
			http.Error(w, `{"error":"book not found in your library"}`, http.StatusNotFound)
			return
		}

		// Check current loan status
		status, patronIDCurrent, err := models.GetCurrentLoanStatus(pool, book.ID)
		if err != nil {
			http.Error(w, `{"error":"failed to check loan status"}`, http.StatusInternalServerError)
			return
		}

		if status == "checked_out" {
			// Auto-return
			event, err := models.CreateLoanEvent(pool, book.ID, nil, "return")
			if err != nil {
				http.Error(w, `{"error":"failed to return book"}`, http.StatusInternalServerError)
				return
			}
			// Get patron name for response
			var patronName *string
			if patronIDCurrent != nil {
				patron, err := models.GetPatronByID(pool, *patronIDCurrent, user.ID)
				if err == nil {
					name := patron.FirstName + " " + patron.LastName
					patronName = &name
				}
			}
			json.NewEncoder(w).Encode(LoanResponse{
				Action:     "return",
				Book:       book,
				LoanEvent:  event,
				PatronName: patronName,
			})
		} else if body.PatronID != nil {
			// Checkout to patron
			event, err := models.CreateLoanEvent(pool, book.ID, body.PatronID, "checkout")
			if err != nil {
				http.Error(w, `{"error":"failed to checkout book"}`, http.StatusInternalServerError)
				return
			}
			patron, _ := models.GetPatronByID(pool, *body.PatronID, user.ID)
			var patronName *string
			if patron != nil {
				name := patron.FirstName + " " + patron.LastName
				patronName = &name
			}
			json.NewEncoder(w).Encode(LoanResponse{
				Action:     "checkout",
				Book:       book,
				LoanEvent:  event,
				PatronName: patronName,
			})
		} else {
			// Needs patron selection
			json.NewEncoder(w).Encode(LoanResponse{
				Action: "needs_patron",
				Book:   book,
			})
		}

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
