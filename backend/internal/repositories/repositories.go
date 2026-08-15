package repositories

import (
	"context"
	"net/netip"
	"time"

	"github.com/kosero/atchannel/backend/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, u *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetPublicByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateProfile(ctx context.Context, id, avatarURL, bio string) (*models.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, userID, familyID, tokenHash string, expiresAt time.Time) (string, error)
	GetByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, id string, replacedBy *string) error
	RevokeFamily(ctx context.Context, familyID string) error
	GetActiveByFamily(ctx context.Context, familyID string) (*models.RefreshToken, error)
	DeleteExpired(ctx context.Context) error
}

type BoardRepository interface {
	List(ctx context.Context) ([]models.Board, error)
	GetBySlug(ctx context.Context, slug string) (*models.Board, error)
}

type ThreadRepository interface {
	ListByBoard(ctx context.Context, boardID string, limit, offset int) ([]models.Thread, error)
	CountByBoard(ctx context.Context, boardID string) (int64, error)
	GetByID(ctx context.Context, id string) (*models.Thread, error)
	Create(ctx context.Context, boardID, userID, title, body string, imageURL *string) (*models.Thread, error)
	TouchBump(ctx context.Context, id string) error
	GetBoardID(ctx context.Context, id string) (string, error)
	SetPinned(ctx context.Context, id string, pinned bool) error
	SetLocked(ctx context.Context, id string, locked bool) error
	Delete(ctx context.Context, id string) error
}

type ReplyRepository interface {
	ListByThread(ctx context.Context, threadID string) ([]models.Reply, error)
	Create(ctx context.Context, threadID, userID, body string, imageURL *string, replyToID *string) (*models.Reply, error)
	GetByID(ctx context.Context, id string) (*models.Reply, error)
	Delete(ctx context.Context, id string) error
	ExistsInThread(ctx context.Context, id, threadID string) (bool, error)
}

type ReportRepository interface {
	Create(ctx context.Context, targetType, targetID, reporterID, reason string) (*models.Report, error)
	GetByID(ctx context.Context, id string) (*models.Report, error)
	UpdateStatus(ctx context.Context, id, status string) (*models.Report, error)
	List(ctx context.Context, status string, limit, offset int) ([]models.Report, error)
}

type BanRepository interface {
	Create(ctx context.Context, userID *string, ip *netip.Addr, reason string, expiresAt *time.Time) (*models.Ban, error)
	GetActiveByUser(ctx context.Context, userID string) (*models.Ban, error)
	GetActiveByIP(ctx context.Context, ip netip.Addr) (*models.Ban, error)
	ListActive(ctx context.Context) ([]models.Ban, error)
}
