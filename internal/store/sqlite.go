package store

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"

	"blog/internal/model"
)

type Store struct {
	db *sql.DB
}

func New(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			date TEXT NOT NULL,
			markdown TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) GetAllPosts() ([]model.Post, error) {
	rows, err := s.db.Query(`
		SELECT id, slug, title, date, markdown, created_at, updated_at
		FROM posts
		ORDER BY date DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var p model.Post
		var createdAt, updatedAt string
		if err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Date, &p.Markdown, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (s *Store) GetPostBySlug(slug string) (*model.Post, error) {
	var p model.Post
	var createdAt, updatedAt string
	err := s.db.QueryRow(`
		SELECT id, slug, title, date, markdown, created_at, updated_at
		FROM posts
		WHERE slug = ?
	`, slug).Scan(&p.ID, &p.Slug, &p.Title, &p.Date, &p.Markdown, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return &p, nil
}

func (s *Store) CreatePost(p *model.Post) error {
	_, err := s.db.Exec(`
		INSERT INTO posts (slug, title, date, markdown)
		VALUES (?, ?, ?, ?)
	`, p.Slug, p.Title, p.Date, p.Markdown)
	return err
}

func (s *Store) UpdatePost(p *model.Post) error {
	_, err := s.db.Exec(`
		UPDATE posts SET title = ?, date = ?, markdown = ?, updated_at = datetime('now')
		WHERE slug = ?
	`, p.Title, p.Date, p.Markdown, p.Slug)
	return err
}

func (s *Store) DeletePost(slug string) error {
	_, err := s.db.Exec(`DELETE FROM posts WHERE slug = ?`, slug)
	return err
}
