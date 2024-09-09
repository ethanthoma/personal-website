package main

import (
	"database/sql"
	"fmt"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"log"
	"os"
	"time"
)

var turso_database_url = os.Getenv("TURSO_DATABASE_URL")
var turso_auth_token = os.Getenv("TURSO_AUTH_TOKEN")

var DB = func() *sql.DB {
	url := fmt.Sprintf("%s?authToken=%s", turso_database_url, turso_auth_token)

	db, err := sql.Open("libsql", url)
	if err != nil {
		log.Fatalf("Failed to open Turso database %s: %s", turso_database_url, err)
	}

	return db
}()

type Post struct {
	Slug    string
	Title   string
	Content string
	Date    time.Time
}

func GetPosts() ([]Post, error) {
	rows, err := DB.Query(`
        SELECT
            *
        FROM
            posts
        ORDER BY
            date
        DESC;
    `)
	if err != nil {
		log.Printf("Turso Database: failed to get post table: %s", err)
		return nil, err
	}
	defer rows.Close()

	return rowsToPosts(rows)
}

func GetPost(slug string) (Post, error) {
	rows, err := DB.Query(fmt.Sprint(`
        SELECT 
            *
        FROM 
            posts
        WHERE
            slug = '`, slug, `'
        ;
    `))
	if err != nil {
		return Post{}, err
	}
	defer rows.Close()

	posts, err := rowsToPosts(rows)
	if err != nil {
		return Post{}, err
	}

	if len(posts) == 0 {
		return Post{}, err
	}

	return posts[0], nil
}

func rowsToPosts(rows *sql.Rows) ([]Post, error) {
	var posts []Post

	for rows.Next() {
		var post Post

		if err := rows.Scan(
			&post.Slug,
			&post.Title,
			&post.Content,
			&post.Date,
		); err != nil {
			log.Printf("Turso Database: error scanning row: %s", err)
			return nil, err
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Turso Database: error during rows iteration: %s", err)
		return nil, err
	}

	return posts, nil
}
