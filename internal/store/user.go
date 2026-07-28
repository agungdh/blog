package store

import (
	"context"

	"blog/internal/model"
)

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	err := s.db.NewSelect().Model(&u).Where("username = ?", username).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	err := s.db.NewSelect().Model(&u).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) GetSessionByToken(ctx context.Context, token string) (*model.Session, error) {
	var sess model.Session
	err := s.db.NewSelect().Model(&sess).Relation("User").Where("token = ?", token).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *Store) CreateSession(ctx context.Context, sess *model.Session) error {
	_, err := s.db.NewInsert().Model(sess).Exec(ctx)
	return err
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.db.NewDelete().Model((*model.Session)(nil)).Where("token = ?", token).Exec(ctx)
	return err
}

func (s *Store) CountUsers(ctx context.Context) (int, error) {
	count, err := s.db.NewSelect().Model((*model.User)(nil)).Count(ctx)
	return count, err
}

func (s *Store) CreateUser(ctx context.Context, u *model.User) error {
	_, err := s.db.NewInsert().Model(u).Exec(ctx)
	return err
}
