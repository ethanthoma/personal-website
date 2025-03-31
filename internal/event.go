package internal

import "time"

type Event interface {
	GetDate() time.Time
}

func (event *Post) GetDate() time.Time    { return event.Date }
func (event *Project) GetDate() time.Time { return event.Date }
