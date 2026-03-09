package main

import (
	"fmt"
	"net/http"
	"os"

	handler "home-library/api"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	godotenv.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handler.HealthHandler)
	mux.HandleFunc("/api/user", handler.UserHandler)
	mux.HandleFunc("/api/book-lookup", handler.BookLookupHandler)
	mux.HandleFunc("/api/books", handler.BooksHandler)
	mux.HandleFunc("/api/patrons", handler.PatronsHandler)
	mux.HandleFunc("/api/loans", handler.LoansHandler)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5179"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8089"
	}

	fmt.Printf("API server listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, c.Handler(mux)); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
