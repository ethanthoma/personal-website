package internal

import "time"

type Project struct {
	Title, Url, Description string
	Date                    time.Time
}
