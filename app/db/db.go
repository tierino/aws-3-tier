package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	_ "github.com/lib/pq"
)

var db *sql.DB

type DBSecret struct {
	DBClusterIdentifier string `json:"dbClusterIdentifier"`
	Password            string `json:"password"`
	DBName              string `json:"dbname"`
	Port                int    `json:"port"`
	Engine              string `json:"engine"`
	Host                string `json:"host"`
	Username            string `json:"username"`
}

func Init() error {
	secretName := os.Getenv("DB_SECRET_NAME")
	if secretName == "" {
		return fmt.Errorf("DB_SECRET_NAME not set in environment")
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	})
	if err != nil {
		return fmt.Errorf("failed to retrieve secret: %w", err)
	}

	var secret DBSecret
	if err := json.Unmarshal([]byte(*result.SecretString), &secret); err != nil {
		return fmt.Errorf("failed to parse secret: %w", err)
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
		secret.Host, secret.Port, secret.Username, secret.Password, secret.DBName)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		log.Printf("Warning: failed to ping database: %v", err)
	}

	return nil
}

func Close() {
	if db != nil {
		db.Close()
	}
}

type Message struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	UserID  int    `json:"userId"`
}

func GetMessages() ([]Message, error) {
	query := "SELECT id, content, userId FROM messages ORDER BY id DESC"

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	messages := []Message{}
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Content, &msg.UserID); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func InsertMessage(content string) (*Message, error) {
	query := "INSERT INTO messages (content) VALUES ($1) RETURNING id"

	var msg Message
	msg.Content = content

	err := db.QueryRow(query, content).Scan(&msg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert message: %w", err)
	}

	return &msg, nil
}
