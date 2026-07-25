package model

type Tag struct {
	ID   int64  `bun:"id,pk,autoincrement"`
	Name string `bun:"name,notnull"`
	Slug string `bun:"slug,notnull,unique"`
}
