package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kosero/atchannel/backend/internal/db"
	"github.com/kosero/atchannel/backend/internal/models"
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
