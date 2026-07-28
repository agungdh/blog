package store

import (
	"regexp"
	"strings"

	"github.com/uptrace/bun"

	"blog/internal/model"
)

type Store struct {
	db *bun.DB
}

type FilterParams struct {
	Search   string
	Category string
	Tags     []string
	DateFrom string
	DateTo   string
}

func (f FilterParams) HasFilter() bool {
	return f.Search != "" || f.Category != "" || len(f.Tags) > 0 || f.DateFrom != "" || f.DateTo != ""
}

func New(db *bun.DB) *Store {
	db.RegisterModel((*model.Category)(nil), (*model.Tag)(nil), (*model.PostTag)(nil), (*model.Post)(nil), (*model.User)(nil), (*model.Session)(nil))
	return &Store{db: db}
}

var nonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func sanitizeFTSQuery(q string) string {
	return strings.TrimSpace(nonAlphaNum.ReplaceAllString(q, " "))
}

func toFTSQuery(input string) string {
	q := sanitizeFTSQuery(input)
	words := strings.Fields(q)
	if len(words) == 0 {
		return ""
	}
	for i, w := range words {
		words[i] = w + "*"
	}
	return strings.Join(words, " ")
}

func applyFilters(q *bun.SelectQuery, f FilterParams) *bun.SelectQuery {
	if f.Search != "" {
		sq := sanitizeFTSQuery(f.Search)
		if sq != "" {
			q = q.Where("post.id IN (SELECT rowid FROM posts_fts WHERE posts_fts MATCH ?)", sq)
		}
	}
	if f.Category != "" {
		q = q.Where("category_id IN (SELECT id FROM categories WHERE slug = ?)", f.Category)
	}
	for _, tag := range f.Tags {
		q = q.Where("post.id IN (SELECT pt.post_id FROM post_tags pt JOIN tags t ON t.id = pt.tag_id WHERE t.slug = ?)", tag)
	}
	if f.DateFrom != "" {
		q = q.Where("post.date >= ?", f.DateFrom)
	}
	if f.DateTo != "" {
		q = q.Where("post.date <= ?", f.DateTo)
	}
	return q
}
