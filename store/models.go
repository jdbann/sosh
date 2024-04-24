package store

import "time"

type Post struct {
	Author      string
	Body        string
	PublishedAt time.Time
}
