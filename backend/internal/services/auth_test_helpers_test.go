package services_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/kosero/atchannel/backend/internal/models"
	"github.com/kosero/atchannel/backend/internal/repositories"
)

// InMemoryUserRepo is a minimal test double for the user repository.
type InMemoryUserRepo struct {
	mu    sync.Mutex
	users []*models.User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{}
}

func (r *InMemoryUserRepo) nextID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (r *InMemoryUserRepo) Create(ctx context.Context, u *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *u
	if cp.ID == "" {
		cp.ID = r.nextID()
	}
	if cp.Role == "" {
		cp.Role = models.RoleUser
	}
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = time.Now()
	}
	r.users = append(r.users, &cp)
	*u = cp
	return nil
}

func (r *InMemoryUserRepo) find(fn func(*models.User) bool) *models.User {
	for _, u := range r.users {
		if fn(u) {
			return u
		}
	}
	return nil
}

func (r *InMemoryUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	u := r.find(func(x *models.User) bool { return x.ID == id })
	if u == nil {
		return nil, repositories.ErrNotFound
	}
	return u, nil
}

func (r *InMemoryUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	u := r.find(func(x *models.User) bool { return x.Username == username })
	if u == nil {
		return nil, repositories.ErrNotFound
	}
	return u, nil
}

func (r *InMemoryUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u := r.find(func(x *models.User) bool { return x.Email == email })
	if u == nil {
		return nil, repositories.ErrNotFound
	}
	return u, nil
}

func (r *InMemoryUserRepo) GetPublicByUsername(ctx context.Context, username string) (*models.User, error) {
	return r.GetByUsername(ctx, username)
}

func (r *InMemoryUserRepo) UpdateProfile(ctx context.Context, id, avatarURL, bio string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u := r.find(func(x *models.User) bool { return x.ID == id })
	if u == nil {
		return nil, repositories.ErrNotFound
	}
	u.AvatarURL = avatarURL
	u.Bio = bio
	return u, nil
}

func (r *InMemoryUserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return r.find(func(x *models.User) bool { return x.Username == username }) != nil, nil
}

func (r *InMemoryUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.find(func(x *models.User) bool { return x.Email == email }) != nil, nil
}
