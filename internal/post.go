package internal

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Post struct {
	Slug    string
	Title   string
	Content string
	Date    time.Time
	HTML    template.HTML
}

func InsertOrUpdatePost(db *sql.DB, slug, title, content string) error {
	date := time.Now()

	if _, err := db.Exec(`
        INSERT INTO posts (slug, title, content, date)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(slug) DO UPDATE SET
            title = excluded.title,
            content = excluded.content
	`, slug, title, content, date); err != nil {
		log.Printf("Turso: failed to upserted post %s\n", slug)

		return err
	}

	return nil
}

func GetPosts(db *sql.DB) ([]Post, error) {
	log.Println("Turso: getting posts...")

	rows, err := db.Query(`
        SELECT
            *
        FROM
            posts
        ORDER BY
            date
        DESC;
    `)
	if err != nil {
		log.Printf("Turso: failed to get post table: %v", err)
		return nil, err
	}
	defer rows.Close()

	log.Println("Turso: fetched posts")
	return rowsToPosts(rows)
}

func GetPost(db *sql.DB, slug string) (Post, error) {
	log.Printf("Turso: fetching post %s...", slug)

	rows, err := db.Query(fmt.Sprint(`
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

	log.Printf("Turso: fetched post %s", slug)
	return posts[0], nil
}

func rowsToPosts(rows *sql.Rows) ([]Post, error) {
	log.Println("Turso: parsing SQL rows to posts...")

	var posts []Post

	for rows.Next() {
		var post Post

		if err := rows.Scan(
			&post.Slug,
			&post.Title,
			&post.Content,
			&post.Date,
		); err != nil {
			log.Printf("Turso: error scanning row: %v", err)
			return nil, err
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Turso: error during rows iteration: %v", err)
		return nil, err
	}

	log.Println("Turso: parsed SQL rows to posts")

	return posts, nil
}
