package model

import "time"

type Session struct {
	ID        int64     `bun:"id,pk,autoincrement"`
	UserID    int64     `bun:"user_id,notnull"`
	Token     string    `bun:"token,notnull,unique"`
	ExpiresAt string    `bun:"expires_at,notnull"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	User      *User     `bun:"rel:belongs-to,join:user_id=id"`
}
