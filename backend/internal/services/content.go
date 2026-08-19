package services

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/net/html"

	"github.com/kosero/atchannel/backend/internal/models"
	"github.com/kosero/atchannel/backend/internal/repositories"
	"github.com/kosero/atchannel/backend/internal/repositories/postgres"
)

var (
	ErrBoardNotFound    = errors.New("board not found")
	ErrBoardExists      = errors.New("board already exists")
	ErrInvalidBoardSlug = errors.New("invalid board slug")
	ErrBoardNameTooLong = errors.New("board name too long")
	ErrBoardDescTooLong = errors.New("board description too long")
	ErrThreadNotFound   = errors.New("thread not found")
	ErrReplyNotFound    = errors.New("reply not found")
	ErrThreadLocked     = errors.New("thread is locked")
)

const (
	maxTitleLength  = 200
	maxBodyLength   = 4000
	maxSlugLength   = 20
	maxBoardNameLen = 60
	maxBoardDescLen = 200
)

// sanitizeText strips all HTML markup and keeps only the raw text content.
// Content is displayed as plain text, so any stored HTML is rejected at the
// API boundary to prevent stored XSS in any client that renders it raw.
func sanitizeText(s string) string {
	var b strings.Builder
	z := html.NewTokenizer(strings.NewReader(s))
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt == html.TextToken {
			b.Write(z.Text())
		}
	}
	return b.String()
}

type BoardService struct {
	boards repositories.BoardRepository
}

func NewBoardService(boards repositories.BoardRepository) *BoardService {
	return &BoardService{boards: boards}
}

func (s *BoardService) List(ctx context.Context) ([]models.Board, error) {
	return s.boards.List(ctx)
}

func (s *BoardService) Create(ctx context.Context, slug, name, description string) (*models.Board, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if slug == "" || name == "" {
		return nil, ErrInvalidBoardSlug
	}
	if len(slug) > maxSlugLength {
		return nil, ErrInvalidBoardSlug
	}
	if len(name) > maxBoardNameLen {
		return nil, ErrBoardNameTooLong
	}
	if len(description) > maxBoardDescLen {
		return nil, ErrBoardDescTooLong
	}
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return nil, ErrInvalidBoardSlug
		}
	}

	board, err := s.boards.Create(ctx, slug, name, description)
	if err != nil {
		if errors.Is(err, repositories.ErrBoardExists) {
			return nil, ErrBoardExists
		}
		return nil, err
	}
	return board, nil
}

type ThreadService struct {
	threads  repositories.ThreadRepository
	replies  repositories.ReplyRepository
	reports  repositories.ReportRepository
	boards   repositories.BoardRepository
	bans     repositories.BanRepository
	uploads  *UploadService
}

func NewThreadService(
	threads repositories.ThreadRepository,
	replies repositories.ReplyRepository,
	reports repositories.ReportRepository,
	boards repositories.BoardRepository,
	bans repositories.BanRepository,
	uploads *UploadService,
) *ThreadService {
	return &ThreadService{
		threads: threads,
		replies: replies,
		reports: reports,
		boards:  boards,
		bans:    bans,
		uploads: uploads,
	}
}

func (s *ThreadService) ListByBoard(ctx context.Context, slug string, page, perPage int) ([]models.Thread, int64, error) {
	board, err := s.boards.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, 0, ErrBoardNotFound
		}
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage
	total, err := s.threads.CountByBoard(ctx, board.ID)
	if err != nil {
		return nil, 0, err
	}
	threads, err := s.threads.ListByBoard(ctx, board.ID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	return threads, total, nil
}

func (s *ThreadService) Get(ctx context.Context, id string) (*models.Thread, error) {
	thread, err := s.threads.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, ErrThreadNotFound
		}
		return nil, err
	}
	return thread, nil
}

func (s *ThreadService) Create(ctx context.Context, user *models.User, boardSlug, title, body, imageURL string) (*models.Thread, error) {
	board, err := s.boards.GetBySlug(ctx, boardSlug)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, ErrBoardNotFound
		}
		return nil, err
	}
	if ban, err := s.bans.GetActiveByUser(ctx, user.ID); err == nil && ban != nil {
		return nil, ErrBanned
	} else if err != nil && !errors.Is(err, postgres.ErrNotFound) {
		return nil, err
	}
	title = sanitizeText(title)
	body = sanitizeText(body)
	if len(title) > maxTitleLength {
		return nil, ErrTitleTooLong
	}
	if len(body) > maxBodyLength {
		return nil, ErrBodyTooLong
	}
	var imageURLPtr *string
	if imageURL != "" {
		imageURLPtr = &imageURL
	}
	thread, err := s.threads.Create(ctx, board.ID, user.ID, title, body, imageURLPtr)
	if err != nil {
		return nil, err
	}
	return thread, nil
}

func (s *ThreadService) Reply(ctx context.Context, user *models.User, threadID, body, imageURL string, replyToID *string) (*models.Reply, error) {
	thread, err := s.threads.GetByID(ctx, threadID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, ErrThreadNotFound
		}
		return nil, err
	}
	if thread.IsLocked {
		return nil, ErrThreadLocked
	}
	if ban, err := s.bans.GetActiveByUser(ctx, user.ID); err == nil && ban != nil {
		return nil, ErrBanned
	} else if err != nil && !errors.Is(err, postgres.ErrNotFound) {
		return nil, err
	}
	body = sanitizeText(body)
	if len(body) > maxBodyLength {
		return nil, ErrBodyTooLong
	}
	if replyToID != nil {
		ok, err := s.replies.ExistsInThread(ctx, *replyToID, threadID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrReplyNotFound
		}
	}
	var imageURLPtr *string
	if imageURL != "" {
		imageURLPtr = &imageURL
	}
	reply, err := s.replies.Create(ctx, threadID, user.ID, body, imageURLPtr, replyToID)
	if err != nil {
		return nil, err
	}
	if err := s.threads.TouchBump(ctx, threadID); err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *ThreadService) ListReplies(ctx context.Context, threadID string) ([]models.Reply, error) {
	if _, err := s.threads.GetBoardID(ctx, threadID); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, ErrThreadNotFound
		}
		return nil, err
	}
	return s.replies.ListByThread(ctx, threadID)
}

func (s *ThreadService) DeleteThread(ctx context.Context, id string) error {
	if err := s.threads.Delete(ctx, id); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return ErrThreadNotFound
		}
		return err
	}
	return nil
}
