package handler

import (
	"encoding/json"
	"strings"
	"testing"

	"atchannel-backend/internal/models"
)

// The author view embedded in posts/comments must never carry the private
// account fields — this is the regression guard for the email leak.
func TestPostResponseHidesPrivateUserFields(t *testing.T) {
	post := models.Post{
		ID:      1,
		Title:   "hello",
		Content: "body",
		User: models.User{
			ID:       2,
			Name:     "alice",
			Email:    "alice@example.com",
			Password: "bcrypt-hash",
			Roles:    []string{"admin"},
			Note:     "private admin note",
		},
	}

	raw, err := json.Marshal(toPostResponse(post, false))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(raw)

	for _, secret := range []string{"alice@example.com", "bcrypt-hash", "private admin note", `"email"`, `"password"`, `"roles"`, `"note"`} {
		if strings.Contains(body, secret) {
			t.Errorf("serialized post leaks %q: %s", secret, body)
		}
	}
	if !strings.Contains(body, `"name":"alice"`) {
		t.Errorf("body = %s, want the author name to remain public", body)
	}
}

// The channeler leaderboard is public like the author view, so it must not
// leak the private account fields either (email, roles, password hash).
func TestChannelerResponseHidesPrivateUserFields(t *testing.T) {
	user := models.User{
		ID:       3,
		Name:     "carol",
		Email:    "carol@example.com",
		Password: "bcrypt-hash",
		Roles:    []string{"moderator"},
		Note:     "private admin note",
	}

	raw, err := json.Marshal(toChannelerResponse(user, 42))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(raw)

	for _, secret := range []string{"carol@example.com", "bcrypt-hash", "private admin note", `"email"`, `"password"`, `"roles"`, `"note"`} {
		if strings.Contains(body, secret) {
			t.Errorf("serialized channeler leaks %q: %s", secret, body)
		}
	}
	if !strings.Contains(body, `"post_count":42`) {
		t.Errorf("body = %s, want the post count to be public", body)
	}
}

func TestCommentResponseHidesPrivateUserFields(t *testing.T) {
	comment := models.Comment{
		ID:      1,
		Content: "hi",
		User: models.User{
			ID:       2,
			Name:     "bob",
			Email:    "bob@example.com",
			Password: "bcrypt-hash",
			Roles:    []string{"admin"},
			Note:     "private admin note",
		},
	}

	raw, err := json.Marshal(toCommentResponse(comment))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(raw)

	for _, secret := range []string{"bob@example.com", "bcrypt-hash", "private admin note", `"email"`, `"password"`, `"roles"`, `"note"`} {
		if strings.Contains(body, secret) {
			t.Errorf("serialized comment leaks %q: %s", secret, body)
		}
	}
}
