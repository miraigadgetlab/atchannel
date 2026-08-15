package postgres

import (
	"context"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kosero/atchannel/backend/internal/db"
	"github.com/kosero/atchannel/backend/internal/models"
)

type BanRepo struct {
	q *db.Queries
}

func NewBanRepo(pool *pgxpool.Pool) *BanRepo {
	return &BanRepo{q: db.New(pool)}
}

func (r *BanRepo) Create(ctx context.Context, userID *string, ip *netip.Addr, reason string, expiresAt *time.Time) (*models.Ban, error) {
	row, err := r.q.CreateBan(ctx, db.CreateBanParams{
		UserID:    pgUUID(userID),
		Ip:        pgInet(ip),
		Reason:    reason,
		ExpiresAt: pgTime(expiresAt),
	})
	if err != nil {
		return nil, err
	}
	return banFromRow(row), nil
}

func (r *BanRepo) GetActiveByUser(ctx context.Context, userID string) (*models.Ban, error) {
	row, err := r.q.GetActiveBanForUser(ctx, pgUUID(ptrStr(userID)))
	if err != nil {
		return nil, mapErr(err)
	}
	return banFromRow(row), nil
}

func (r *BanRepo) GetActiveByIP(ctx context.Context, ip netip.Addr) (*models.Ban, error) {
	row, err := r.q.GetActiveBanForIP(ctx, pgInet(&ip))
	if err != nil {
		return nil, mapErr(err)
	}
	return banFromRow(row), nil
}

func (r *BanRepo) ListActive(ctx context.Context) ([]models.Ban, error) {
	rows, err := r.q.ListActiveBans(ctx)
	if err != nil {
		return nil, err
	}
	bans := make([]models.Ban, 0, len(rows))
	for _, row := range rows {
		bans = append(bans, *banFromRow(row))
	}
	return bans, nil
}

func banFromRow(row db.Ban) *models.Ban {
	b := &models.Ban{
		ID:        row.ID,
		Reason:    row.Reason,
		CreatedAt: row.CreatedAt,
	}
	if row.UserID.Valid {
		u := uuid.UUID(row.UserID.Bytes)
		s := u.String()
		b.UserID = &s
	}
	if row.Ip != nil {
		a := *row.Ip
		b.IP = &a
	}
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		b.ExpiresAt = &t
	}
	return b
}
