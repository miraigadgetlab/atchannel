package services

import (
	"context"
	"errors"
	"net/netip"
	"time"

	"github.com/kosero/atchannel/backend/internal/models"
	"github.com/kosero/atchannel/backend/internal/repositories"
	"github.com/kosero/atchannel/backend/internal/repositories/postgres"
)

var (
	ErrInvalidReportTarget = errors.New("invalid report target")
	ErrReportNotFound      = errors.New("report not found")
	ErrBanNotFound         = errors.New("ban not found")
)

type ModerationService struct {
	reports  repositories.ReportRepository
	bans     repositories.BanRepository
	threads  repositories.ThreadRepository
	replies  repositories.ReplyRepository
	users    repositories.UserRepository
}

func NewModerationService(
	reports repositories.ReportRepository,
	bans repositories.BanRepository,
	threads repositories.ThreadRepository,
	replies repositories.ReplyRepository,
	users repositories.UserRepository,
) *ModerationService {
	return &ModerationService{
		reports: reports,
		bans:    bans,
		threads: threads,
		replies: replies,
		users:   users,
	}
}

func (s *ModerationService) CreateReport(ctx context.Context, reporterID, targetType, targetID, reason string) (*models.Report, error) {
	if targetType != "thread" && targetType != "reply" {
		return nil, ErrInvalidReportTarget
	}
	if len(reason) == 0 || len(reason) > 500 {
		return nil, ErrInvalidReportTarget
	}
	switch targetType {
	case "thread":
		if _, err := s.threads.GetByID(ctx, targetID); err != nil {
			if errors.Is(err, postgres.ErrNotFound) {
				return nil, ErrThreadNotFound
			}
			return nil, err
		}
	case "reply":
		if _, err := s.replies.GetByID(ctx, targetID); err != nil {
			if errors.Is(err, postgres.ErrNotFound) {
				return nil, ErrReplyNotFound
			}
			return nil, err
		}
	}
	return s.reports.Create(ctx, targetType, targetID, reporterID, reason)
}

func (s *ModerationService) ListReports(ctx context.Context, status string, page, perPage int) ([]models.Report, error) {
	if status == "" {
		status = "open"
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.reports.List(ctx, status, perPage, (page-1)*perPage)
}

func (s *ModerationService) ResolveReport(ctx context.Context, id, status string) (*models.Report, error) {
	if status != "resolved" && status != "dismissed" {
		return nil, ErrInvalidReportTarget
	}
	return s.reports.UpdateStatus(ctx, id, status)
}

func (s *ModerationService) BanUser(ctx context.Context, userID, reason string, expiresAt *time.Time) (*models.Ban, error) {
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, ErrBanNotFound
		}
		return nil, err
	}
	return s.bans.Create(ctx, &userID, nil, reason, expiresAt)
}

func (s *ModerationService) BanIP(ctx context.Context, ip netip.Addr, reason string, expiresAt *time.Time) (*models.Ban, error) {
	return s.bans.Create(ctx, nil, &ip, reason, expiresAt)
}

func (s *ModerationService) DeleteThread(ctx context.Context, id string) error {
	return s.threads.Delete(ctx, id)
}

func (s *ModerationService) DeleteReply(ctx context.Context, id string) error {
	return s.replies.Delete(ctx, id)
}

func (s *ModerationService) ListBans(ctx context.Context) ([]models.Ban, error) {
	return s.bans.ListActive(ctx)
}
