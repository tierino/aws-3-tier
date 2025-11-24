package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Message struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Replace with actual query
	query := "SELECT id, content FROM messages ORDER BY id DESC"

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying messages: %v", err)
		http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	messages := []Message{}
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Content); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func postMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Replace with actual insert query
	query := "INSERT INTO messages (content) VALUES ($1) RETURNING id"

	err := db.QueryRow(query, msg.Content).Scan(&msg.ID)
	if err != nil {
		log.Printf("Error inserting message: %v", err)
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

func messagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getMessages(w, r)
	case http.MethodPost:
		postMessage(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
