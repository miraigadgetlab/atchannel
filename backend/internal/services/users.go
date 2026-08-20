package services

import (
	"context"
	"errors"

	"github.com/kosero/atchannel/backend/internal/middleware"
	"github.com/kosero/atchannel/backend/internal/models"
	"github.com/kosero/atchannel/backend/internal/repositories"
	"github.com/kosero/atchannel/backend/internal/repositories/postgres"
)

var ErrUserNotFound = errors.New("user not found")

type UserService struct {
	users repositories.UserRepository
}

func NewUserService(users repositories.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.users.GetPublicByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// CurrentUser resolves the full user record for the ID carried in the
// request context (injected by the auth middleware).
func (s *UserService) CurrentUser(ctx context.Context) (*models.User, error) {
	id := middleware.ContextUserID(ctx)
	if id == "" {
		return nil, ErrUserNotFound
	}
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, id, avatarURL, bio, color string) (*models.User, error) {
	if len(bio) > 500 {
		return nil, ErrBodyTooLong
	}
	return s.users.UpdateProfile(ctx, id, avatarURL, bio, color)
}
