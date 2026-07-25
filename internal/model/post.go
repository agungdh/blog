package model

import "time"

type Post struct {
	ID         int64     `bun:"id,pk,autoincrement"`
	Slug       string    `bun:"slug,notnull,unique"`
	Title      string    `bun:"title,notnull"`
	Date       string    `bun:"date,notnull"`
	Markdown   string    `bun:"markdown,notnull"`
	CategoryID int64     `bun:"category_id"`
	Category   *Category `bun:"rel:belongs-to,join:category_id=id"`
	Tags       []Tag     `bun:"m2m:post_tags,join:Post=Tag"`
	Summary    string    `bun:"-"`
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt  time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}
