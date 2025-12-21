package api

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davidahmann/relia/internal/aws"
	"github.com/davidahmann/relia/internal/ledger"
)

type flipBroker struct {
	calls int
}

func (b *flipBroker) AssumeRoleWithWebIdentity(input aws.AssumeRoleInput) (aws.Credentials, error) {
	b.calls++
	if b.calls == 1 {
		return aws.Credentials{}, fmt.Errorf("temporary sts failure")
	}
	return aws.Credentials{
		AccessKeyID:     "AKIA_TEST",
		SecretAccessKey: "SECRET_TEST",
		SessionToken:    "TOKEN_TEST",
		ExpiresAt:       time.Now().UTC().Add(15 * time.Minute),
	}, nil
}

func TestAuthorizeRetriesWhenIssuing(t *testing.T) {
	dir := t.TempDir()
	policyPath := filepath.Join(dir, "policy.yaml")

	policyYAML := `
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
	if err := os.WriteFile(policyPath, []byte(policyYAML), 0o600); err != nil {
		t.Fatalf("write policy: %v", err)
	}

	store := ledger.NewInMemoryStore()
	broker := &flipBroker{}
	service, err := NewAuthorizeService(NewAuthorizeServiceInput{
		PolicyPath: policyPath,
		Ledger:     store,
		Broker:     broker,
	})
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-dev",
		RunID:    "123",
		SHA:      "abc",
		Token:    "jwt",
	}
	req := AuthorizeRequest{Action: "terraform.apply", Resource: "res", Env: "dev", RequestID: "req-1"}

	_, err = service.Authorize(claims, req, "2025-12-21T00:00:00Z")
	if err == nil {
		t.Fatalf("expected first issuance error")
	}

	resp, err := service.Authorize(claims, req, "2025-12-21T00:00:05Z")
	if err != nil {
		t.Fatalf("expected retry success, got: %v", err)
	}
	if resp.Verdict != string(VerdictAllow) {
		t.Fatalf("expected allow, got %s", resp.Verdict)
	}
	if resp.ReceiptID == "" {
		t.Fatalf("missing receipt_id")
	}
	if resp.AWSCredentials == nil || resp.AWSCredentials.AccessKeyID == "" {
		t.Fatalf("missing aws credentials")
	}
}
