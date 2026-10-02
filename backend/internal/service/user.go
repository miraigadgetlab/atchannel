package service

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"atchannel-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email is already registered")
	ErrNameTaken          = errors.New("name is already taken")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidRoles       = errors.New("invalid roles")
)

// AllowedRoles is the closed set of roles a channeler can hold.
// There is deliberately no moderator tier.
var AllowedRoles = []string{"user", "admin"}

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Create(ctx context.Context, name, email, password string) (*models.Channeler, error) {
	var count int64

	if err := s.db.WithContext(ctx).Model(&models.Channeler{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrEmailTaken
	}

	if err := s.db.WithContext(ctx).Model(&models.Channeler{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrNameTaken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.Channeler{
		Name:     name,
		Email:    email,
		Password: string(hashed),
		Roles:    []string{"user"},
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) Authenticate(ctx context.Context, email, password string) (*models.Channeler, error) {
	var user models.Channeler

	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}

// ValidRoles reports whether roles is non-empty and only contains allowed values.
func ValidRoles(roles []string) bool {
	if len(roles) == 0 {
		return false
	}

	seen := make(map[string]bool, len(roles))
	for _, role := range roles {
		if !slices.Contains(AllowedRoles, role) || seen[role] {
			return false
		}
		seen[role] = true
	}

	return true
}

func (s *UserService) List(ctx context.Context) ([]models.Channeler, error) {
	var users []models.Channeler

	if err := s.db.WithContext(ctx).Order("id").Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserService) SetRoles(ctx context.Context, userID uint, roles []string) (*models.Channeler, error) {
	var user models.Channeler

	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user.Roles = roles

	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// Delete removes a user together with everything they created:
// comments on their posts, their posts, and their comments elsewhere.
func (s *UserService) Delete(ctx context.Context, userID uint) error {
	var user models.Channeler

	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM comments WHERE post_id IN (SELECT id FROM posts WHERE user_id = ?)", userID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM posts WHERE user_id = ?", userID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM comments WHERE user_id = ?", userID).Error; err != nil {
			return err
		}
		return tx.Delete(&user).Error
	})
}

// EnsureAdmin creates the bootstrap admin account if it does not exist yet,
// or promotes the existing account to admin. Name collisions are resolved by
// appending a counter so a different user holding the name never blocks boot.
func (s *UserService) EnsureAdmin(ctx context.Context, name, email, password string) (*models.Channeler, error) {
	var user models.Channeler

	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	switch {
	case err == nil:
		if slices.Contains(user.Roles, "admin") {
			return &user, nil
		}
		user.Roles = append(user.Roles, "admin")
		if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	for i := 0; ; i++ {
		candidate := name
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", name, i+1)
		}

		var count int64
		if err := s.db.WithContext(ctx).Model(&models.Channeler{}).Where("name = ?", candidate).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			continue
		}

		admin := models.Channeler{
			Name:     candidate,
			Email:    email,
			Password: string(hashed),
			Roles:    []string{"admin"},
		}
		if err := s.db.WithContext(ctx).Create(&admin).Error; err != nil {
			return nil, err
		}
		return &admin, nil
	}
}
