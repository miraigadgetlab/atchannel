package service

import (
	"context"
	"errors"

	"atchannel-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email is already registered")
	ErrNameTaken          = errors.New("name is already taken")
)

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
