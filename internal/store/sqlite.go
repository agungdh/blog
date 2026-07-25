package store

import (
	"database/sql"
	"fmt"
	"strings"
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

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-2000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return err
		}
	}

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			slug          TEXT NOT NULL UNIQUE,
			title         TEXT NOT NULL,
			date          TEXT NOT NULL,
			markdown      TEXT NOT NULL,
			created_at    TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at    TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`ALTER TABLE posts ADD COLUMN category_id INTEGER REFERENCES categories(id)`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS post_tags (
			post_id INTEGER NOT NULL REFERENCES posts(id),
			tag_id  INTEGER NOT NULL REFERENCES tags(id),
			PRIMARY KEY (post_id, tag_id)
		)
	`)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) GetAllPosts() ([]model.Post, error) {
	rows, err := s.db.Query(`
		SELECT p.id, p.slug, p.title, p.date, p.markdown, p.created_at, p.updated_at,
		       c.id, c.name, c.slug
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		ORDER BY p.date DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	var postIDs []int64
	for rows.Next() {
		var p model.Post
		var createdAt, updatedAt string
		var catID sql.NullInt64
		var catName, catSlug sql.NullString
		if err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Date, &p.Markdown, &createdAt, &updatedAt,
			&catID, &catName, &catSlug); err != nil {
			return nil, err
		}
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		if catID.Valid {
			p.CategoryID = catID.Int64
			p.Category = &model.Category{
				ID:   catID.Int64,
				Name: catName.String,
				Slug: catSlug.String,
			}
		}
		posts = append(posts, p)
		postIDs = append(postIDs, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	tagsByPost, err := s.getTagsForPostIDs(postIDs)
	if err != nil {
		return nil, err
	}
	for i := range posts {
		posts[i].Tags = tagsByPost[posts[i].ID]
	}

	return posts, nil
}

func (s *Store) GetPostBySlug(slug string) (*model.Post, error) {
	var p model.Post
	var createdAt, updatedAt string
	var catID sql.NullInt64
	var catName, catSlug sql.NullString
	err := s.db.QueryRow(`
		SELECT p.id, p.slug, p.title, p.date, p.markdown, p.created_at, p.updated_at,
		       c.id, c.name, c.slug
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.slug = ?
	`, slug).Scan(&p.ID, &p.Slug, &p.Title, &p.Date, &p.Markdown, &createdAt, &updatedAt,
		&catID, &catName, &catSlug)
	if err != nil {
		return nil, err
	}
	p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	if catID.Valid {
		p.CategoryID = catID.Int64
		p.Category = &model.Category{
			ID:   catID.Int64,
			Name: catName.String,
			Slug: catSlug.String,
		}
	}

	tags, err := s.getTagsForPostID(p.ID)
	if err != nil {
		return nil, err
	}
	p.Tags = tags

	return &p, nil
}

