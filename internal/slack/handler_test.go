package slack

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

type testApprover struct {
	approvalID string
	status     string
}

func (t *testApprover) Approve(approvalID string, status string, createdAt string) (string, error) {
	t.approvalID = approvalID
	t.status = status
	return "receipt-1", nil
}

func TestHandleInteractionsApprove(t *testing.T) {
	payload := `{"actions":[{"action_id":"approve","value":"appr-1"}]}`
	form := url.Values{}
	form.Set("payload", payload)
	body := []byte(form.Encode())

	signingSecret := "secret"
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	base := []byte("v0:" + timestamp + ":" + string(body))
	mac := hmac.New(sha256.New, []byte(signingSecret))
	_, _ = mac.Write(base)
	sig := "v0=" + hex.EncodeToString(mac.Sum(nil))

	approver := &testApprover{}
	h := &InteractionHandler{SigningSecret: signingSecret, Approver: approver}

	req := httptest.NewRequest(http.MethodPost, "/v1/slack/interactions", bytes.NewBuffer(body))
	req.Header.Set("X-Slack-Signature", sig)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res := httptest.NewRecorder()
	h.HandleInteractions(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	if approver.approvalID != "appr-1" || approver.status != "approved" {
		t.Fatalf("unexpected approval %s %s", approver.approvalID, approver.status)
	}
}

func TestHandleInteractionsDenyFallbackValue(t *testing.T) {
	payload := `{"actions":[{"value":"deny:appr-2"}]}`
	form := url.Values{}
	form.Set("payload", payload)
	body := []byte(form.Encode())

	approver := &testApprover{}
	h := &InteractionHandler{Approver: approver}

	req := httptest.NewRequest(http.MethodPost, "/v1/slack/interactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res := httptest.NewRecorder()
	h.HandleInteractions(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	if approver.approvalID != "appr-2" || approver.status != "denied" {
		t.Fatalf("unexpected approval %s %s", approver.approvalID, approver.status)
	}
}

func TestHandleInteractionsMissingPayload(t *testing.T) {
	approver := &testApprover{}
	h := &InteractionHandler{Approver: approver}

	req := httptest.NewRequest(http.MethodPost, "/v1/slack/interactions", bytes.NewBufferString("payload="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res := httptest.NewRecorder()
	h.HandleInteractions(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestHandleInteractionsInvalidSignature(t *testing.T) {
	payload := `{"actions":[{"action_id":"approve","value":"appr-1"}]}`
	form := url.Values{}
	form.Set("payload", payload)
	body := []byte(form.Encode())

	approver := &testApprover{}
	h := &InteractionHandler{SigningSecret: "secret", Approver: approver, Now: func() time.Time { return time.Now() }}

	req := httptest.NewRequest(http.MethodPost, "/v1/slack/interactions", bytes.NewBuffer(body))
	req.Header.Set("X-Slack-Signature", "v0=bad")
	req.Header.Set("X-Slack-Request-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	res := httptest.NewRecorder()
	h.HandleInteractions(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}
}
