package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kosero/atchannel/backend/internal/db"
	"github.com/kosero/atchannel/backend/internal/models"
)

type UserRepo struct {
	q *db.Queries
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{q: db.New(pool)}
}

func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	row, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		AvatarUrl:    pgText(u.AvatarURL),
		Bio:          u.Bio,
		Role:         string(u.Role),
	})
	if err != nil {
		return err
	}
	created := userFromRow(row)
	u.ID = created.ID
	u.CreatedAt = created.CreatedAt
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return userFromRow(row), nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, mapErr(err)
	}
	return userFromRow(row), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, mapErr(err)
	}
	return userFromRow(row), nil
}

func (r *UserRepo) GetPublicByUsername(ctx context.Context, username string) (*models.User, error) {
	row, err := r.q.GetUserPublicByUsername(ctx, username)
	if err != nil {
		return nil, mapErr(err)
	}
	return &models.User{
		ID:        row.ID,
		Username:  row.Username,
		AvatarURL: row.AvatarUrl.String,
		Bio:       row.Bio,
		Role:      models.Role(row.Role),
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *UserRepo) UpdateProfile(ctx context.Context, id, avatarURL, bio string) (*models.User, error) {
	row, err := r.q.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID:        id,
		AvatarUrl: pgText(avatarURL),
		Bio:       bio,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return userFromRow(row), nil
}

func (r *UserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return r.q.UserExistsByUsername(ctx, username)
}

func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.q.UserExistsByEmail(ctx, email)
}

func userFromRow(row db.User) *models.User {
	return &models.User{
		ID:           row.ID,
		Username:     row.Username,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		AvatarURL:    row.AvatarUrl.String,
		Bio:          row.Bio,
		Role:         models.Role(row.Role),
		CreatedAt:    row.CreatedAt,
	}
}
