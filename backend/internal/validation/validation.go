package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Channel/Post/Comment limits, kept in one place so handlers and services
// agree on what "too long" means.
const (
	MaxChannelName        = 32
	MaxChannelTitle       = 64
	MaxChannelDescription = 300

	MaxPostTitle   = 200
	MaxPostContent = 65536

	MaxCommentContent = 5000

	MaxAboutMe   = 1000
	MaxAvatarURL = 500
)

// ChannelName is the public handle shown as at/<name>: lowercase, URL safe,
// and short enough to sit in a header without wrapping.
var ChannelName = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,31}$`)

// DisplayName is a display name / username. Case is preserved (it is shown
// to people); uniqueness in the database is what actually matters.
var DisplayName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{2,31}$`)

// Length counts runes, not bytes, so a name full of emoji or CJK is not
// rejected for being "too long" when it is visually short.
func Length(s string) int {
	return utf8.RuneCountInString(s)
}

// TooLong returns a human readable reason when s exceeds max runes.
func TooLong(field string, s string, max int) string {
	if Length(s) > max {
		return fmt.Sprintf("%s must be at most %d characters long", field, max)
	}
	return ""
}

// ChannelNameError returns a reason, or "" when valid.
func ChannelNameError(name string) string {
	switch {
	case Length(name) < 2 || Length(name) > MaxChannelName:
		return fmt.Sprintf("name must be 2-%d characters long", MaxChannelName)
	case !ChannelName.MatchString(name):
		return "name must be lowercase letters, numbers, hyphen or underscore, and start with a letter or number"
	default:
		return ""
	}
}

// NameError returns a reason, or "" when valid.
func NameError(name string) string {
	switch {
	case Length(name) < 3 || Length(name) > 32:
		return "name must be 3-32 characters long"
	case !DisplayName.MatchString(name):
		return "name must start with a letter or number, and may only contain letters, numbers, dot, hyphen or underscore"
	default:
		return ""
	}
}

// EmailError is deliberately permissive: a full RFC 5322 check rejects
// legitimate addresses, and the only real proof is a confirmation link.
// This just rejects obvious garbage.
func EmailError(email string) string {
	at := strings.Index(email, "@")

	switch {
	case Length(email) < 3 || Length(email) > 254:
		return "email must be 3-254 characters long"
	case at <= 0:
		return "invalid email address"
	case strings.Count(email, "@") != 1:
		return "invalid email address"
	case at == len(email)-1:
		return "invalid email address"
	default:
		return ""
	}
}
