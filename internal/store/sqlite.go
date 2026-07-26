package store

import (
	"context"
	"fmt"

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
	db.RegisterModel((*model.Category)(nil), (*model.Tag)(nil), (*model.PostTag)(nil), (*model.Post)(nil))
	return &Store{db: db}
}

func (s *Store) GetLatestPosts(ctx context.Context, limit int) ([]model.Post, error) {
	var posts []model.Post
	err := s.db.NewSelect().
		Model(&posts).
		Relation("Category").
		Relation("Tags").
		Order("post.date DESC", "post.id DESC").
		Limit(limit).
		Scan(ctx)
	return posts, err
}

func (s *Store) GetPostsAfter(ctx context.Context, slug string, limit int) ([]model.Post, error) {
	cursor, err := s.GetPostBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	var posts []model.Post
	err = s.db.NewSelect().
		Model(&posts).
		Relation("Category").
		Relation("Tags").
		Where("(post.date < ? OR (post.date = ? AND post.id < ?))", cursor.Date, cursor.Date, cursor.ID).
		Order("post.date DESC", "post.id DESC").
		Limit(limit).
		Scan(ctx)
	return posts, err
}

func (s *Store) CountPosts(ctx context.Context) (int, error) {
	count, err := s.db.NewSelect().Model((*model.Post)(nil)).Count(ctx)
	return count, err
}

func (s *Store) GetAllPostTitles(ctx context.Context) ([]string, error) {
	var titles []string
	err := s.db.NewSelect().
		Model((*model.Post)(nil)).
		Column("title").
		Scan(ctx, &titles)
	return titles, err
}

