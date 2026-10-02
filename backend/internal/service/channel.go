package service

import (
	"context"
	"errors"

	"atchannel-backend/internal/models"

	"gorm.io/gorm"
)

var (
	ErrChannelNotFound  = errors.New("channel not found")
	ErrChannelNameTaken = errors.New("channel name is already taken")
)

type ChannelService struct {
	db *gorm.DB
}

func NewChannelService(db *gorm.DB) *ChannelService {
	return &ChannelService{db: db}
}

func (s *ChannelService) Create(ctx context.Context, name, title, description string) (*models.Channel, error) {
	var count int64

	if err := s.db.WithContext(ctx).Model(&models.Channel{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrChannelNameTaken
	}

	channel := models.Channel{
		Name:        name,
		Title:       title,
		Description: description,
	}

	if err := s.db.WithContext(ctx).Create(&channel).Error; err != nil {
		return nil, err
	}

	return &channel, nil
}

func (s *ChannelService) List(ctx context.Context) ([]models.Channel, error) {
	var channels []models.Channel

	if err := s.db.WithContext(ctx).Order("id").Find(&channels).Error; err != nil {
		return nil, err
	}

	return channels, nil
}

func (s *ChannelService) GetByID(ctx context.Context, id uint) (*models.Channel, error) {
	var channel models.Channel

	if err := s.db.WithContext(ctx).First(&channel, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}

	return &channel, nil
}
