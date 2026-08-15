package models

import (
	"net/netip"
	"time"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleMod   Role = "mod"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	AvatarURL    string    `json:"avatarUrl"`
	Bio          string    `json:"bio"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
}

type UserPublic struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatarUrl"`
	Bio       string    `json:"bio"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type Board struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Thread struct {
	ID        string    `json:"id"`
	BoardID   string    `json:"-"`
	UserID    string    `json:"userId"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	ImageURL  string    `json:"imageUrl"`
	IsPinned  bool      `json:"isPinned"`
	IsLocked  bool      `json:"isLocked"`
	BumpedAt  time.Time `json:"bumpedAt"`
	CreatedAt time.Time `json:"createdAt"`

	BoardSlug     string     `json:"boardSlug"`
	AuthorName    string     `json:"authorName"`
	AuthorRole    Role       `json:"authorRole"`
	ReplyCount    int64      `json:"replyCount"`
	LastReplyAt   *time.Time `json:"lastReplyAt,omitempty"`
	Bumped        bool       `json:"bumped"`
	BumpLimit     bool       `json:"bumpLimit"`
	ImageArchived bool       `json:"imageArchived"`
	Resto         string     `json:"resto"`
	Sticky        bool       `json:"sticky"`
	Closed        bool       `json:"closed"`
}

type Reply struct {
	ID         string     `json:"id"`
	ThreadID   string     `json:"threadId"`
	UserID     string     `json:"userId"`
	Body       string     `json:"body"`
	ImageURL   string     `json:"imageUrl"`
	ReplyToID  *string    `json:"replyToId,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`

	AuthorName   string    `json:"authorName"`
	AuthorRole   Role      `json:"authorRole"`
	Deleted      bool      `json:"deleted"`
}

type Report struct {
	ID         string    `json:"id"`
	TargetType string    `json:"targetType"`
	TargetID   string    `json:"targetId"`
	ReporterID string    `json:"reporterId"`
	Reason     string    `json:"reason"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`

	TargetBoardSlug string `json:"targetBoardSlug,omitempty"`
	TargetThreadID  string `json:"targetThreadId,omitempty"`
	TargetBody      string `json:"targetBody,omitempty"`
	ReporterName    string `json:"reporterName,omitempty"`
}

type Ban struct {
	ID        string     `json:"id"`
	UserID    *string    `json:"userId,omitempty"`
	IP        *netip.Addr `json:"ip,omitempty"`
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
}

type RefreshToken struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userID"`
	FamilyID   string    `json:"familyID"`
	TokenHash  string    `json:"-"`
	Revoked    bool      `json:"revoked"`
	ReplacedBy *string   `json:"replacedBy,omitempty"`
	ExpiresAt  time.Time `json:"expiresAt"`
	CreatedAt  time.Time `json:"createdAt"`
}
