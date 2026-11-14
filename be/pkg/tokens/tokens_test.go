package tokens

import (
	"errors"
	"testing"
	"time"
)

func TestJWTIssueAndVerify(t *testing.T) {
	m := NewJWTManager("test-secret-test-secret-test-secret", time.Minute)
	token, exp, err := m.Issue("user-1")
	if err != nil {
		t.Fatal(err)
	}
	if time.Until(exp) <= 0 {
		t.Fatalf("expiry should be in the future")
	}
	sub, err := m.Verify(token)
	if err != nil || sub != "user-1" {
		t.Fatalf("got sub=%q err=%v", sub, err)
	}
}

func TestJWTExpired(t *testing.T) {
	m := NewJWTManager("test-secret-test-secret-test-secret", time.Minute)
	m.now = func() time.Time { return time.Now().Add(-2 * time.Minute) }
	token, _, _ := m.Issue("user-1")
	m.now = time.Now
	if _, err := m.Verify(token); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestJWTWrongSecretAndGarbage(t *testing.T) {
	a := NewJWTManager("secret-a-secret-a-secret-a-secret-a", time.Minute)
	b := NewJWTManager("secret-b-secret-b-secret-b-secret-b", time.Minute)
	token, _, _ := a.Issue("user-1")
	if _, err := b.Verify(token); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
	if _, err := a.Verify("not-a-jwt"); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestOpaqueTokenAndHash(t *testing.T) {
	x, _ := NewOpaqueToken()
	y, _ := NewOpaqueToken()
	if x == y || len(x) < 40 {
		t.Fatalf("tokens should be random and long: %q %q", x, y)
	}
	if HashToken(x) != HashToken(x) || HashToken(x) == HashToken(y) || len(HashToken(x)) != 64 {
		t.Fatal("hash should be deterministic sha256 hex")
	}
}
