package main

import (
	"log"
	"net/http"
)

func main() {
	// Initialize database connection
	if err := initDB(); err != nil {
		log.Printf("Warning: Database initialization failed: %v", err)
		log.Println("Server will start but database operations may fail")
	}
	defer closeDB()

	// Health check endpoint (must stay as-is)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Messages API endpoints
	http.HandleFunc("/api/messages", messagesHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
