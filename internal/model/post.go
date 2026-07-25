package model

import "time"

type Post struct {
	ID         int64
	Slug       string
	Title      string
	Date       string
	Markdown   string
	Summary    string
	CategoryID int64
	Category   *Category
	Tags       []Tag
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
