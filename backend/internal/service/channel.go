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

// ChannelFilter narrows the channel listing.
type ChannelFilter struct {
	// Query, when non-empty, matches channels whose name, title or
	// description contains it (case-insensitive, wildcards literal).
	Query string
}

// UpdateChannelInput carries only the fields the client wants to change;
// a nil field is left untouched.
type UpdateChannelInput struct {
	Name        *string
	Title       *string
	Description *string
}

type ChannelService struct {
	db *gorm.DB
}

func NewChannelService(db *gorm.DB) *ChannelService {
	return &ChannelService{db: db}
}

func (s *ChannelService) Create(ctx context.Context, userID uint, name, title, description string) (*models.Channel, error) {
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
		CreatedBy:   &userID,
	}

	if err := s.db.WithContext(ctx).Create(&channel).Error; err != nil {
		return nil, err
	}

	return &channel, nil
}

func (s *ChannelService) List(ctx context.Context, filter ChannelFilter) ([]models.Channel, error) {
	query := s.db.WithContext(ctx)

	if filter.Query != "" {
		pattern := likePattern(filter.Query)
		query = query.Where("(name ILIKE ? OR title ILIKE ? OR description ILIKE ?)", pattern, pattern, pattern)
	}

	var channels []models.Channel

	if err := query.Order("id").Find(&channels).Error; err != nil {
		return nil, err
	}

	return channels, nil
}

// Update partially edits a channel. Admins may edit anything; everyone
// else only channels they opened (CreatedBy). Channels with an unknown
// owner (nil) are admin-only.
func (s *ChannelService) Update(ctx context.Context, id, userID uint, isAdmin bool, in UpdateChannelInput) (*models.Channel, error) {
	var channel models.Channel

	if err := s.db.WithContext(ctx).First(&channel, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}

	ownedByCaller := channel.CreatedBy != nil && *channel.CreatedBy == userID
	if !isAdmin && !ownedByCaller {
		return nil, ErrForbidden
	}

	if in.Name != nil && *in.Name != channel.Name {
		var count int64
		err := s.db.WithContext(ctx).Model(&models.Channel{}).
			Where("name = ? AND id <> ?", *in.Name, channel.ID).
			Count(&count).Error
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrChannelNameTaken
		}
		channel.Name = *in.Name
	}
	if in.Title != nil {
		channel.Title = *in.Title
	}
	if in.Description != nil {
		channel.Description = *in.Description
	}

	if err := s.db.WithContext(ctx).Save(&channel).Error; err != nil {
		return nil, err
	}

	return &channel, nil
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

func (s *ChannelService) CountPosts(ctx context.Context, channelID uint) (int64, error) {
	var count int64

	err := s.db.WithContext(ctx).Model(&models.Post{}).Where("channel_id = ?", channelID).Count(&count).Error
	return count, err
}

func (s *ChannelService) Delete(ctx context.Context, id uint) error {
	var channel models.Channel

	if err := s.db.WithContext(ctx).First(&channel, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChannelNotFound
		}
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM comments WHERE post_id IN (SELECT id FROM posts WHERE channel_id = ?)", id).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM posts WHERE channel_id = ?", id).Error; err != nil {
			return err
		}
		return tx.Delete(&channel).Error
	})
}
