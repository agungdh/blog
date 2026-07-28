package store

import (
	"context"

	"blog/internal/model"
)

func (s *Store) CreateTag(ctx context.Context, t *model.Tag) error {
	_, err := s.db.NewInsert().Model(t).Exec(ctx)
	return err
}

func (s *Store) SearchTags(ctx context.Context, query string, limit int) ([]model.Tag, error) {
	if query == "" {
		return nil, nil
	}
	ftsQuery := toFTSQuery(query)
	if ftsQuery == "" {
		return nil, nil
	}
	var tags []model.Tag
	err := s.db.NewSelect().
		Model(&tags).
		Where("id IN (SELECT rowid FROM tags_fts WHERE tags_fts MATCH ?)", ftsQuery).
		Order("name").
		Limit(limit).
		Scan(ctx)
	return tags, err
}

func (s *Store) GetTagsBySlugs(ctx context.Context, slugs []string) ([]model.Tag, error) {
	if len(slugs) == 0 {
		return nil, nil
	}
	var tags []model.Tag
	err := s.db.NewSelect().Model(&tags).Where("slug IN (?)", slugs).Scan(ctx)
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

func (s *Store) GetLatestTags(ctx context.Context, limit int) ([]model.Tag, error) {
	var tags []model.Tag
	err := s.db.NewSelect().
		Model(&tags).
		Order("id DESC").
		Limit(limit).
		Scan(ctx)
	return tags, err
}

func (s *Store) GetTagsAfter(ctx context.Context, slug string, limit int) ([]model.Tag, error) {
	cursor, err := s.GetTagBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	var tags []model.Tag
	err = s.db.NewSelect().
		Model(&tags).
		Where("id < ?", cursor.ID).
		Order("id DESC").
		Limit(limit).
		Scan(ctx)
	return tags, err
}

func (s *Store) GetTagByID(ctx context.Context, id int64) (*model.Tag, error) {
	var t model.Tag
	err := s.db.NewSelect().Model(&t).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) UpdateTag(ctx context.Context, t *model.Tag) error {
	_, err := s.db.NewUpdate().Model(t).WherePK().Exec(ctx)
	return err
}

func (s *Store) DeleteTag(ctx context.Context, id int64) error {
	_, err := s.db.NewDelete().Model((*model.Tag)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
