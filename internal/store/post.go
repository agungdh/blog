package store

import (
	"context"
	"fmt"

	"blog/internal/model"
)

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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.NewInsert().Model(p).Exec(ctx)
	if err != nil {
		return err
	}
	for _, t := range p.Tags {
		pt := &model.PostTag{PostID: p.ID, TagID: t.ID}
		if _, err := tx.NewInsert().Model(pt).Ignore().Exec(ctx); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) UpdatePost(ctx context.Context, p *model.Post) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var existing model.Post
	err = tx.NewSelect().
		Model(&existing).
		Where("slug = ?", p.Slug).
		Scan(ctx)
	if err != nil {
		return err
	}

	p.ID = existing.ID
	p.CreatedAt = existing.CreatedAt

	_, err = tx.NewUpdate().Model(p).WherePK().Exec(ctx)
	if err != nil {
		return err
	}

	if _, err = tx.NewDelete().Model((*model.PostTag)(nil)).Where("post_id = ?", p.ID).Exec(ctx); err != nil {
		return err
	}
	for _, t := range p.Tags {
		pt := &model.PostTag{PostID: p.ID, TagID: t.ID}
		if _, err := tx.NewInsert().Model(pt).Ignore().Exec(ctx); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) DeletePost(ctx context.Context, slug string) error {
	var post model.Post
	_, err := s.db.NewDelete().Model(&post).Where("slug = ?", slug).Exec(ctx)
	return err
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

func (s *Store) GetPostByID(ctx context.Context, id int64) (*model.Post, error) {
	var post model.Post
	err := s.db.NewSelect().
		Model(&post).
		Relation("Category").
		Relation("Tags").
		Where("post.id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *Store) UpdatePostByID(ctx context.Context, p *model.Post) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.NewUpdate().Model(p).WherePK().Exec(ctx)
	if err != nil {
		return err
	}

	if _, err = tx.NewDelete().Model((*model.PostTag)(nil)).Where("post_id = ?", p.ID).Exec(ctx); err != nil {
		return err
	}
	for _, t := range p.Tags {
		pt := &model.PostTag{PostID: p.ID, TagID: t.ID}
		if _, err := tx.NewInsert().Model(pt).Ignore().Exec(ctx); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) DeletePostByID(ctx context.Context, id int64) error {
	_, err := s.db.NewDelete().Model((*model.Post)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
