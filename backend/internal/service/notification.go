package service

import (
	"context"
	"time"

	"atchannel-backend/internal/models"

	"gorm.io/gorm"
)

// Notification kinds.
const (
	NotificationComment = "comment"
	NotificationLike    = "like"
)

// NotificationService records and reads the lightweight activity feed: someone
// commented on or liked your post.
type NotificationService struct {
	db *gorm.DB
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// Notify records an event for recipientID. Self-actions are dropped: nobody
// needs a notification about their own comment or like.
func (s *NotificationService) Notify(ctx context.Context, recipientID, actorID uint, kind string, postID, commentID *uint) error {
	if recipientID == 0 || recipientID == actorID {
		return nil
	}

	return s.db.WithContext(ctx).Create(&models.Notification{
		UserID:    recipientID,
		ActorID:   actorID,
		Type:      kind,
		PostID:    postID,
		CommentID: commentID,
	}).Error
}

// List returns the recipient's notifications, newest first, actor preloaded.
func (s *NotificationService) List(ctx context.Context, userID uint, limit, offset int) ([]models.Notification, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var items []models.Notification
	if err := query.Preload("Actor").Order("id desc").Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// UnreadCount backs the badge in the UI.
func (s *NotificationService) UnreadCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

// MarkAllRead clears every unread notification for the user.
func (s *NotificationService) MarkAllRead(ctx context.Context, userID uint) (int64, error) {
	res := s.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", time.Now())
	return res.RowsAffected, res.Error
}
