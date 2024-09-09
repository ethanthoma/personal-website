package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var turso_database_url = os.Getenv("TURSO_DATABASE_URL")
var turso_auth_token = os.Getenv("TURSO_AUTH_TOKEN")

var db = func() *sql.DB {
	url := fmt.Sprintf("%s?authToken=%s", turso_database_url, turso_auth_token)

	db, err := sql.Open("libsql", url)
	if err != nil {
		log.Fatalf("Failed to open Turso database %s: %s", turso_database_url, err)
	}

	return db
}()

func createTableIfNotExists(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS posts (
            slug    VARCHAR(255) NOT NULL UNIQUE,
            title   VARCHAR(255) NOT NULL,
            content TEXT NOT NULL,
            date    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY (slug)
        );
	`)
	return err
}

func insertOrUpdatePost(db *sql.DB, slug, title, content string) error {
	date := time.Now()

	_, err := db.Exec(`
        INSERT INTO posts (slug, title, content, date)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(slug) DO UPDATE SET
            title = excluded.title,
            content = excluded.content,
            date = ?
	`, slug, title, content, date, date)

	if err == nil {
		log.Printf("Upserted post: %s\n", slug)
	}

	return err
}

func extractTitle(content string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# "), nil
		}
	}
	return "", errors.New("no level one header in file")
}

func processMarkdownFile(filePath string) error {
	url := fmt.Sprintf("%s?authToken=%s", turso_database_url, turso_auth_token)

	db, err := sql.Open("libsql", url)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	err = createTableIfNotExists(db)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	slug := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	title, err := extractTitle(string(content))
	if err != nil {
		return fmt.Errorf("failed to extract title from file %s: %v", filePath, err)
	}

	err = insertOrUpdatePost(db, slug, title, string(content))
	if err != nil {
		return fmt.Errorf("failed to insert/update post %s: %v", slug, err)
	}

	return nil
}

func main() {
	if len(os.Args) == 1 {
		log.Fatal("Please provide at least one markdown file path")
	}

	for _, filePath := range os.Args[1:] {
		err := processMarkdownFile(filePath)
		if err != nil {
			log.Printf("Error processing file %s: %v", filePath, err)
		}
	}
}
