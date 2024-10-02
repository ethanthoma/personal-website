package internal

import (
	"database/sql"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func CreateTableIfNotExists(db *sql.DB) error {
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
