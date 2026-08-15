package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kosero/atchannel/backend/internal/db"
	"github.com/kosero/atchannel/backend/internal/models"
)

type RefreshTokenRepo struct {
	q *db.Queries
}

func NewRefreshTokenRepo(pool *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{q: db.New(pool)}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, userID, familyID, tokenHash string, expiresAt time.Time) (string, error) {
	row, err := r.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		FamilyID:  familyID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}
	return row.ID, nil
}

func (r *RefreshTokenRepo) GetByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, mapErr(err)
	}
	return refreshTokenFromRow(row), nil
}

func (r *RefreshTokenRepo) Revoke(ctx context.Context, id string, replacedBy *string) error {
	return r.q.RevokeRefreshToken(ctx, db.RevokeRefreshTokenParams{
		ID:         id,
		ReplacedBy: pgUUID(replacedBy),
	})
}

func (r *RefreshTokenRepo) RevokeFamily(ctx context.Context, familyID string) error {
	return r.q.RevokeFamily(ctx, familyID)
}

func (r *RefreshTokenRepo) GetActiveByFamily(ctx context.Context, familyID string) (*models.RefreshToken, error) {
	row, err := r.q.GetActiveTokenByFamily(ctx, familyID)
	if err != nil {
		return nil, mapErr(err)
	}
	return refreshTokenFromRow(row), nil
}

func (r *RefreshTokenRepo) DeleteExpired(ctx context.Context) error {
	return r.q.DeleteExpiredRefreshTokens(ctx)
}

func refreshTokenFromRow(row db.RefreshToken) *models.RefreshToken {
	t := &models.RefreshToken{
		ID:        row.ID,
		UserID:    row.UserID,
		FamilyID:  row.FamilyID,
		TokenHash: row.TokenHash,
		Revoked:   row.Revoked,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}
	if row.ReplacedBy.Valid {
		u := uuid.UUID(row.ReplacedBy.Bytes)
		s := u.String()
		t.ReplacedBy = &s
	}
	return t
}
