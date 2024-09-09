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
		log.Fatalf("Turso Database: Failed to open database %s: %s", turso_database_url, err)
	}

	log.Printf("Turso Database: Connected to %s", turso_database_url)
	return db
}()

type Post struct {
	Slug    string
	Title   string
	Content string
	Date    time.Time
}

func GetPosts() ([]Post, error) {
	log.Println("Turso Database: getting posts...")

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
	log.Println("Turso Database: fetched posts")

	return rowsToPosts(rows)
}

func GetPost(slug string) (Post, error) {
	log.Printf("Turso Database: fetching post %s...", slug)

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

	log.Printf("Turso Database: fetched post %s", slug)
	return posts[0], nil
}

func rowsToPosts(rows *sql.Rows) ([]Post, error) {
	log.Println("Turso Database: parsing SQL rows to posts...")

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

	log.Println("Turso Database: parsed SQL rows to posts")

	return posts, nil
}
