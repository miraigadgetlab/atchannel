package service

import (
	"context"
	"errors"

	"atchannel-backend/internal/models"

	"gorm.io/gorm"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrForbidden       = errors.New("you do not own this resource")
)

type CommentService struct {
	db *gorm.DB
}

func NewCommentService(db *gorm.DB) *CommentService {
	return &CommentService{db: db}
}

func (s *CommentService) Create(ctx context.Context, userID, postID uint, content string) (*models.Comment, error) {
	var count int64

	if err := s.db.WithContext(ctx).Model(&models.Post{}).Where("id = ?", postID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, ErrPostNotFound
	}

	comment := models.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: content,
	}

	if err := s.db.WithContext(ctx).Create(&comment).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Preload("User").First(&comment, comment.ID).Error; err != nil {
		return nil, err
	}

	return &comment, nil
}

func (s *CommentService) ListByPost(ctx context.Context, postID uint, limit, offset int) ([]models.Comment, int64, error) {
	var count int64

	if err := s.db.WithContext(ctx).Model(&models.Post{}).Where("id = ?", postID).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if count == 0 {
		return nil, 0, ErrPostNotFound
	}

	query := s.db.WithContext(ctx).
		Preload("User").
		Where("post_id = ?", postID)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var comments []models.Comment

	if err := query.Order("id").Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	var total int64
	if err := s.db.WithContext(ctx).Model(&models.Comment{}).
		Where("post_id = ?", postID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (s *CommentService) Delete(ctx context.Context, commentID, userID uint, isAdmin bool) error {
	var comment models.Comment

	if err := s.db.WithContext(ctx).First(&comment, commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommentNotFound
		}
		return err
	}

	if comment.UserID != userID && !isAdmin {
		return ErrForbidden
	}

	return s.db.WithContext(ctx).Delete(&comment).Error
}
