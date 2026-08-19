package services_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/netip"
	"sync"
	"time"

	"github.com/kosero/atchannel/backend/internal/models"
	"github.com/kosero/atchannel/backend/internal/repositories"
)

type InMemoryBoardRepo struct {
	mu     sync.Mutex
	boards []models.Board
}

func NewInMemoryBoardRepo() *InMemoryBoardRepo {
	return &InMemoryBoardRepo{}
}

func id() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (r *InMemoryBoardRepo) List(ctx context.Context) ([]models.Board, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]models.Board, len(r.boards))
	copy(out, r.boards)
	return out, nil
}

func (r *InMemoryBoardRepo) GetBySlug(ctx context.Context, slug string) (*models.Board, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.boards {
		if r.boards[i].Slug == slug {
			b := r.boards[i]
			return &b, nil
		}
	}
	return nil, repositories.ErrNotFound
}

func (r *InMemoryBoardRepo) Create(ctx context.Context, slug, name, description string) (*models.Board, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, b := range r.boards {
		if b.Slug == slug {
			return nil, repositories.ErrBoardExists
		}
	}
	b := models.Board{ID: id(), Slug: slug, Name: name, Description: description, CreatedAt: time.Now()}
	r.boards = append(r.boards, b)
	return &b, nil
}

func (r *InMemoryBoardRepo) Seed(slug, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.boards = append(r.boards, models.Board{ID: id(), Slug: slug, Name: name, CreatedAt: time.Now()})
}

type InMemoryThreadRepo struct {
	mu      sync.Mutex
	threads []models.Thread
}

func NewInMemoryThreadRepo() *InMemoryThreadRepo { return &InMemoryThreadRepo{} }

func (r *InMemoryThreadRepo) ListByBoard(ctx context.Context, boardID string, limit, offset int) ([]models.Thread, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []models.Thread
	for _, t := range r.threads {
		if t.BoardID == boardID {
			out = append(out, t)
		}
	}
	if offset >= len(out) {
		return []models.Thread{}, nil
	}
	end := offset + limit
	if end > len(out) {
		end = len(out)
	}
	return out[offset:end], nil
}

func (r *InMemoryThreadRepo) CountByBoard(ctx context.Context, boardID string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var n int64
	for _, t := range r.threads {
		if t.BoardID == boardID {
			n++
		}
	}
	return n, nil
}

func (r *InMemoryThreadRepo) GetByID(ctx context.Context, id string) (*models.Thread, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.threads {
		if r.threads[i].ID == id {
			t := r.threads[i]
			return &t, nil
		}
	}
	return nil, repositories.ErrNotFound
}

func (r *InMemoryThreadRepo) Create(ctx context.Context, boardID, userID, title, body string, imageURL *string) (*models.Thread, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	img := ""
	if imageURL != nil {
		img = *imageURL
	}
	t := models.Thread{
		ID:        id(),
		BoardID:   boardID,
		UserID:    userID,
		Title:     title,
		Body:      body,
		ImageURL:  img,
		BumpedAt:  now,
		CreatedAt: now,
	}
	r.threads = append(r.threads, t)
	return &t, nil
}

func (r *InMemoryThreadRepo) TouchBump(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.threads {
		if r.threads[i].ID == id {
			r.threads[i].BumpedAt = time.Now()
		}
	}
	return nil
}

func (r *InMemoryThreadRepo) GetBoardID(ctx context.Context, id string) (string, error) {
	t, err := r.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	return t.BoardID, nil
}

func (r *InMemoryThreadRepo) SetPinned(ctx context.Context, id string, pinned bool) error { return nil }
func (r *InMemoryThreadRepo) SetLocked(ctx context.Context, id string, locked bool) error { return nil }
func (r *InMemoryThreadRepo) Delete(ctx context.Context, id string) error                 { return nil }

type InMemoryReplyRepo struct {
	mu      sync.Mutex
	replies []models.Reply
}

func NewInMemoryReplyRepo() *InMemoryReplyRepo { return &InMemoryReplyRepo{} }

func (r *InMemoryReplyRepo) ListByThread(ctx context.Context, threadID string) ([]models.Reply, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []models.Reply
	for _, rp := range r.replies {
		if rp.ThreadID == threadID {
			out = append(out, rp)
		}
	}
	return out, nil
}

func (r *InMemoryReplyRepo) Create(ctx context.Context, threadID, userID, body string, imageURL *string, replyToID *string) (*models.Reply, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	img := ""
	if imageURL != nil {
		img = *imageURL
	}
	rp := models.Reply{
		ID:        id(),
		ThreadID:  threadID,
		UserID:    userID,
		Body:      body,
		ImageURL:  img,
		ReplyToID: replyToID,
		CreatedAt: time.Now(),
	}
	r.replies = append(r.replies, rp)
	return &rp, nil
}

func (r *InMemoryReplyRepo) GetByID(ctx context.Context, id string) (*models.Reply, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.replies {
		if r.replies[i].ID == id {
			rp := r.replies[i]
			return &rp, nil
		}
	}
	return nil, repositories.ErrNotFound
}

func (r *InMemoryReplyRepo) Delete(ctx context.Context, id string) error { return nil }

func (r *InMemoryReplyRepo) ExistsInThread(ctx context.Context, id, threadID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, rp := range r.replies {
		if rp.ID == id && rp.ThreadID == threadID {
			return true, nil
		}
	}
	return false, nil
}

type InMemoryBanRepo struct {
	mu   sync.Mutex
	bans []models.Ban
}

func NewInMemoryBanRepo() *InMemoryBanRepo { return &InMemoryBanRepo{} }

func (r *InMemoryBanRepo) Create(ctx context.Context, userID *string, ip *netip.Addr, reason string, expiresAt *time.Time) (*models.Ban, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b := models.Ban{ID: id(), UserID: userID, IP: ip, Reason: reason, ExpiresAt: expiresAt, CreatedAt: time.Now()}
	r.bans = append(r.bans, b)
	return &b, nil
}

func (r *InMemoryBanRepo) GetActiveByUser(ctx context.Context, userID string) (*models.Ban, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.bans {
		if r.bans[i].UserID != nil && *r.bans[i].UserID == userID {
			if r.bans[i].ExpiresAt == nil || r.bans[i].ExpiresAt.After(time.Now()) {
				b := r.bans[i]
				return &b, nil
			}
		}
	}
	return nil, repositories.ErrNotFound
}

func (r *InMemoryBanRepo) GetActiveByIP(ctx context.Context, ip netip.Addr) (*models.Ban, error) {
	return nil, repositories.ErrNotFound
}

func (r *InMemoryBanRepo) ListActive(ctx context.Context) ([]models.Ban, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.bans, nil
}

type InMemoryReportRepo struct{}

func NewInMemoryReportRepo() *InMemoryReportRepo { return &InMemoryReportRepo{} }

func (r *InMemoryReportRepo) Create(ctx context.Context, targetType, targetID, reporterID, reason string) (*models.Report, error) {
	return &models.Report{ID: id(), TargetType: targetType, TargetID: targetID, ReporterID: reporterID, Reason: reason, Status: "open", CreatedAt: time.Now()}, nil
}
func (r *InMemoryReportRepo) GetByID(ctx context.Context, id string) (*models.Report, error) {
	return nil, repositories.ErrNotFound
}
func (r *InMemoryReportRepo) UpdateStatus(ctx context.Context, id, status string) (*models.Report, error) {
	return &models.Report{ID: id, Status: status}, nil
}
func (r *InMemoryReportRepo) List(ctx context.Context, status string, limit, offset int) ([]models.Report, error) {
	return nil, nil
}
