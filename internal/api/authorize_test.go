package api

import (
	"testing"

	"github.com/davidahmann/relia_oss/internal/crypto"
)

func TestComputeIdemKeyDeterministic(t *testing.T) {
	actor := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "https://token.actions.githubusercontent.com",
		Repo:     "org/repo",
		Workflow: "terraform-prod",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	req := AuthorizeRequest{
		Action:   "terraform.apply",
		Resource: "aws:account:123456789012:stack/prod",
		Env:      "prod",
		Intent: map[string]any{
			"change_id": "CHG-1234",
			"ticket":    nil,
		},
		Evidence: AuthorizeEvidence{
			PlanDigest: "sha256:plan",
		},
	}

	keyA, err := ComputeIdemKey(actor, req)
	if err != nil {
		t.Fatalf("compute idem: %v", err)
	}
	keyB, err := ComputeIdemKey(actor, req)
	if err != nil {
		t.Fatalf("compute idem: %v", err)
	}

	if keyA == "" {
		t.Fatalf("expected idem key")
	}
	if keyA != keyB {
		t.Fatalf("expected deterministic idem key")
	}
}

func TestComputeIdemKeyMissingFields(t *testing.T) {
	actor := ActorContext{Subject: "sub", Issuer: "iss", Repo: "org/repo", RunID: "1"}
	_, err := ComputeIdemKey(actor, AuthorizeRequest{Action: "", Resource: "r", Env: "e"})
	if err == nil {
		t.Fatalf("expected error for missing action")
	}
}

func TestComputeIdemKeyRejectsFloat(t *testing.T) {
	actor := ActorContext{
		Subject: "sub",
		Issuer:  "iss",
		Repo:    "org/repo",
		RunID:   "1",
	}

	req := AuthorizeRequest{
		Action:   "action",
		Resource: "res",
		Env:      "prod",
		Intent: map[string]any{
			"score": 0.5,
		},
	}

	_, err := ComputeIdemKey(actor, req)
	if err != crypto.ErrFloatNotAllowed {
		t.Fatalf("expected ErrFloatNotAllowed, got %v", err)
	}
}
