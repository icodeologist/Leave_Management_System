package store

import (
	"context"
	"errors"
	"strings"

	"leave-management/internal/model"

	"gorm.io/gorm"
)

var (
	ErrNotFound    = errors.New("user not found")
	ErrEmailExists = errors.New("email already exists")
)

type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) UserByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := s.db.WithContext(ctx).
		Where("email = ?", strings.ToLower(strings.TrimSpace(email))).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user, ErrNotFound
	}
	return user, err
}

func (s *Store) CreateUser(ctx context.Context, user *model.User) error {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	err := s.db.WithContext(ctx).Create(user).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrEmailExists
	}
	return err
}
