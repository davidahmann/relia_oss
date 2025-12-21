package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/davidahmann/relia/internal/auth"
	"github.com/davidahmann/relia/internal/slack"
)

func TestSlackApprovalFlowEndToEnd(t *testing.T) {
	os.Setenv("RELIA_DEV_TOKEN", "test-token")
	defer os.Unsetenv("RELIA_DEV_TOKEN")

	service, err := NewAuthorizeService("../../policies/relia.yaml")
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	signingSecret := "secret"
	h := &Handler{
		Auth:             auth.NewAuthenticatorFromEnv(),
		AuthorizeService: service,
		SlackHandler: &slack.InteractionHandler{
			SigningSecret: signingSecret,
			Approver:      service,
		},
	}

	srv := httptest.NewServer(NewRouter(h))
	defer srv.Close()

	// Trigger approval-required flow (prod terraform rule).
	approvalID := authorizeForApproval(t, srv.URL)

	// Approve via Slack interaction callback.
	postSlackApprove(t, srv.URL, signingSecret, approvalID)

	// Re-call authorize; should mint creds (dev broker) and return allow.
	resp := authorize(t, srv.URL, `{"action":"terraform.apply","resource":"res","env":"prod","request_id":"req-1"}`)
	if resp.Verdict != string(VerdictAllow) {
		t.Fatalf("expected allow, got %s", resp.Verdict)
	}
	if resp.AWSCredentials == nil || resp.AWSCredentials.AccessKeyID == "" {
		t.Fatalf("expected aws credentials")
	}
}

type authorizeResponse struct {
	Verdict        string `json:"verdict"`
	ReceiptID      string `json:"receipt_id"`
	AWSCredentials *struct {
		AccessKeyID string `json:"access_key_id"`
	} `json:"aws_credentials,omitempty"`
	Approval *struct {
		ApprovalID string `json:"approval_id"`
		Status     string `json:"status"`
	} `json:"approval,omitempty"`
}

func authorizeForApproval(t *testing.T, baseURL string) string {
	t.Helper()

	resp := authorize(t, baseURL, `{"action":"terraform.apply","resource":"res","env":"prod","request_id":"req-1"}`)
	if resp.Verdict != string(VerdictRequireApproval) {
		t.Fatalf("expected require_approval, got %s", resp.Verdict)
	}
	if resp.Approval == nil || resp.Approval.ApprovalID == "" {
		t.Fatalf("missing approval")
	}
	return resp.Approval.ApprovalID
}

func authorize(t *testing.T, baseURL, body string) authorizeResponse {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/authorize", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("new req: %v", err)
	}
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("authorize status: %d", res.StatusCode)
	}

	var resp authorizeResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ReceiptID == "" {
		t.Fatalf("missing receipt_id")
	}
	return resp
}

func postSlackApprove(t *testing.T, baseURL, signingSecret, approvalID string) {
	t.Helper()

	payload := fmt.Sprintf(`{"actions":[{"action_id":"approve","value":"%s"}]}`, approvalID)
	form := url.Values{}
	form.Set("payload", payload)
	body := []byte(form.Encode())

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	base := []byte("v0:" + timestamp + ":" + string(body))
	mac := hmac.New(sha256.New, []byte(signingSecret))
	_, _ = mac.Write(base)
	sig := "v0=" + hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/slack/interactions", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("new req: %v", err)
	}
	req.Header.Set("X-Slack-Signature", sig)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("slack: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("slack status: %d", res.StatusCode)
	}
}
