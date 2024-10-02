package internal

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func CreateConnection() (*sql.DB, error) {
	var turso_database_url = os.Getenv("TURSO_DATABASE_URL")
	if turso_database_url == "" {
		log.Printf("Turso: env variable `%s` is not set", "TURSO_DATABASE_URL")
		return nil, errors.New("Environment variables not set")
	}

	var turso_auth_token = os.Getenv("TURSO_AUTH_TOKEN")
	if turso_auth_token == "" {
		log.Printf("Turso: env variable `%s` is not set", "TURSO_AUTH_TOKEN")
		return nil, errors.New("Environment variables not set")
	}

	var url = fmt.Sprintf("%s?authToken=%s", turso_database_url, turso_auth_token)
	db, err := sql.Open("libsql", url)
	if err != nil {
		log.Printf("Turso: Failed to open database %s: %v", turso_database_url, err)
		return nil, errors.New("Failed to connect to database")
	}

	return db, nil
}
