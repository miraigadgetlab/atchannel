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

type PostResponse struct {
	ID        uint       `json:"id"`
	ChannelID uint       `json:"channel_id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	UserID    uint       `json:"user_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	User      PublicUser `json:"user"`
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

func toPostResponse(p models.Post) PostResponse {
	resp := PostResponse{
		ID:        p.ID,
		ChannelID: p.ChannelID,
		Title:     p.Title,
		Content:   p.Content,
		UserID:    p.UserID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		User:      toPublicUser(p.User),
	}
	if len(p.Comments) > 0 {
		resp.Comments = toCommentResponses(p.Comments)
	}
	return resp
}

func toPostResponses(posts []models.Post) []PostResponse {
	out := make([]PostResponse, len(posts))
	for i, p := range posts {
		out[i] = toPostResponse(p)
	}
	return out
}