func (s *Store) GetPostBySlug(ctx context.Context, slug string) (*model.Post, error) {
	var post model.Post
	err := s.db.NewSelect().
		Model(&post).
		Relation("Category").
		Relation("Tags").
		Where("post.slug = ?", slug).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *Store) CreatePost(ctx context.Context, p *model.Post) error {
	_, err := s.db.NewInsert().Model(p).Exec(ctx)
	if err != nil {
		return err
	}
	for _, t := range p.Tags {
		_, err := s.db.NewRaw("INSERT OR IGNORE INTO post_tags (post_id, tag_id) VALUES (?, ?)", p.ID, t.ID).Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) UpdatePost(ctx context.Context, p *model.Post) error {
	var existing model.Post
	err := s.db.NewSelect().
		Model(&existing).
		Where("slug = ?", p.Slug).
		Scan(ctx)
	if err != nil {
		return err
	}

	p.ID = existing.ID
	p.CreatedAt = existing.CreatedAt

	_, err = s.db.NewUpdate().Model(p).WherePK().Exec(ctx)
	if err != nil {
		return err
	}

	_, err = s.db.NewRaw("DELETE FROM post_tags WHERE post_id = ?", p.ID).Exec(ctx)
	if err != nil {
		return err
	}
	for _, t := range p.Tags {
		_, err = s.db.NewRaw("INSERT OR IGNORE INTO post_tags (post_id, tag_id) VALUES (?, ?)", p.ID, t.ID).Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) DeletePost(ctx context.Context, slug string) error {
	var post model.Post
	_, err := s.db.NewDelete().Model(&post).Where("slug = ?", slug).Exec(ctx)
	return err
}

func (s *Store) CreateCategory(ctx context.Context, c *model.Category) error {
	_, err := s.db.NewInsert().Model(c).Exec(ctx)
	return err
}

func (s *Store) GetAllCategories(ctx context.Context) ([]model.Category, error) {
	var categories []model.Category
	err := s.db.NewSelect().Model(&categories).Order("name").Scan(ctx)
	return categories, err
}

func (s *Store) GetCategoryBySlug(ctx context.Context, slug string) (*model.Category, error) {
	var c model.Category
	err := s.db.NewSelect().Model(&c).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) CreateTag(ctx context.Context, t *model.Tag) error {
	_, err := s.db.NewInsert().Model(t).Exec(ctx)
	return err
}

func (s *Store) GetAllTags(ctx context.Context) ([]model.Tag, error) {
	var tags []model.Tag
	err := s.db.NewSelect().Model(&tags).Order("name").Scan(ctx)
	return tags, err
}

func (s *Store) GetTagBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	var t model.Tag
	err := s.db.NewSelect().Model(&t).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) GetLatestPostsByCategorySlug(ctx context.Context, slug string, limit int) ([]model.Post, error) {
	var posts []model.Post
	err := s.db.NewSelect().
		Model(&posts).
		Relation("Category").
		Relation("Tags").
		Where("category_id IN (SELECT id FROM categories WHERE slug = ?)", slug).
		Order("post.date DESC", "post.id DESC").
		Limit(limit).
		Scan(ctx)
	return posts, err
}

func (s *Store) GetPostsByCategorySlugAfter(ctx context.Context, catSlug, postSlug string, limit int) ([]model.Post, error) {
	cursor, err := s.GetPostBySlug(ctx, postSlug)
	if err != nil {
		return nil, err
	}
	var posts []model.Post
	err = s.db.NewSelect().
		Model(&posts).
		Relation("Category").
		Relation("Tags").
		Where("category_id IN (SELECT id FROM categories WHERE slug = ?)", catSlug).
		Where("(post.date < ? OR (post.date = ? AND post.id < ?))", cursor.Date, cursor.Date, cursor.ID).
		Order("post.date DESC", "post.id DESC").
		Limit(limit).
		Scan(ctx)
	return posts, err
}

func (s *Store) GetLatestPostsByTagSlug(ctx context.Context, slug string, limit int) ([]model.Post, error) {
	var posts []model.Post
	err := s.db.NewSelect().
		Model(&posts).
		Relation("Category").
		Relation("Tags").
		Where("post.id IN (SELECT pt.post_id FROM post_tags pt JOIN tags t ON t.id = pt.tag_id WHERE t.slug = ?)", slug).
		Order("post.date DESC", "post.id DESC").
		Limit(limit).
		Scan(ctx)
	return posts, err
}

func (s *Store) GetFilteredPosts(ctx context.Context, filter FilterParams, limit int, afterSlug string) ([]model.Post, error) {
	var posts []model.Post
	q := s.db.NewSelect().
		Model(&posts).
		Relation("Category").
		Relation("Tags")

	q = applyFilters(q, filter)

	if afterSlug != "" {
		cursor, err := s.GetPostBySlug(ctx, afterSlug)
		if err != nil {
			return nil, fmt.Errorf("cursor post: %w", err)
		}
		q = q.Where("(post.date < ? OR (post.date = ? AND post.id < ?))", cursor.Date, cursor.Date, cursor.ID)
	}

	err := q.Order("post.date DESC", "post.id DESC").Limit(limit).Scan(ctx)
	return posts, err
}

func applyFilters(q *bun.SelectQuery, f FilterParams) *bun.SelectQuery {
	if f.Search != "" {
		q = q.Where("post.id IN (SELECT rowid FROM posts_fts WHERE posts_fts MATCH ?)", f.Search)
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

func (s *Store) GetPostsByTagSlugAfter(ctx context.Context, tagSlug, postSlug string, limit int) ([]model.Post, error) {
	cursor, err := s.GetPostBySlug(ctx, postSlug)
	if err != nil {
		return nil, err
	}
	var posts []model.Post
	err = s.db.NewSelect().
		Model(&posts).
		Relation("Category").
		Relation("Tags").
		Where("post.id IN (SELECT pt.post_id FROM post_tags pt JOIN tags t ON t.id = pt.tag_id WHERE t.slug = ?)", tagSlug).
		Where("(post.date < ? OR (post.date = ? AND post.id < ?))", cursor.Date, cursor.Date, cursor.ID).
		Order("post.date DESC", "post.id DESC").
		Limit(limit).
		Scan(ctx)
	return posts, err
}
