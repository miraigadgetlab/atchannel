package middleware

import (
	"testing"
	"time"
)

func TestAccountLimiterAllowsUntilMax(t *testing.T) {
	l := NewAccountLimiter(3, time.Minute)

	for i := 0; i < 2; i++ {
		l.Fail("alice")
		if _, blocked := l.Blocked("alice"); blocked {
			t.Fatalf("blocked after %d failures, want allowed until 3", i+1)
		}
	}

	l.Fail("alice")
	if _, blocked := l.Blocked("alice"); !blocked {
		t.Fatal("not blocked after reaching max failures")
	}
}

func TestAccountLimiterIsPerAccount(t *testing.T) {
	l := NewAccountLimiter(1, time.Minute)

	l.Fail("alice")

	if _, blocked := l.Blocked("bob"); blocked {
		t.Error("unrelated account is blocked by another account's failures")
	}
	if _, blocked := l.Blocked("alice"); !blocked {
		t.Error("target account should be blocked")
	}
}

func TestAccountLimiterNormalizesName(t *testing.T) {
	l := NewAccountLimiter(1, time.Minute)

	l.Fail("  Alice  ")

	if _, blocked := l.Blocked("alice"); !blocked {
		t.Error("expected case/whitespace insensitive key")
	}
}

func TestAccountLimiterClearResets(t *testing.T) {
	l := NewAccountLimiter(2, time.Minute)

	l.Fail("alice")
	l.Fail("alice")
	l.Clear("alice")

	if _, blocked := l.Blocked("alice"); blocked {
		t.Error("still blocked after a successful login cleared the counter")
	}
}

func TestAccountLimiterWindowExpires(t *testing.T) {
	l := NewAccountLimiter(1, 10*time.Millisecond)

	l.Fail("alice")

	if _, blocked := l.Blocked("alice"); !blocked {
		t.Fatal("should be blocked immediately")
	}

	time.Sleep(20 * time.Millisecond)

	if _, blocked := l.Blocked("alice"); blocked {
		t.Error("still blocked after the window elapsed")
	}
}

func TestAccountLimiterBlockedWaitIsBounded(t *testing.T) {
	window := time.Minute
	l := NewAccountLimiter(1, window)

	l.Fail("alice")

	wait, blocked := l.Blocked("alice")
	if !blocked {
		t.Fatal("expected blocked")
	}
	if wait <= 0 || wait > window {
		t.Errorf("wait = %v, want in (0, %s]", wait, window)
	}
}
