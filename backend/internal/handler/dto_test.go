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
		},
	}

	raw, err := json.Marshal(toPostResponse(post))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(raw)

	for _, secret := range []string{"alice@example.com", "bcrypt-hash", `"email"`, `"password"`, `"roles"`} {
		if strings.Contains(body, secret) {
			t.Errorf("serialized post leaks %q: %s", secret, body)
		}
	}
	if !strings.Contains(body, `"name":"alice"`) {
		t.Errorf("body = %s, want the author name to remain public", body)
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
		},
	}

	raw, err := json.Marshal(toCommentResponse(comment))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(raw)

	for _, secret := range []string{"bob@example.com", "bcrypt-hash", `"email"`, `"password"`, `"roles"`} {
		if strings.Contains(body, secret) {
			t.Errorf("serialized comment leaks %q: %s", secret, body)
		}
	}
}
