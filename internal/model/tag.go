package model

type Tag struct {
	ID   int64  `bun:"id,pk,autoincrement" json:"-"`
	Name string `bun:"name,notnull" json:"name"`
	Slug string `bun:"slug,notnull,unique" json:"slug"`
}
