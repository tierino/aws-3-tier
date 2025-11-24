package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

func initDB() error {
	// TODO: Replace with actual connection details from environment variables
	host := "localhost"
	port := 5432
	user := "postgres"
	password := "password"
	dbname := "appdb"

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		log.Printf("Warning: failed to ping database: %v", err)
		// Don't return error - allow server to start even if DB is unavailable
	}

	return nil
}

func closeDB() {
	if db != nil {
		db.Close()
	}
}
