package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/davidahmann/relia_oss/internal/auth"
)

func TestAuthorizeRequiresAuth(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv()})

	req := httptest.NewRequest(http.MethodPost, "/v1/authorize", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}
}

func TestAuthorizeWithDevToken(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv()})

	req := httptest.NewRequest(http.MethodPost, "/v1/authorize", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", res.Code)
	}
}

func TestOtherEndpointsRequireAuth(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv()})

	paths := []string{"/v1/approvals/abc", "/v1/verify/abc", "/v1/pack/abc"}
	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for %s, got %d", path, res.Code)
		}
	}
}

func TestSlackInteractionsNoAuth(t *testing.T) {
	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv()})

	req := httptest.NewRequest(http.MethodPost, "/v1/slack/interactions", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", res.Code)
	}
}
