package model

import "time"

type User struct {
	ID        int64     `bun:"id,pk,autoincrement"`
	Username  string    `bun:"username,notnull,unique"`
	Password  string    `bun:"password,notnull"`
	Nama      string    `bun:"nama,notnull"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}
