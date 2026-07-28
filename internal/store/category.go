package store

import (
	"context"

	"blog/internal/model"
)

func (s *Store) CreateCategory(ctx context.Context, c *model.Category) error {
	_, err := s.db.NewInsert().Model(c).Exec(ctx)
	return err
}

func (s *Store) SearchCategories(ctx context.Context, query string, limit int) ([]model.Category, error) {
	if query == "" {
		return nil, nil
	}
	ftsQuery := toFTSQuery(query)
	if ftsQuery == "" {
		return nil, nil
	}
	var categories []model.Category
	err := s.db.NewSelect().
		Model(&categories).
		Where("id IN (SELECT rowid FROM categories_fts WHERE categories_fts MATCH ?)", ftsQuery).
		Order("name").
		Limit(limit).
		Scan(ctx)
	return categories, err
}

func (s *Store) GetCategoriesBySlugs(ctx context.Context, slugs []string) ([]model.Category, error) {
	if len(slugs) == 0 {
		return nil, nil
	}
	var categories []model.Category
	err := s.db.NewSelect().Model(&categories).Where("slug IN (?)", slugs).Scan(ctx)
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

func (s *Store) GetLatestCategories(ctx context.Context, limit int) ([]model.Category, error) {
	var categories []model.Category
	err := s.db.NewSelect().
		Model(&categories).
		Order("id DESC").
		Limit(limit).
		Scan(ctx)
	return categories, err
}

func (s *Store) GetCategoriesAfter(ctx context.Context, slug string, limit int) ([]model.Category, error) {
	cursor, err := s.GetCategoryBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	var categories []model.Category
	err = s.db.NewSelect().
		Model(&categories).
		Where("id < ?", cursor.ID).
		Order("id DESC").
		Limit(limit).
		Scan(ctx)
	return categories, err
}

func (s *Store) GetCategoryByID(ctx context.Context, id int64) (*model.Category, error) {
	var c model.Category
	err := s.db.NewSelect().Model(&c).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) UpdateCategory(ctx context.Context, c *model.Category) error {
	_, err := s.db.NewUpdate().Model(c).WherePK().Exec(ctx)
	return err
}

func (s *Store) DeleteCategory(ctx context.Context, id int64) error {
	_, err := s.db.NewDelete().Model((*model.Category)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
