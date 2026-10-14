package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"atchannel-backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrPostNotFound = errors.New("post not found")
var ErrPostLocked = errors.New("post is locked")

// Sort orders accepted by List. Anything else falls back to newest.
const (
	SortNewest = "new"
	SortOldest = "old"
	SortTop    = "top"
)

type PostFilter struct {
	ChannelID *uint
	// AuthorID, when set, limits the listing to one account's posts.
	AuthorID *uint
	// Query, when non-empty, matches posts whose title or content contains
	// it (case-insensitive, wildcards treated literally).
	Query  string
	Sort   string
	Limit  int
	Offset int
}

// UpdatePostInput carries only the fields the client wants to change;
// a nil field is left untouched.
type UpdatePostInput struct {
	Title   *string
	Content *string
}

type PostService struct {
	db *gorm.DB
}

// postColumns is the shared SELECT for post queries: every post column plus the
// two derived values that are not columns — the like tally and the channel name
// — so a feed costs one round trip instead of one query per post.
const postColumns = "posts.*, " +
	"(SELECT COUNT(*) FROM post_likes pl WHERE pl.post_id = posts.id) AS like_count, " +
	"(SELECT name FROM channels c WHERE c.id = posts.channel_id) AS channel_name"

func NewPostService(db *gorm.DB) *PostService {
	return &PostService{db: db}
}

// Pin lifts (or lowers) a post above the feed.
func (s *PostService) Pin(ctx context.Context, id uint, pinned bool) error {
	var post models.Post
	if err := s.db.WithContext(ctx).First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		return err
	}

	var value any
	if pinned {
		value = time.Now()
	} else {
		value = nil
	}
	return s.db.WithContext(ctx).Model(&post).Update("pinned_at", value).Error
}

// Lock freezes (or unfreezes) a thread for commenting.
func (s *PostService) Lock(ctx context.Context, id uint, locked bool) error {
	var post models.Post
	if err := s.db.WithContext(ctx).First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		return err
	}

	var value any
	if locked {
		value = time.Now()
	} else {
		value = nil
	}
	return s.db.WithContext(ctx).Model(&post).Update("locked_at", value).Error
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

	if err := s.db.WithContext(ctx).Preload("User").Select(postColumns).First(&post, post.ID).Error; err != nil {
		return nil, err
	}

	return &post, nil
}

// Count returns how many posts match the filter, independent of paging.
func (s *PostService) Count(ctx context.Context, filter PostFilter) (int64, error) {
	query := s.db.WithContext(ctx).Model(&models.Post{})

	if filter.ChannelID != nil {
		query = query.Where("channel_id = ?", *filter.ChannelID)
	}
	if filter.AuthorID != nil {
		query = query.Where("user_id = ?", *filter.AuthorID)
	}
	if filter.Query != "" {
		pattern := likePattern(filter.Query)
		query = query.Where("(title ILIKE ? OR content ILIKE ?)", pattern, pattern)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

func (s *PostService) List(ctx context.Context, filter PostFilter) ([]models.Post, error) {
	// like_count rides along as a correlated subquery so feed ordering by
	// popularity and showing the tally cost one round trip, not one per post.
	query := s.db.WithContext(ctx).
		Preload("User").
		Select(postColumns)

	if filter.ChannelID != nil {
		query = query.Where("channel_id = ?", *filter.ChannelID)
	}
	if filter.AuthorID != nil {
		query = query.Where("user_id = ?", *filter.AuthorID)
	}

	if filter.Query != "" {
		pattern := likePattern(filter.Query)
		query = query.Where("(title ILIKE ? OR content ILIKE ?)", pattern, pattern)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Pinned posts always float above the feed order; everything else follows
	// the requested sort. gorm appends each Order call, so the primary sort
	// decides among non-pinned posts.
	query = query.Order("posts.pinned_at IS NULL")

	switch filter.Sort {
	case SortOldest:
		query = query.Order("posts.id asc")
	case SortTop:
		query = query.Order("like_count desc, posts.id desc")
	default:
		query = query.Order("posts.id desc")
	}

	var posts []models.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}

	return posts, nil
}

// LikedSet reports which of the given posts the viewer has liked, in one
// query, so a feed can mark every item without an N+1.
func (s *PostService) LikedSet(ctx context.Context, viewerID uint, postIDs []uint) (map[uint]bool, error) {
	liked := make(map[uint]bool)
	if viewerID == 0 || len(postIDs) == 0 {
		return liked, nil
	}

	var rows []uint
	if err := s.db.WithContext(ctx).Model(&models.PostLike{}).
		Where("user_id = ? AND post_id IN ?", viewerID, postIDs).
		Pluck("post_id", &rows).Error; err != nil {
		return nil, err
	}
	for _, id := range rows {
		liked[id] = true
	}

	return liked, nil
}

// Like records a like, if it is not already there. It reports whether a new
// row was created (so callers only notify on the first like, not every tap)
// along with the post author's id for that notification.
func (s *PostService) Like(ctx context.Context, postID, userID uint) (bool, uint, error) {
	var post models.Post
	if err := s.db.WithContext(ctx).Select("id", "user_id").First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, 0, ErrPostNotFound
		}
		return false, 0, err
	}

	res := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.PostLike{PostID: postID, UserID: userID})
	if res.Error != nil {
		return false, 0, res.Error
	}

	return res.RowsAffected > 0, post.UserID, nil
}

// Unlike removes a like. It is idempotent.
func (s *PostService) Unlike(ctx context.Context, postID, userID uint) error {
	return s.db.WithContext(ctx).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Delete(&models.PostLike{}).Error
}

func (s *PostService) Update(ctx context.Context, id, userID uint, canModerate bool, in UpdatePostInput) (*models.Post, error) {
	var post models.Post

	if err := s.db.WithContext(ctx).First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	if post.UserID != userID && !canModerate {
		return nil, ErrForbidden
	}

	if in.Title != nil {
		post.Title = strings.TrimSpace(*in.Title)
	}
	if in.Content != nil {
		post.Content = strings.TrimSpace(*in.Content)
	}

	if err := s.db.WithContext(ctx).Save(&post).Error; err != nil {
		return nil, err
	}

	err := s.db.WithContext(ctx).
		Preload("User").
		Select(postColumns).
		First(&post, post.ID).Error
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *PostService) Delete(ctx context.Context, id, userID uint, canModerate bool) error {
	var post models.Post

	if err := s.db.WithContext(ctx).First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		return err
	}

	if post.UserID != userID && !canModerate {
		return ErrForbidden
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", id).Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("post_id = ?", id).Delete(&models.PostLike{}).Error; err != nil {
			return err
		}
		if err := tx.Where("post_id = ?", id).Delete(&models.Notification{}).Error; err != nil {
			return err
		}
		return tx.Delete(&post).Error
	})
}

func (s *PostService) GetByID(ctx context.Context, id uint) (*models.Post, error) {
	var post models.Post

	err := s.db.WithContext(ctx).
		Preload("User").
		Select(postColumns).
		First(&post, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	return &post, nil
}
