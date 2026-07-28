package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"blog/internal/model"
	"blog/internal/store"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenNotFound      = errors.New("token not found")
)

const sessionTTL = 30 * 24 * time.Hour // 30 days

type AuthService struct {
	store *store.Store
}

func NewAuthService(s *store.Store) *AuthService {
	return &AuthService{store: s}
}

func (a *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (a *AuthService) CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (a *AuthService) GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (a *AuthService) Login(ctx context.Context, username, password string) (string, string, error) {
	user, err := a.store.GetUserByUsername(ctx, username)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	if !a.CheckPassword(user.Password, password) {
		return "", "", ErrInvalidCredentials
	}

	token, err := a.GenerateToken()
	if err != nil {
		return "", "", err
	}

	expiresAt := time.Now().Add(sessionTTL).Format(time.RFC3339)
	sess := &model.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	if err := a.store.CreateSession(ctx, sess); err != nil {
		return "", "", err
	}

	return token, user.Nama, nil
}

func (a *AuthService) ValidateToken(ctx context.Context, token string) (*model.User, error) {
	sess, err := a.store.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, ErrTokenNotFound
	}

	expiresAt, err := time.Parse(time.RFC3339, sess.ExpiresAt)
	if err != nil {
		return nil, ErrTokenExpired
	}
	if time.Now().After(expiresAt) {
		return nil, ErrTokenExpired
	}

	return sess.User, nil
}

func (a *AuthService) Logout(ctx context.Context, token string) error {
	return a.store.DeleteSession(ctx, token)
}
