package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kosero/atchannel/backend/internal/db"
	"github.com/kosero/atchannel/backend/internal/models"
)

type ThreadRepo struct {
	q *db.Queries
}

func NewThreadRepo(pool *pgxpool.Pool) *ThreadRepo {
	return &ThreadRepo{q: db.New(pool)}
}

func (r *ThreadRepo) ListByBoard(ctx context.Context, boardID string, limit, offset int) ([]models.Thread, error) {
	rows, err := r.q.GetBoardThreads(ctx, db.GetBoardThreadsParams{
		BoardID: boardID,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, err
	}
	threads := make([]models.Thread, 0, len(rows))
	for _, row := range rows {
		threads = append(threads, threadFromBoardRow(row))
	}
	return threads, nil
}

func (r *ThreadRepo) CountByBoard(ctx context.Context, boardID string) (int64, error) {
	return r.q.GetBoardThreadCount(ctx, boardID)
}

func (r *ThreadRepo) GetByID(ctx context.Context, id string) (*models.Thread, error) {
	row, err := r.q.GetThreadByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	t := threadFromThreadRow(row)
	return &t, nil
}

func (r *ThreadRepo) Create(ctx context.Context, boardID, userID, title, body string, imageURL *string) (*models.Thread, error) {
	row, err := r.q.CreateThread(ctx, db.CreateThreadParams{
		BoardID:  boardID,
		UserID:   userID,
		Title:    title,
		Body:     body,
		ImageUrl: pgTextOrNil(imageURL),
	})
	if err != nil {
		return nil, err
	}
	return &models.Thread{
		ID:        row.ID,
		BoardID:   row.BoardID,
		UserID:    row.UserID,
		Title:     row.Title,
		Body:      row.Body,
		ImageURL:  row.ImageUrl.String,
		IsPinned:  row.IsPinned,
		IsLocked:  row.IsLocked,
		BumpedAt:  row.BumpedAt,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *ThreadRepo) TouchBump(ctx context.Context, id string) error {
	return r.q.TouchThreadBump(ctx, id)
}

func (r *ThreadRepo) GetBoardID(ctx context.Context, id string) (string, error) {
	boardID, err := r.q.GetThreadBoardID(ctx, id)
	if err != nil {
		return "", mapErr(err)
	}
	return boardID, nil
}

func (r *ThreadRepo) SetPinned(ctx context.Context, id string, pinned bool) error {
	return r.q.SetThreadPinned(ctx, db.SetThreadPinnedParams{ID: id, IsPinned: pinned})
}

func (r *ThreadRepo) SetLocked(ctx context.Context, id string, locked bool) error {
	return r.q.SetThreadLocked(ctx, db.SetThreadLockedParams{ID: id, IsLocked: locked})
}

func (r *ThreadRepo) Delete(ctx context.Context, id string) error {
	return r.q.DeleteThread(ctx, id)
}

func threadFromBoardRow(row db.GetBoardThreadsRow) models.Thread {
	t := models.Thread{
		ID:           row.ID,
		BoardID:      row.BoardID,
		UserID:       row.UserID,
		Title:        row.Title,
		Body:         row.Body,
		ImageURL:     row.ImageUrl.String,
		IsPinned:     row.IsPinned,
		IsLocked:     row.IsLocked,
		BumpedAt:     row.BumpedAt,
		CreatedAt:    row.CreatedAt,
		BoardSlug:    row.BoardSlug,
		AuthorName:   row.AuthorName,
		AuthorRole:   models.Role(row.AuthorRole),
		AuthorAvatar: row.AuthorAvatar.String,
		AuthorColor:  row.AuthorColor,
		ReplyCount:   row.ReplyCount,
		LastReplyAt: &row.LastReplyAt,
		Bumped:      row.Bumped.Bool,
		BumpLimit:   false,
	}
	return t
}

func threadFromThreadRow(row db.GetThreadByIDRow) models.Thread {
	t := models.Thread{
		ID:           row.ID,
		BoardID:      row.BoardID,
		UserID:       row.UserID,
		Title:        row.Title,
		Body:         row.Body,
		ImageURL:     row.ImageUrl.String,
		IsPinned:     row.IsPinned,
		IsLocked:     row.IsLocked,
		BumpedAt:     row.BumpedAt,
		CreatedAt:    row.CreatedAt,
		BoardSlug:    row.BoardSlug,
		AuthorName:   row.AuthorName,
		AuthorRole:   models.Role(row.AuthorRole),
		AuthorAvatar: row.AuthorAvatar.String,
		AuthorColor:  row.AuthorColor,
		ReplyCount:   row.ReplyCount,
		LastReplyAt: &row.LastReplyAt,
		Bumped:      row.Bumped.Bool,
		BumpLimit:   false,
	}
	return t
}
