package models

import "time"

type User struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	Name      string      `gorm:"uniqueIndex;not null" json:"name"`
	Password  string      `json:"-" db:"password"`
	Roles     StringArray `gorm:"type:text[];not null;default:'{}'" json:"roles"`
	AboutMe   string      `json:"about_me"`
	AvatarUrl string      `json:"avatar_url,omitempty"`
	// Note is a private message the account can leave for the admins (for
	// example, how they can be recognised if they ever lose access). It is
	// returned to the owner and to admins only, and is deliberately absent
	// from every public DTO (PublicUser, PublicProfile, ChannelerResponse).
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	// BannedAt is set when an admin suspends the account. Banned users cannot
	// log in; existing sessions are revoked at ban time.
	BannedAt *time.Time `json:"banned_at,omitempty"`
}

// RecoveryCode is a single-use, hashed account-recovery code. Only the
// SHA-256 is stored, so a database leak cannot be replayed. Codes do not
// expire: each stays valid until it is used or the owner regenerates the
// batch (see AccountService).
type RecoveryCode struct {
	ID        uint       `gorm:"primaryKey" json:"-"`
	UserID    uint       `gorm:"not null;index" json:"-"`
	TokenHash string     `gorm:"not null;uniqueIndex" json:"-"`
	UsedAt    *time.Time `json:"-"`
	CreatedAt time.Time  `json:"-"`
}

type Channel struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"uniqueIndex;not null" json:"name"`
	Title       string `gorm:"not null" json:"title"`
	Description string `json:"description"`
	// CreatedBy is the account that opened the channel: the creator and
	// admins may edit it. nil means "unknown" (rows that predate the
	// column), which only an admin can touch.
	CreatedBy *uint     `gorm:"index" json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	Posts     []Post    `json:"posts,omitempty"`
}

type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChannelID uint      `gorm:"not null;index" json:"channel_id"`
	Title     string    `gorm:"not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	Comments  []Comment `json:"comments,omitempty"`
	// LikeCount is not a column: it is filled from a correlated subquery when
	// listing. `->` keeps it read-only and `-:migration` stops AutoMigrate from
	// adding the column. A plain `-` would set Readable=false, which silently
	// drops the subquery value during scan (likes always came back as 0).
	LikeCount int64 `gorm:"->;-:migration" json:"like_count"`
	// ChannelName is derived the same way, so feeds can render "#name" without
	// resolving each channel id client-side.
	ChannelName string `gorm:"->;-:migration" json:"channel_name"`
	// PinnedAt is set by a moderator/creator to lift a post above the feed.
	PinnedAt *time.Time `json:"pinned_at,omitempty"`
	// LockedAt is set by a moderator to freeze a thread: nobody but the
	// author (and moderators) may add comments.
	LockedAt *time.Time `json:"locked_at,omitempty"`
}

// PostLike is one user's like on one post. The composite unique index makes a
// repeat like a no-op instead of a duplicate row.
type PostLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `gorm:"not null;uniqueIndex:idx_post_likes_user_post,priority:1" json:"post_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_post_likes_user_post,priority:2" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Notification is something that happened to the recipient's content. Type is
// "comment" or "like"; Actor is who caused it.
type Notification struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"not null;index" json:"user_id"`
	ActorID   uint       `gorm:"not null" json:"actor_id"`
	Type      string     `gorm:"not null" json:"type"`
	PostID    *uint      `gorm:"index" json:"post_id,omitempty"`
	CommentID *uint      `json:"comment_id,omitempty"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	Actor     User       `gorm:"foreignKey:ActorID" json:"actor"`
}

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `gorm:"not null;index" json:"post_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
}

// RefreshToken is one active session. Only the SHA-256 of the token is stored,
// so a database leak cannot be replayed as a login.
//
// FamilyID groups every token descended from one original login, which is
// what makes refresh token reuse detectable: when a revoked token is
// presented again, the whole family is burned (see SessionService).
type RefreshToken struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"not null;index" json:"user_id"`
	FamilyID   string     `gorm:"not null;index" json:"family_id"`
	TokenHash  string     `gorm:"not null;uniqueIndex" json:"-"`
	ExpiresAt  time.Time  `gorm:"not null;index" json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	RevokedFor string     `json:"-"` // why it died: "rotated", "reuse", "logout", "password"
}
