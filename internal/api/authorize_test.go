package api

import (
	"encoding/json"
	"testing"

	"github.com/davidahmann/relia/internal/crypto"
	"github.com/davidahmann/relia/pkg/types"
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

func TestComputeIdemKeyIncludesInteractionRef(t *testing.T) {
	actor := ActorContext{
		Subject: "sub",
		Issuer:  "iss",
		Repo:    "org/repo",
		RunID:   "1",
	}

	req := AuthorizeRequest{
		Action:         "action",
		Resource:       "res",
		Env:            "prod",
		InteractionRef: &types.InteractionRef{Mode: "voice", CallID: "call-1", TurnID: "turn-1", TurnIndex: 1},
		Intent:         map[string]any{"x": "y"},
		Evidence:       AuthorizeEvidence{PlanDigest: "sha256:plan"},
		ContextRef:     &types.ContextRef{ContextID: "ctx-1", RecordHash: "sha256:ctx"},
		DecisionRef:    &types.DecisionRef{DecisionID: "dec-1", InputsDigest: "sha256:inputs"},
	}

	keyA, err := ComputeIdemKey(actor, req)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}

	req.InteractionRef.CallID = "call-2"
	keyB, err := ComputeIdemKey(actor, req)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if keyA == keyB {
		t.Fatalf("expected different idem keys when interaction_ref changes")
	}
}

func TestInteractionRefFromBody(t *testing.T) {
	body, err := json.Marshal(map[string]any{
		"interaction_ref": &types.InteractionRef{Mode: "voice", CallID: "call-1", TurnID: "turn-1", TurnIndex: 1},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := interactionRefFromBody(body)
	if got == nil || got.CallID != "call-1" || got.TurnIndex != 1 {
		t.Fatalf("unexpected interaction ref: %+v", got)
	}

	if got := interactionRefFromBody([]byte("not-json")); got != nil {
		t.Fatalf("expected nil for invalid json")
	}
}
