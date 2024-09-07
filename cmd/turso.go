package main

import (
	"database/sql"
	"fmt"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"log"
	"os"
	"strings"
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

type Blog struct {
	Slug        string
	Title       string
	Description string
	Content     string
	Date        time.Time
	Tags        []string
}

func GetBlogTable() ([]Blog, error) {
	rows, err := DB.Query(`
        SELECT
            *
        FROM
            blogs
        ORDER BY
            date
        DESC;
    `)
	if err != nil {
		log.Printf("Turso Database: failed to get blog table: %s", err)
		return nil, err
	}
	defer rows.Close()

	return rowsToBlogs(rows)
}

func GetBlogBySlug(slug string) (Blog, error) {
	rows, err := DB.Query(fmt.Sprint(`
        SELECT 
            *
        FROM 
            blogs
        WHERE
            slug = '`, slug, `'
        LIMIT 
            1
        ;
    `))
	if err != nil {
		return Blog{}, err
	}
	defer rows.Close()

	blogs, err := rowsToBlogs(rows)
	if err != nil {
		return Blog{}, err
	}

	if len(blogs) == 0 {
		return Blog{}, err
	}

	return blogs[0], nil
}

func rowsToBlogs(rows *sql.Rows) ([]Blog, error) {
	var blogs []Blog

	for rows.Next() {
		var blog Blog

		var tags string

		if err := rows.Scan(
			&blog.Slug,
			&blog.Title,
			&blog.Description,
			&blog.Content,
			&blog.Date,
			&tags,
		); err != nil {
			log.Printf("Turso Database: error scanning row: %s", err)
			return nil, err
		}

		blog.Tags = strings.Split(tags, ",")

		blogs = append(blogs, blog)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Turso Database: error during rows iteration: %s", err)
		return nil, err
	}

	return blogs, nil
}
