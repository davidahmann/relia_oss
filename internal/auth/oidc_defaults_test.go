package auth

import (
	"testing"
	"time"
)

func TestNewGitHubOIDCAuthenticatorHasTimeout(t *testing.T) {
	a := NewGitHubOIDCAuthenticator("")
	if a.Audience != "relia" {
		t.Fatalf("expected default audience relia, got %q", a.Audience)
	}
	if a.http == nil {
		t.Fatalf("expected http client")
	}
	if a.http.Timeout != 5*time.Second {
		t.Fatalf("expected 5s timeout, got %s", a.http.Timeout)
	}
}
