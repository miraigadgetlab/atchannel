package validation

import "testing"

func TestChannelNameError(t *testing.T) {
	valid := []string{"tech", "a1", "go-lang", "c_a", "0day", "x"[:1] + stringsRepeat("y", 31)}

	for _, name := range valid {
		if got := ChannelNameError(name); got != "" {
			t.Errorf("ChannelNameError(%q) = %q, want \"\" (should be valid)", name, got)
		}
	}

	invalid := []struct {
		name   string
		reason string
	}{
		{"", "empty"},
		{"a", "too short"},
		{"Tech", "uppercase"},
		{"-tech", "leading hyphen"},
		{"_tech", "leading underscore"},
		{"tech!", "punctuation"},
		{"te ch", "space"},
		{"te/ch", "slash"},
		{"тех", "non-ascii"},
		{stringsRepeat("a", 33), "too long"},
	}

	for _, tc := range invalid {
		if got := ChannelNameError(tc.name); got == "" {
			t.Errorf("ChannelNameError(%q) = \"\", want rejection (%s)", tc.name, tc.reason)
		}
	}
}

func TestNameError(t *testing.T) {
	if got := NameError("alice"); got != "" {
		t.Errorf("NameError(\"alice\") = %q, want \"\"", got)
	}

	invalid := []struct {
		name   string
		reason string
	}{
		{"", "empty"},
		{"ab", "too short"},
		{"a" + stringsRepeat("b", 32), "too long"},
		{" alice", "leading space"},
		{"ali ce", "space"},
		{"al/ce", "slash"},
		{".alice", "leading dot"},
	}

	for _, tc := range invalid {
		if got := NameError(tc.name); got == "" {
			t.Errorf("NameError(%q) = \"\", want rejection (%s)", tc.name, tc.reason)
		}
	}
}

func TestEmailError(t *testing.T) {
	valid := []string{"a@b.co", "first.last@example.com", "user+tag@sub.example.org"}

	for _, email := range valid {
		if got := EmailError(email); got != "" {
			t.Errorf("EmailError(%q) = %q, want \"\"", email, got)
		}
	}

	invalid := []struct {
		email  string
		reason string
	}{
		{"", "empty"},
		{"abc", "no at sign"},
		{"@example.com", "no local part"},
		{"user@", "no domain"},
		{"a@@b.com", "double at"},
		{"a@b@c.com", "two at signs"},
		{stringsRepeat("a", 250) + "@b.co", "too long"},
	}

	for _, tc := range invalid {
		if got := EmailError(tc.email); got == "" {
			t.Errorf("EmailError(%q) = \"\", want rejection (%s)", tc.email, tc.reason)
		}
	}
}

func TestTooLongCountsRunesNotBytes(t *testing.T) {
	// Four emoji are 4 runes but 16 bytes: a byte-based check would reject
	// this well under the limit.
	emoji := "😀😁😂🤣"
	if got := TooLong("title", emoji, 10); got != "" {
		t.Errorf("TooLong(emoji, 10) = %q, want \"\"", got)
	}

	if got := TooLong("title", stringsRepeat("x", 11), 10); got == "" {
		t.Error("TooLong(11 chars, 10) = \"\", want rejection")
	}
}

func TestLengthCountsRunes(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"abc", 3},
		{"😀", 1},
		{"日本語", 3},
	}

	for _, tc := range cases {
		if got := Length(tc.in); got != tc.want {
			t.Errorf("Length(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func stringsRepeat(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}
