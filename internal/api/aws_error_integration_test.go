package api

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidahmann/relia/internal/aws"
)

type errorBroker struct{}

func (e errorBroker) AssumeRoleWithWebIdentity(aws.AssumeRoleInput) (aws.Credentials, error) {
	return aws.Credentials{}, fmt.Errorf("broker error")
}

func TestAuthorizeBrokerError(t *testing.T) {
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

	service, err := NewAuthorizeService(policyPath)
	if err != nil {
		t.Fatalf("service: %v", err)
	}
	service.Broker = errorBroker{}

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-prod",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	_, err = service.Authorize(claims, AuthorizeRequest{Action: "terraform.apply", Resource: "res", Env: "dev", RequestID: "req-1"}, "2025-12-20T16:34:14Z")
	if err == nil {
		t.Fatalf("expected error")
	}
}
