package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kosero/atchannel/backend/internal/db"
	"github.com/kosero/atchannel/backend/internal/models"
)

type ReplyRepo struct {
	q *db.Queries
}

func NewReplyRepo(pool *pgxpool.Pool) *ReplyRepo {
	return &ReplyRepo{q: db.New(pool)}
}

func (r *ReplyRepo) ListByThread(ctx context.Context, threadID string) ([]models.Reply, error) {
	rows, err := r.q.ListThreadReplies(ctx, threadID)
	if err != nil {
		return nil, err
	}
	replies := make([]models.Reply, 0, len(rows))
	for _, row := range rows {
		rp := models.Reply{
			ID:           row.ID,
			ThreadID:     row.ThreadID,
			UserID:       row.UserID,
			Body:         row.Body,
			ImageURL:     row.ImageUrl.String,
			CreatedAt:    row.CreatedAt,
			AuthorName:   row.AuthorName,
			AuthorRole:   models.Role(row.AuthorRole),
			AuthorAvatar: row.AuthorAvatar.String,
			AuthorColor:  row.AuthorColor,
		}
		if row.ReplyToID.Valid {
			u := uuid.UUID(row.ReplyToID.Bytes)
			s := u.String()
			rp.ReplyToID = &s
		}
		replies = append(replies, rp)
	}
	return replies, nil
}

func (r *ReplyRepo) Create(ctx context.Context, threadID, userID, body string, imageURL *string, replyToID *string) (*models.Reply, error) {
	row, err := r.q.CreateReply(ctx, db.CreateReplyParams{
		ThreadID:  threadID,
		UserID:    userID,
		Body:      body,
		ImageUrl:  pgTextOrNil(imageURL),
		ReplyToID: pgUUID(replyToID),
	})
	if err != nil {
		return nil, err
	}
	rp := &models.Reply{
		ID:           row.ID,
		ThreadID:     row.ThreadID,
		UserID:       row.UserID,
		Body:         row.Body,
		ImageURL:     row.ImageUrl.String,
		CreatedAt:    row.CreatedAt,
		AuthorName:   row.AuthorName,
		AuthorRole:   models.Role(row.AuthorRole),
		AuthorAvatar: row.AuthorAvatar.String,
		AuthorColor:  row.AuthorColor,
	}
	if row.ReplyToID.Valid {
		u := uuid.UUID(row.ReplyToID.Bytes)
		s := u.String()
		rp.ReplyToID = &s
	}
	return rp, nil
}

func (r *ReplyRepo) GetByID(ctx context.Context, id string) (*models.Reply, error) {
	row, err := r.q.GetReplyByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return &models.Reply{
		ID:       row.ID,
		ThreadID: row.ThreadID,
		UserID:   row.UserID,
		Body:     row.Body,
		ImageURL: row.ImageUrl.String,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *ReplyRepo) Delete(ctx context.Context, id string) error {
	return r.q.DeleteReply(ctx, id)
}

func (r *ReplyRepo) ExistsInThread(ctx context.Context, id, threadID string) (bool, error) {
	return r.q.ReplyExists(ctx, db.ReplyExistsParams{ID: id, ThreadID: threadID})
}
