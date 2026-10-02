package models

import "time"

type Channeler struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	Name      string      `gorm:"uniqueIndex;not null" json:"name"`
	Email     string      `gorm:"uniqueIndex;not null" json:"email" db:"email"`
	Password  string      `json:"-" db:"password"`
	Roles     StringArray `gorm:"type:text[];not null;default:'{}'" json:"roles"`
	AboutMe   string      `json:"about_me"`
	AvatarUrl string      `json:"avatar_url,omitempty"`
}

type Channel struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Posts       []Post    `json:"posts,omitempty"`
}

type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChannelID uint      `gorm:"not null" json:"channel_id"`
	Title     string    `gorm:"not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      Channeler `gorm:"foreignKey:UserID" json:"user"`
	Comments  []Comment `json:"comments,omitempty"`
}

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `gorm:"not null" json:"post_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      Channeler `gorm:"foreignKey:UserID" json:"user"`
	CreatedAt time.Time `json:"created_at"`
}
