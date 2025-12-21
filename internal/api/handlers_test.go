package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/davidahmann/relia/internal/auth"
	"github.com/davidahmann/relia/internal/slack"
)

func TestAuthorizeRequiresAuth(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})

	req := httptest.NewRequest(http.MethodPost, "/v1/authorize", bytes.NewBufferString(`{"action":"terraform.apply","resource":"res","env":"prod"}`))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}
}

func TestAuthorizeWithDevToken(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})

	req := httptest.NewRequest(http.MethodPost, "/v1/authorize", bytes.NewBufferString(`{"action":"terraform.apply","resource":"res","env":"prod"}`))
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestAuthorizeInvalidJSON(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})

	req := httptest.NewRequest(http.MethodPost, "/v1/authorize", bytes.NewBufferString("{invalid"))
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestAuthorizeServiceNotConfigured(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: nil})

	req := httptest.NewRequest(http.MethodPost, "/v1/authorize", bytes.NewBufferString(`{"action":"terraform.apply","resource":"res","env":"prod"}`))
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

	service := newTestService(t, "../../policies/relia.yaml")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})

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

func TestHealthz(t *testing.T) {
	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: nil})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestSlackInteractionsNoAuth(t *testing.T) {
	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: nil})

	req := httptest.NewRequest(http.MethodPost, "/v1/slack/interactions", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", res.Code)
	}
}

func TestSlackInteractionsHandlerConfigured(t *testing.T) {
	handler := &Handler{
		Auth:         auth.NewAuthenticatorFromEnv(),
		SlackHandler: &slack.InteractionHandler{},
	}
	router := NewRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/v1/slack/interactions", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", res.Code)
	}
}

func TestApprovalsEndpoint(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-prod",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	resp, err := service.Authorize(claims, AuthorizeRequest{Action: "terraform.apply", Resource: "res", Env: "prod"}, "2025-12-20T16:34:14Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if resp.Approval == nil || resp.Approval.ApprovalID == "" {
		t.Fatalf("expected approval id")
	}

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})
	req := httptest.NewRequest(http.MethodGet, "/v1/approvals/"+resp.Approval.ApprovalID, nil)
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestApprovalsMissingID(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})
	req := httptest.NewRequest(http.MethodGet, "/v1/approvals/", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestApprovalsNotFound(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})
	req := httptest.NewRequest(http.MethodGet, "/v1/approvals/missing", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.Code)
	}
}

func TestVerifyEndpoint(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-prod",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	resp, err := service.Authorize(claims, AuthorizeRequest{Action: "terraform.apply", Resource: "res", Env: "prod"}, "2025-12-20T16:34:14Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})
	req := httptest.NewRequest(http.MethodGet, "/v1/verify/"+resp.ReceiptID, nil)
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestPackEndpoint(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-prod",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	resp, err := service.Authorize(claims, AuthorizeRequest{Action: "terraform.apply", Resource: "res", Env: "prod"}, "2025-12-20T16:34:14Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})
	req := httptest.NewRequest(http.MethodGet, "/v1/pack/"+resp.ReceiptID, nil)
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	if ct := res.Header().Get("Content-Type"); ct != "application/zip" {
		t.Fatalf("expected zip content-type, got %s", ct)
	}
}

func TestVerifyMissingReceiptID(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})
	req := httptest.NewRequest(http.MethodGet, "/v1/verify/", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestVerifyReceiptNotFound(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})
	req := httptest.NewRequest(http.MethodGet, "/v1/verify/missing", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.Code)
	}
}

func TestVerifyInvalidSignature(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-prod",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	resp, err := service.Authorize(claims, AuthorizeRequest{Action: "terraform.apply", Resource: "res", Env: "prod"}, "2025-12-20T16:34:14Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}

	receipt, ok := service.Ledger.GetReceipt(resp.ReceiptID)
	if !ok {
		t.Fatalf("receipt missing")
	}
	receipt.Sig = []byte("bad")
	if err := service.Ledger.PutReceipt(receipt); err != nil {
		t.Fatalf("put receipt: %v", err)
	}

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})
	req := httptest.NewRequest(http.MethodGet, "/v1/verify/"+resp.ReceiptID, nil)
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestPackMissingReceiptID(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})
	req := httptest.NewRequest(http.MethodGet, "/v1/pack/", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestPackReceiptNotFound(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})
	req := httptest.NewRequest(http.MethodGet, "/v1/pack/missing", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.Code)
	}
}

func TestPackMissingContext(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service := newTestService(t, "../../policies/relia.yaml")

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-prod",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	resp, err := service.Authorize(claims, AuthorizeRequest{Action: "terraform.apply", Resource: "res", Env: "prod"}, "2025-12-20T16:34:14Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}

	receipt, ok := service.Ledger.GetReceipt(resp.ReceiptID)
	if !ok {
		t.Fatalf("receipt missing")
	}
	receipt.ContextID = "missing"
	if err := service.Ledger.PutReceipt(receipt); err != nil {
		t.Fatalf("put receipt: %v", err)
	}

	router := NewRouter(&Handler{Auth: auth.NewAuthenticatorFromEnv(), AuthorizeService: service})
	req := httptest.NewRequest(http.MethodGet, "/v1/pack/"+resp.ReceiptID, nil)
	req.Header.Set("Authorization", "Bearer test-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.Code)
	}
}
