package services_test

import (
	"context"
	"testing"

	"github.com/kosero/atchannel/backend/internal/models"
	"github.com/kosero/atchannel/backend/internal/services"
)

func newContentServices(t *testing.T) (*services.ThreadService, *services.BoardService, *models.User) {
	t.Helper()
	boards := NewInMemoryBoardRepo()
	boards.Seed("debug", "Debug")

	users := NewInMemoryUserRepo()
	user := &models.User{Username: "poster", Email: "p@example.com", Role: models.RoleUser}
	if err := users.Create(context.Background(), user); err != nil {
		t.Fatal(err)
	}

	threads := NewInMemoryThreadRepo()
	replies := NewInMemoryReplyRepo()
	reports := NewInMemoryReportRepo()
	bans := NewInMemoryBanRepo()

	threadSvc := services.NewThreadService(threads, replies, reports, boards, bans, nil)
	boardSvc := services.NewBoardService(boards)
	return threadSvc, boardSvc, user
}

func TestBoardList(t *testing.T) {
	_, boardSvc, _ := newContentServices(t)
	boards, err := boardSvc.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(boards) != 1 {
		t.Fatalf("expected 1 board, got %d", len(boards))
	}
}

func TestCreateAndGetThread(t *testing.T) {
	threadSvc, _, user := newContentServices(t)

	thread, err := threadSvc.Create(context.Background(), user, "debug", "Hello world", "First thread", "")
	if err != nil {
		t.Fatalf("create thread: %v", err)
	}
	if thread.Title != "Hello world" {
		t.Fatalf("unexpected title %q", thread.Title)
	}
	got, err := threadSvc.Get(context.Background(), thread.ID)
	if err != nil {
		t.Fatalf("get thread: %v", err)
	}
	if got.ID != thread.ID {
		t.Fatalf("mismatched thread id")
	}
}

func TestCreateThreadUnknownBoard(t *testing.T) {
	threadSvc, _, user := newContentServices(t)
	_, err := threadSvc.Create(context.Background(), user, "nope", "t", "b", "")
	if err != services.ErrBoardNotFound {
		t.Fatalf("expected ErrBoardNotFound, got %v", err)
	}
}

func TestCreateThreadTitleTooLong(t *testing.T) {
	threadSvc, _, user := newContentServices(t)
	long := make([]rune, 201)
	for i := range long {
		long[i] = 'x'
	}
	_, err := threadSvc.Create(context.Background(), user, "debug", string(long), "b", "")
	if err != services.ErrTitleTooLong {
		t.Fatalf("expected ErrTitleTooLong, got %v", err)
	}
}

func TestReplyToThread(t *testing.T) {
	threadSvc, _, user := newContentServices(t)
	thread, err := threadSvc.Create(context.Background(), user, "debug", "Thread", "op", "")
	if err != nil {
		t.Fatal(err)
	}
	reply, err := threadSvc.Reply(context.Background(), user, thread.ID, "nice", "", nil)
	if err != nil {
		t.Fatalf("reply: %v", err)
	}
	if reply.Body != "nice" {
		t.Fatalf("unexpected reply body %q", reply.Body)
	}
	replies, err := threadSvc.ListReplies(context.Background(), thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(replies) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(replies))
	}
}

func TestReplyUnknownThread(t *testing.T) {
	threadSvc, _, user := newContentServices(t)
	_, err := threadSvc.Reply(context.Background(), user, "does-not-exist", "x", "", nil)
	if err != services.ErrThreadNotFound {
		t.Fatalf("expected ErrThreadNotFound, got %v", err)
	}
}

func TestReplyToMissingPost(t *testing.T) {
	threadSvc, _, user := newContentServices(t)
	thread, _ := threadSvc.Create(context.Background(), user, "debug", "t", "op", "")
	missing := "no-such-reply-id"
	_, err := threadSvc.Reply(context.Background(), user, thread.ID, "x", "", &missing)
	if err != services.ErrReplyNotFound {
		t.Fatalf("expected ErrReplyNotFound, got %v", err)
	}
}

func TestCreateThreadPreservesAngleBrackets(t *testing.T) {
	threadSvc, _, user := newContentServices(t)

	thread, err := threadSvc.Create(
		context.Background(),
		user,
		"debug",
		`#include <stdio.h>`,
		"hello & goodbye",
		"",
	)
	if err != nil {
		t.Fatalf("create thread: %v", err)
	}
	if thread.Title != `#include <stdio.h>` {
		t.Fatalf("unexpected title %q", thread.Title)
	}
	if thread.Body != "hello & goodbye" {
		t.Fatalf("unexpected body %q", thread.Body)
	}
}

func TestCreateThreadTrimsWhitespace(t *testing.T) {
	threadSvc, _, user := newContentServices(t)

	thread, err := threadSvc.Create(context.Background(), user, "debug", "  padded  ", "body", "")
	if err != nil {
		t.Fatalf("create thread: %v", err)
	}
	if thread.Title != "padded" {
		t.Fatalf("unexpected title %q", thread.Title)
	}
}

func TestListByBoardPagination(t *testing.T) {
	threadSvc, _, user := newContentServices(t)
	for i := 0; i < 3; i++ {
		if _, err := threadSvc.Create(context.Background(), user, "debug", "t", "b", ""); err != nil {
			t.Fatal(err)
		}
	}
	threads, total, err := threadSvc.ListByBoard(context.Background(), "debug", 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 2 {
		t.Fatalf("expected 2 threads on page 1, got %d", len(threads))
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
}
