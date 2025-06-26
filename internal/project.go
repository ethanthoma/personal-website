package internal

import "time"

type Project struct {
	Title, Url, Description string
	Lang                    []string
	Date                    time.Time
}
