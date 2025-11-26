package main

import (
	"log"
	"net/http"
	"server/db"
	"server/routes"
)

func main() {
	if err := db.Init(); err != nil {
		log.Printf("Warning: Database initialization failed: %v", err)
		log.Println("Server will start but database operations may fail")
	}
	defer db.Close()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/api/messages", routes.MessagesHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
