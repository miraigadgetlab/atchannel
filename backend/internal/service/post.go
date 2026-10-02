package service

import (
	"context"
	"errors"

	"atchannel-backend/internal/models"

	"gorm.io/gorm"
)

var ErrPostNotFound = errors.New("post not found")

type PostFilter struct {
	ChannelID *uint
	Limit     int
	Offset    int
}

type PostService struct {
	db *gorm.DB
}

func NewPostService(db *gorm.DB) *PostService {
	return &PostService{db: db}
}

func (s *PostService) Create(ctx context.Context, userID, channelID uint, title, content string) (*models.Post, error) {
	var count int64

	if err := s.db.WithContext(ctx).Model(&models.Channel{}).Where("id = ?", channelID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, ErrChannelNotFound
	}

	post := models.Post{
		ChannelID: channelID,
		Title:     title,
		Content:   content,
		UserID:    userID,
	}

	if err := s.db.WithContext(ctx).Create(&post).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Preload("User").First(&post, post.ID).Error; err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *PostService) List(ctx context.Context, filter PostFilter) ([]models.Post, error) {
	query := s.db.WithContext(ctx).Preload("User")

	if filter.ChannelID != nil {
		query = query.Where("channel_id = ?", *filter.ChannelID)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var posts []models.Post

	if err := query.Order("id desc").Find(&posts).Error; err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *PostService) GetByID(ctx context.Context, id uint) (*models.Post, error) {
	var post models.Post

	if err := s.db.WithContext(ctx).Preload("User").Preload("Comments").First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	return &post, nil
}
