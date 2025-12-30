package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidahmann/relia/internal/auth"
	"github.com/davidahmann/relia/pkg/types"
)

func TestAuthorizePreservesRefsInReceiptAndPack(t *testing.T) {
	t.Setenv("RELIA_DEV_TOKEN", "test-token")

	dir := t.TempDir()
	policyPath := filepath.Join(dir, "policy.yaml")
	policy := `
policy_id: test
policy_version: "1"
defaults:
  ttl_seconds: 900
  require_approval: false
  deny: false
rules:
  - id: allow-dev
    match:
      action: "terraform.apply"
      env: "dev"
    effect:
      ttl_seconds: 900
      aws_role_arn: "arn:aws:iam::123456789012:role/test"
`
	if err := os.WriteFile(policyPath, []byte(policy), 0o600); err != nil {
		t.Fatalf("write policy: %v", err)
	}

	service, err := NewAuthorizeService(NewAuthorizeServiceInput{PolicyPath: policyPath})
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	router := NewRouter(&Handler{
		Auth:             auth.NewAuthenticatorFromEnv(),
		AuthorizeService: service,
	})
	srv := httptest.NewServer(router)
	defer srv.Close()

	reqBody := `{
  "action":"terraform.apply",
  "resource":"res",
  "env":"dev",
  "interaction_ref":{"mode":"voice","call_id":"call-1","turn_id":"turn-1","turn_index":1,"consent_state":"recording_ok"},
  "context_ref":{"context_id":"context-1","record_hash":"sha256:ctxrecord"},
  "decision_ref":{"decision_id":"decision-1","inputs_digest":"sha256:decinputs"}
}`

	receiptID := authorizeJSON(t, srv.URL, reqBody)

	verifyResp := verifyJSON(t, srv.URL, receiptID)
	receipt := verifyResp["receipt"].(map[string]any)
	interaction := receipt["interaction_ref"].(map[string]any)
	if interaction["mode"] != "voice" {
		t.Fatalf("unexpected mode: %v", interaction["mode"])
	}
	if interaction["call_id"] != "call-1" {
		t.Fatalf("unexpected call_id: %v", interaction["call_id"])
	}
	if interaction["turn_id"] != "turn-1" {
		t.Fatalf("unexpected turn_id: %v", interaction["turn_id"])
	}
	refs := receipt["refs"].(map[string]any)

	ctx := refs["context"].(map[string]any)
	if ctx["context_id"] != "context-1" {
		t.Fatalf("unexpected context_id: %v", ctx["context_id"])
	}
	if ctx["record_hash"] != "sha256:ctxrecord" {
		t.Fatalf("unexpected record_hash: %v", ctx["record_hash"])
	}

	dec := refs["decision"].(map[string]any)
	if dec["decision_id"] != "decision-1" {
		t.Fatalf("unexpected decision_id: %v", dec["decision_id"])
	}
	if dec["inputs_digest"] != "sha256:decinputs" {
		t.Fatalf("unexpected inputs_digest: %v", dec["inputs_digest"])
	}

	packBytes := packZIP(t, srv.URL, receiptID)
	manifestBytes := zipReadFile(t, packBytes, "manifest.json")
	var manifest types.PackManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	if manifest.Refs == nil || manifest.Refs.Context == nil || manifest.Refs.Decision == nil {
		t.Fatalf("expected refs in manifest")
	}
	if manifest.Refs.Context.ContextID != "context-1" || manifest.Refs.Context.RecordHash != "sha256:ctxrecord" {
		t.Fatalf("manifest context refs mismatch: %+v", manifest.Refs.Context)
	}
	if manifest.Refs.Decision.DecisionID != "decision-1" || manifest.Refs.Decision.InputsDigest != "sha256:decinputs" {
		t.Fatalf("manifest decision refs mismatch: %+v", manifest.Refs.Decision)
	}
	if manifest.InteractionRef == nil || manifest.InteractionRef.Mode != "voice" || manifest.InteractionRef.CallID != "call-1" || manifest.InteractionRef.TurnID != "turn-1" {
		t.Fatalf("manifest interaction_ref mismatch: %+v", manifest.InteractionRef)
	}
}

func authorizeJSON(t *testing.T, baseURL string, body string) string {
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
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("authorize status: %d body=%s", res.StatusCode, string(b))
	}

	var payload struct {
		ReceiptID string `json:"receipt_id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.ReceiptID == "" {
		t.Fatalf("missing receipt_id")
	}
	return payload.ReceiptID
}

func verifyJSON(t *testing.T, baseURL string, receiptID string) map[string]any {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, baseURL+"/v1/verify/"+receiptID, nil)
	if err != nil {
		t.Fatalf("new req: %v", err)
	}
	req.Header.Set("Authorization", "Bearer test-token")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("verify status: %d body=%s", res.StatusCode, string(b))
	}

	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if ok, _ := payload["valid"].(bool); !ok {
		t.Fatalf("expected valid receipt")
	}
	return payload
}

func packZIP(t *testing.T, baseURL string, receiptID string) []byte {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, baseURL+"/v1/pack/"+receiptID, nil)
	if err != nil {
		t.Fatalf("new req: %v", err)
	}
	req.Header.Set("Authorization", "Bearer test-token")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("pack status: %d body=%s", res.StatusCode, string(b))
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return b
}

func zipReadFile(t *testing.T, zipBytes []byte, name string) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip: %v", err)
	}
	for _, f := range reader.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		defer rc.Close()
		b, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		return b
	}
	t.Fatalf("missing %s in zip", name)
	return nil
}
