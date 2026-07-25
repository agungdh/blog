package model

type PostTag struct {
	PostID int64 `bun:"post_id,pk"`
	TagID  int64 `bun:"tag_id,pk"`
	Post   *Post `bun:"rel:belongs-to,join:post_id=id"`
	Tag    *Tag  `bun:"rel:belongs-to,join:tag_id=id"`
}
