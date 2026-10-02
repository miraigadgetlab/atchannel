package models

import "time"

type User struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	Name      string      `gorm:"uniqueIndex;not null" json:"name"`
	Email     string      `gorm:"uniqueIndex;not null" json:"email" db:"email"`
	Password  string      `json:"-" db:"password"`
	Roles     StringArray `gorm:"type:text[];not null;default:'{}'" json:"roles"`
	AboutMe   string      `json:"about_me"`
	AvatarUrl string      `json:"avatar_url,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `gorm:"not null" json:"updated_at"`
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
