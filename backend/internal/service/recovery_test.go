package service

import (
	"strings"
	"testing"
)

// A code typed by hand (lowercase, spaces instead of the dash) has to
// normalize to exactly what was stored, or a valid code is rejected.
func TestNormalizeRecoveryCode(t *testing.T) {
	cases := map[string]string{
		"ABCDE-FGHJK": "ABCDEFGHJK",
		"abcde-fghjk": "ABCDEFGHJK",
		"abcde fghjk": "ABCDEFGHJK",
		"ABCDEFGHJK":  "ABCDEFGHJK",
		"  ab-cd  ":   "ABCD",
	}

	for in, want := range cases {
		if got := normalizeRecoveryCode(in); got != want {
			t.Errorf("normalizeRecoveryCode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRandomRecoveryCodeShape(t *testing.T) {
	seen := make(map[string]bool)

	for i := 0; i < 100; i++ {
		code, err := randomRecoveryCode()
		if err != nil {
			t.Fatalf("randomRecoveryCode: %v", err)
		}
		if len(code) != 11 || code[5] != '-' {
			t.Fatalf("code = %q, want XXXXX-XXXXX", code)
		}
		for _, r := range strings.ReplaceAll(code, "-", "") {
			if !strings.ContainsRune(recoveryAlphabet, r) {
				t.Fatalf("code %q contains %q, which is not in the alphabet", code, r)
			}
		}
		if seen[code] {
			t.Fatalf("duplicate code generated: %q", code)
		}
		seen[code] = true
	}
}

func TestRecoveryCodeNormalizesBeforeHashing(t *testing.T) {
	code, err := randomRecoveryCode()
	if err != nil {
		t.Fatalf("randomRecoveryCode: %v", err)
	}

	typed := strings.ToLower(strings.ReplaceAll(code, "-", " "))

	if HashToken(normalizeRecoveryCode(code)) != HashToken(normalizeRecoveryCode(typed)) {
		t.Errorf("hand-typed code %q does not hash the same as the stored form %q", typed, code)
	}
}
