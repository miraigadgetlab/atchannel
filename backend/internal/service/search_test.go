package service

import "testing"

func TestLikePatternEscapesWildcards(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"tech", "%tech%"},
		{"100%", `%100\%%`},
		{"snake_case", `%snake\_case%`},
		{`back\slash`, `%back\\slash%`},
		{"%_%", `%\%\_\%%`},
		{"", "%%"},
	}

	for _, tc := range cases {
		if got := likePattern(tc.in); got != tc.want {
			t.Errorf("likePattern(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
