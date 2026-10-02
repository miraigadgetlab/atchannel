package middleware

import (
	"testing"
	"time"
)

func TestAccountLimiterAllowsUntilMax(t *testing.T) {
	l := NewAccountLimiter(3, time.Minute)

	for i := 0; i < 2; i++ {
		l.Fail("user@example.com")
		if _, blocked := l.Blocked("user@example.com"); blocked {
			t.Fatalf("blocked after %d failures, want allowed until 3", i+1)
		}
	}

	l.Fail("user@example.com")
	if _, blocked := l.Blocked("user@example.com"); !blocked {
		t.Fatal("not blocked after reaching max failures")
	}
}

func TestAccountLimiterIsPerAccount(t *testing.T) {
	l := NewAccountLimiter(1, time.Minute)

	l.Fail("a@example.com")

	if _, blocked := l.Blocked("b@example.com"); blocked {
		t.Error("unrelated account is blocked by another account's failures")
	}
	if _, blocked := l.Blocked("a@example.com"); !blocked {
		t.Error("target account should be blocked")
	}
}

func TestAccountLimiterNormalizesEmail(t *testing.T) {
	l := NewAccountLimiter(1, time.Minute)

	l.Fail("  User@Example.COM  ")

	if _, blocked := l.Blocked("user@example.com"); !blocked {
		t.Error("expected case/whitespace insensitive key")
	}
}

func TestAccountLimiterClearResets(t *testing.T) {
	l := NewAccountLimiter(2, time.Minute)

	l.Fail("user@example.com")
	l.Fail("user@example.com")
	l.Clear("user@example.com")

	if _, blocked := l.Blocked("user@example.com"); blocked {
		t.Error("still blocked after a successful login cleared the counter")
	}
}

func TestAccountLimiterWindowExpires(t *testing.T) {
	l := NewAccountLimiter(1, 10*time.Millisecond)

	l.Fail("user@example.com")

	if _, blocked := l.Blocked("user@example.com"); !blocked {
		t.Fatal("should be blocked immediately")
	}

	time.Sleep(20 * time.Millisecond)

	if _, blocked := l.Blocked("user@example.com"); blocked {
		t.Error("still blocked after the window elapsed")
	}
}

func TestAccountLimiterBlockedWaitIsBounded(t *testing.T) {
	window := time.Minute
	l := NewAccountLimiter(1, window)

	l.Fail("user@example.com")

	wait, blocked := l.Blocked("user@example.com")
	if !blocked {
		t.Fatal("expected blocked")
	}
	if wait <= 0 || wait > window {
		t.Errorf("wait = %v, want in (0, %s]", wait, window)
	}
}