func (s *Store) CreatePost(p *model.Post) error {
	res, err := s.db.Exec(`
		INSERT INTO posts (slug, title, date, markdown, category_id)
		VALUES (?, ?, ?, ?, NULLIF(?, 0))
	`, p.Slug, p.Title, p.Date, p.Markdown, p.CategoryID)
	if err != nil {
		return err
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = postID

	return s.setPostTags(postID, p.Tags)
}

func (s *Store) UpdatePost(p *model.Post) error {
	var postID int64
	err := s.db.QueryRow(`SELECT id FROM posts WHERE slug = ?`, p.Slug).Scan(&postID)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		UPDATE posts SET title = ?, date = ?, markdown = ?, category_id = NULLIF(?, 0), updated_at = datetime('now')
		WHERE slug = ?
	`, p.Title, p.Date, p.Markdown, p.CategoryID, p.Slug)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`DELETE FROM post_tags WHERE post_id = ?`, postID)
	if err != nil {
		return err
	}

	return s.setPostTags(postID, p.Tags)
}

func (s *Store) DeletePost(slug string) error {
	var postID int64
	err := s.db.QueryRow(`SELECT id FROM posts WHERE slug = ?`, slug).Scan(&postID)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`DELETE FROM post_tags WHERE post_id = ?`, postID)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`DELETE FROM posts WHERE slug = ?`, slug)
	return err
}

func (s *Store) CreateCategory(c *model.Category) error {
	res, err := s.db.Exec(`INSERT INTO categories (name, slug) VALUES (?, ?)`, c.Name, c.Slug)
	if err != nil {
		return err
	}
	c.ID, _ = res.LastInsertId()
	return nil
}

func (s *Store) GetAllCategories() ([]model.Category, error) {
	rows, err := s.db.Query(`SELECT id, name, slug FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (s *Store) GetCategoryBySlug(slug string) (*model.Category, error) {
	var c model.Category
	err := s.db.QueryRow(`SELECT id, name, slug FROM categories WHERE slug = ?`, slug).Scan(&c.ID, &c.Name, &c.Slug)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) CreateTag(t *model.Tag) error {
	res, err := s.db.Exec(`INSERT INTO tags (name, slug) VALUES (?, ?)`, t.Name, t.Slug)
	if err != nil {
		return err
	}
	t.ID, _ = res.LastInsertId()
	return nil
}

func (s *Store) GetAllTags() ([]model.Tag, error) {
	rows, err := s.db.Query(`SELECT id, name, slug FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (s *Store) GetTagBySlug(slug string) (*model.Tag, error) {
	var t model.Tag
	err := s.db.QueryRow(`SELECT id, name, slug FROM tags WHERE slug = ?`, slug).Scan(&t.ID, &t.Name, &t.Slug)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) GetPostsByCategorySlug(slug string) ([]model.Post, error) {
	rows, err := s.db.Query(`
		SELECT p.id, p.slug, p.title, p.date, p.markdown, p.created_at, p.updated_at,
		       c.id, c.name, c.slug
		FROM posts p
		JOIN categories c ON c.id = p.category_id
		WHERE c.slug = ?
		ORDER BY p.date DESC
	`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPostsWithTags(rows, s)
}

func (s *Store) GetPostsByTagSlug(slug string) ([]model.Post, error) {
	rows, err := s.db.Query(`
		SELECT p.id, p.slug, p.title, p.date, p.markdown, p.created_at, p.updated_at,
		       c.id, c.name, c.slug
		FROM posts p
		JOIN post_tags pt ON pt.post_id = p.id
		JOIN tags t ON t.id = pt.tag_id
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE t.slug = ?
		ORDER BY p.date DESC
	`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPostsWithTags(rows, s)
}

func (s *Store) setPostTags(postID int64, tags []model.Tag) error {
	for _, t := range tags {
		_, err := s.db.Exec(`INSERT OR IGNORE INTO post_tags (post_id, tag_id) VALUES (?, ?)`, postID, t.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) getTagsForPostID(postID int64) ([]model.Tag, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.name, t.slug
		FROM post_tags pt
		JOIN tags t ON t.id = pt.tag_id
		WHERE pt.post_id = ?
		ORDER BY t.name
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (s *Store) getTagsForPostIDs(ids []int64) (map[int64][]model.Tag, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	rows, err := s.db.Query(fmt.Sprintf(`
		SELECT pt.post_id, t.id, t.name, t.slug
		FROM post_tags pt
		JOIN tags t ON t.id = pt.tag_id
		WHERE pt.post_id IN (%s)
		ORDER BY t.name
	`, strings.Join(placeholders, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tagsByPost := make(map[int64][]model.Tag)
	for rows.Next() {
		var postID int64
		var t model.Tag
		if err := rows.Scan(&postID, &t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		tagsByPost[postID] = append(tagsByPost[postID], t)
	}
	return tagsByPost, rows.Err()
}

func scanPostsWithTags(rows *sql.Rows, s *Store) ([]model.Post, error) {
	var posts []model.Post
	var postIDs []int64
	for rows.Next() {
		var p model.Post
		var createdAt, updatedAt string
		var catID sql.NullInt64
		var catName, catSlug sql.NullString
		if err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Date, &p.Markdown, &createdAt, &updatedAt,
			&catID, &catName, &catSlug); err != nil {
			return nil, err
		}
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		p.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		if catID.Valid {
			p.CategoryID = catID.Int64
			p.Category = &model.Category{
				ID:   catID.Int64,
				Name: catName.String,
				Slug: catSlug.String,
			}
		}
		posts = append(posts, p)
		postIDs = append(postIDs, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	tagsByPost, err := s.getTagsForPostIDs(postIDs)
	if err != nil {
		return nil, err
	}
	for i := range posts {
		posts[i].Tags = tagsByPost[posts[i].ID]
	}

	return posts, nil
}
