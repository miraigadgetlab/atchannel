package service

import (
	"context"

	"atchannel-backend/internal/models"

	"gorm.io/gorm"
)

type Stats struct {
	Users    int64 `json:"users"`
	Channels int64 `json:"channels"`
	Posts    int64 `json:"posts"`
	Comments int64 `json:"comments"`
}

type StatsService struct {
	db *gorm.DB
}

func NewStatsService(db *gorm.DB) *StatsService {
	return &StatsService{db: db}
}

func (s *StatsService) Get(ctx context.Context) (*Stats, error) {
	var stats Stats

	counts := []struct {
		model any
		out   *int64
	}{
		{&models.User{}, &stats.Users},
		{&models.Channel{}, &stats.Channels},
		{&models.Post{}, &stats.Posts},
		{&models.Comment{}, &stats.Comments},
	}

	for _, c := range counts {
		if err := s.db.WithContext(ctx).Model(c.model).Count(c.out).Error; err != nil {
			return nil, err
		}
	}

	return &stats, nil
}
