package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kosero/atchannel/backend/internal/db"
	"github.com/kosero/atchannel/backend/internal/models"
	"github.com/kosero/atchannel/backend/internal/repositories"
)

type BoardRepo struct {
	q *db.Queries
}

func NewBoardRepo(pool *pgxpool.Pool) *BoardRepo {
	return &BoardRepo{q: db.New(pool)}
}

func (r *BoardRepo) List(ctx context.Context) ([]models.Board, error) {
	rows, err := r.q.ListBoards(ctx)
	if err != nil {
		return nil, err
	}
	boards := make([]models.Board, 0, len(rows))
	for _, row := range rows {
		boards = append(boards, models.Board{
			ID:          row.ID,
			Slug:        row.Slug,
			Name:        row.Name,
			Description: row.Description,
			CreatedAt:   row.CreatedAt,
		})
	}
	return boards, nil
}

func (r *BoardRepo) GetBySlug(ctx context.Context, slug string) (*models.Board, error) {
	row, err := r.q.GetBoardBySlug(ctx, slug)
	if err != nil {
		return nil, mapErr(err)
	}
	return &models.Board{
		ID:          row.ID,
		Slug:        row.Slug,
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   row.CreatedAt,
	}, nil
}

func (r *BoardRepo) Create(ctx context.Context, slug, name, description string) (*models.Board, error) {
	row, err := r.q.CreateBoard(ctx, db.CreateBoardParams{
		Slug:        slug,
		Name:        name,
		Description: description,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, repositories.ErrBoardExists
		}
		return nil, err
	}
	return &models.Board{
		ID:          row.ID,
		Slug:        row.Slug,
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   row.CreatedAt,
	}, nil
}
