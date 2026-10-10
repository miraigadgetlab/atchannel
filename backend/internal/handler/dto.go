package handler

import (
	"time"

	"atchannel-backend/internal/models"
)

// PublicUser is the author view embedded in posts and comments. It
// deliberately omits email, roles and the password hash: those are private,
// and serializing the raw models.User used to leak every author's email to
// anyone who could list posts or comments.
type PublicUser struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	AboutMe   string `json:"about_me"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type CommentResponse struct {
	ID        uint       `json:"id"`
	PostID    uint       `json:"post_id"`
	Content   string     `json:"content"`
	UserID    uint       `json:"user_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	User      PublicUser `json:"user"`
}

// PublicProfile is the account view shown on someone else's profile page: the
// safe author fields plus activity counts, and nothing private.
type PublicProfile struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	AboutMe      string    `json:"about_me"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	PostCount    int64     `json:"post_count"`
	CommentCount int64     `json:"comment_count"`
}

type PostResponse struct {
	ID        uint       `json:"id"`
	ChannelID uint       `json:"channel_id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	UserID    uint       `json:"user_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	User      PublicUser `json:"user"`
	Likes     int64      `json:"likes"`
	LikedByMe bool       `json:"liked_by_me"`
	Pinned    bool       `json:"pinned"`
	Locked    bool       `json:"locked"`
	// Comments is only populated by endpoints that actually load them; it is
	// omitted otherwise. The dedicated /posts/:id/comments endpoint is the
	// paginated source of truth.
	Comments []CommentResponse `json:"comments,omitempty"`
}

func toPublicUser(u models.User) PublicUser {
	return PublicUser{
		ID:        u.ID,
		Name:      u.Name,
		AboutMe:   u.AboutMe,
		AvatarURL: u.AvatarUrl,
	}
}

func toCommentResponse(c models.Comment) CommentResponse {
	return CommentResponse{
		ID:        c.ID,
		PostID:    c.PostID,
		Content:   c.Content,
		UserID:    c.UserID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		User:      toPublicUser(c.User),
	}
}

func toCommentResponses(comments []models.Comment) []CommentResponse {
	out := make([]CommentResponse, len(comments))
	for i, c := range comments {
		out[i] = toCommentResponse(c)
	}
	return out
}

// NotificationResponse is a single activity-feed entry.
type NotificationResponse struct {
	ID        uint       `json:"id"`
	Type      string     `json:"type"`
	PostID    *uint      `json:"post_id,omitempty"`
	CommentID *uint      `json:"comment_id,omitempty"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	Actor     PublicUser `json:"actor"`
}

func toNotificationResponse(n models.Notification) NotificationResponse {
	return NotificationResponse{
		ID:        n.ID,
		Type:      n.Type,
		PostID:    n.PostID,
		CommentID: n.CommentID,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
		Actor:     toPublicUser(n.Actor),
	}
}

func toNotificationResponses(items []models.Notification) []NotificationResponse {
	out := make([]NotificationResponse, len(items))
	for i, n := range items {
		out[i] = toNotificationResponse(n)
	}
	return out
}

// toPostResponse builds the public view. likedByMe is whether the requesting
// account (when there is one) has liked the post.
func toPostResponse(p models.Post, likedByMe bool) PostResponse {
	resp := PostResponse{
		ID:        p.ID,
		ChannelID: p.ChannelID,
		Title:     p.Title,
		Content:   p.Content,
		UserID:    p.UserID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		User:      toPublicUser(p.User),
		Likes:     p.LikeCount,
		LikedByMe: likedByMe,
		Pinned:    p.PinnedAt != nil,
		Locked:    p.LockedAt != nil,
	}
	if len(p.Comments) > 0 {
		resp.Comments = toCommentResponses(p.Comments)
	}
	return resp
}

func toPostResponses(posts []models.Post, likedByMe map[uint]bool) []PostResponse {
	out := make([]PostResponse, len(posts))
	for i, p := range posts {
		out[i] = toPostResponse(p, likedByMe[p.ID])
	}
	return out
}
