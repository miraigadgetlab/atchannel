package services_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/kosero/atchannel/backend/internal/models"
	"github.com/kosero/atchannel/backend/internal/repositories"
)

// InMemoryRefreshRepo is a test double for the refresh-token repository,
// implementing the token-family rotation state machine in memory.
type InMemoryRefreshRepo struct {
	mu      sync.Mutex
	tokens  []*models.RefreshToken
	nextSeq int
}

func NewInMemoryRefreshRepo() *InMemoryRefreshRepo {
	return &InMemoryRefreshRepo{}
}

func (r *InMemoryRefreshRepo) id() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	r.nextSeq++
	return hex.EncodeToString(b)
}

func (r *InMemoryRefreshRepo) Create(ctx context.Context, userID, familyID, tokenHash string, expiresAt time.Time) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t := &models.RefreshToken{
		ID:        r.id(),
		UserID:    userID,
		FamilyID:  familyID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	r.tokens = append(r.tokens, t)
	return t.ID, nil
}

func (r *InMemoryRefreshRepo) GetByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.tokens {
		if t.TokenHash == tokenHash {
			return t, nil
		}
	}
	return nil, repositories.ErrNotFound
}

func (r *InMemoryRefreshRepo) Revoke(ctx context.Context, id string, replacedBy *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.tokens {
		if t.ID == id {
			t.Revoked = true
			t.ReplacedBy = replacedBy
			return nil
		}
	}
	return nil
}

func (r *InMemoryRefreshRepo) RevokeFamily(ctx context.Context, familyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.tokens {
		if t.FamilyID == familyID {
			t.Revoked = true
		}
	}
	return nil
}

func (r *InMemoryRefreshRepo) GetActiveByFamily(ctx context.Context, familyID string) (*models.RefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.tokens {
		if t.FamilyID == familyID && !t.Revoked {
			return t, nil
		}
	}
	return nil, repositories.ErrNotFound
}

func (r *InMemoryRefreshRepo) DeleteExpired(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	out := r.tokens[:0]
	for _, t := range r.tokens {
		if !t.ExpiresAt.Before(now) {
			out = append(out, t)
		}
	}
	r.tokens = out
	return nil
}
