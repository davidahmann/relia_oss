package ledger

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"testing"

	"github.com/davidahmann/relia/internal/crypto"
	"github.com/davidahmann/relia/pkg/types"
)

type testSigner struct {
	keyID string
	priv  ed25519.PrivateKey
}

func (s testSigner) KeyID() string {
	return s.keyID
}

func (s testSigner) SignEd25519(message []byte) ([]byte, error) {
	return crypto.SignEd25519(s.priv, message)
}

func TestMakeReceiptAndVerify(t *testing.T) {
	seed := bytes.Repeat([]byte{0x01}, 32)
	priv, pub, err := crypto.KeyPairFromSeed(seed)
	if err != nil {
		t.Fatalf("keypair: %v", err)
	}

	signer := testSigner{keyID: "test-key", priv: priv}

	input := MakeReceiptInput{
		Schema:     ReceiptSchema,
		CreatedAt:  "2025-12-20T16:34:14Z",
		IdemKey:    "idem:v1:sha256:abc",
		ContextID:  "sha256:ctx",
		DecisionID: "sha256:dec",
		Actor: types.ReceiptActor{
			Kind:     "workload",
			Subject:  "repo:org/repo:ref:refs/heads/main",
			Issuer:   "https://token.actions.githubusercontent.com",
			Repo:     "org/repo",
			Workflow: "wf",
			RunID:    "123",
			SHA:      "abc",
		},
		Request: types.ReceiptRequest{
			RequestID: "01JTEST",
			Action:    "deploy",
			Resource:  "resource",
			Env:       "prod",
			Intent: map[string]any{
				"change_id": "CHG-1",
			},
		},
		Policy: types.ReceiptPolicy{
			PolicyID:      "relia-default",
			PolicyVersion: "2025-12-20",
			PolicyHash:    "sha256:policy",
		},
		Outcome: types.ReceiptOutcome{
			Status: types.OutcomeDenied,
		},
	}

	receipt, err := MakeReceipt(input, signer)
	if err != nil {
		t.Fatalf("make receipt: %v", err)
	}

	if receipt.ReceiptID == "" || receipt.BodyDigest == "" {
		t.Fatalf("missing digest")
	}
	if receipt.ReceiptID != receipt.BodyDigest {
		t.Fatalf("receipt id should equal body digest")
	}

	if err := VerifyReceipt(receipt, pub); err != nil {
		t.Fatalf("verify receipt: %v", err)
	}
}

func TestMakeReceiptIncludesRefs(t *testing.T) {
	seed := bytes.Repeat([]byte{0x01}, 32)
	priv, pub, err := crypto.KeyPairFromSeed(seed)
	if err != nil {
		t.Fatalf("keypair: %v", err)
	}

	signer := testSigner{keyID: "test-key", priv: priv}

	input := MakeReceiptInput{
		Schema:     ReceiptSchema,
		CreatedAt:  "2025-12-20T16:34:14Z",
		IdemKey:    "idem:v1:sha256:abc",
		ContextID:  "sha256:ctx",
		DecisionID: "sha256:dec",
		Actor:      types.ReceiptActor{Kind: "workload", Subject: "dev"},
		Request:    types.ReceiptRequest{Action: "deploy", Resource: "resource", Env: "prod"},
		Policy:     types.ReceiptPolicy{PolicyHash: "sha256:policy"},
		InteractionRef: &types.InteractionRef{
			Mode:         "voice",
			CallID:       "call-1",
			TurnID:       "turn-1",
			TurnIndex:    1,
			ConsentState: "recording_ok",
		},
		Refs: &types.ReceiptRefs{
			Context:  &types.ContextRef{ContextID: "context-1", RecordHash: "sha256:ctxrecord"},
			Decision: &types.DecisionRef{DecisionID: "decision-1", InputsDigest: "sha256:decinputs"},
		},
		Outcome: types.ReceiptOutcome{Status: types.OutcomeDenied},
	}

	receipt, err := MakeReceipt(input, signer)
	if err != nil {
		t.Fatalf("make receipt: %v", err)
	}
	if err := VerifyReceipt(receipt, pub); err != nil {
		t.Fatalf("verify receipt: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(receipt.BodyJSON, &body); err != nil {
		t.Fatalf("unmarshal receipt: %v", err)
	}
	interaction, ok := body["interaction_ref"].(map[string]any)
	if !ok {
		t.Fatalf("expected interaction_ref object")
	}
	if interaction["mode"] != "voice" {
		t.Fatalf("unexpected interaction_ref.mode: %v", interaction["mode"])
	}
	refs, ok := body["refs"].(map[string]any)
	if !ok {
		t.Fatalf("expected refs object")
	}
	ctx, ok := refs["context"].(map[string]any)
	if !ok || ctx["record_hash"] != "sha256:ctxrecord" {
		t.Fatalf("unexpected context refs: %+v", ctx)
	}
	dec, ok := refs["decision"].(map[string]any)
	if !ok || dec["inputs_digest"] != "sha256:decinputs" {
		t.Fatalf("unexpected decision refs: %+v", dec)
	}
}

func TestMakeReceiptRejectsOutcome(t *testing.T) {
	seed := bytes.Repeat([]byte{0x01}, 32)
	priv, _, err := crypto.KeyPairFromSeed(seed)
	if err != nil {
		t.Fatalf("keypair: %v", err)
	}

	signer := testSigner{keyID: "test-key", priv: priv}

	input := MakeReceiptInput{
		Schema:     ReceiptSchema,
		CreatedAt:  "2025-12-20T16:34:14Z",
		IdemKey:    "idem:v1:sha256:abc",
		ContextID:  "sha256:ctx",
		DecisionID: "sha256:dec",
		Actor:      types.ReceiptActor{Kind: "workload"},
		Request:    types.ReceiptRequest{Action: "deploy", Resource: "res", Env: "prod"},
		Policy:     types.ReceiptPolicy{PolicyHash: "sha256:policy"},
		Outcome:    types.ReceiptOutcome{Status: types.OutcomeStatus("invalid")},
	}

	_, err = MakeReceipt(input, signer)
	if err == nil {
		t.Fatalf("expected error for invalid outcome")
	}
}

func TestVerifyReceiptDigestMismatch(t *testing.T) {
	seed := bytes.Repeat([]byte{0x01}, 32)
	priv, pub, err := crypto.KeyPairFromSeed(seed)
	if err != nil {
		t.Fatalf("keypair: %v", err)
	}

	signer := testSigner{keyID: "test-key", priv: priv}

	input := MakeReceiptInput{
		Schema:     ReceiptSchema,
		CreatedAt:  "2025-12-20T16:34:14Z",
		IdemKey:    "idem:v1:sha256:abc",
		ContextID:  "sha256:ctx",
		DecisionID: "sha256:dec",
		Actor:      types.ReceiptActor{Kind: "workload"},
		Request:    types.ReceiptRequest{Action: "deploy", Resource: "res", Env: "prod"},
		Policy:     types.ReceiptPolicy{PolicyHash: "sha256:policy"},
		Outcome:    types.ReceiptOutcome{Status: types.OutcomeDenied},
	}

	receipt, err := MakeReceipt(input, signer)
	if err != nil {
		t.Fatalf("make receipt: %v", err)
	}

	receipt.BodyDigest = "sha256:tampered"
	if err := VerifyReceipt(receipt, pub); err != ErrReceiptDigestMismatch {
		t.Fatalf("expected ErrReceiptDigestMismatch, got %v", err)
	}
}

func TestVerifyReceiptSignatureInvalid(t *testing.T) {
	seed := bytes.Repeat([]byte{0x01}, 32)
	priv, pub, err := crypto.KeyPairFromSeed(seed)
	if err != nil {
		t.Fatalf("keypair: %v", err)
	}

	signer := testSigner{keyID: "test-key", priv: priv}

	input := MakeReceiptInput{
		Schema:     ReceiptSchema,
		CreatedAt:  "2025-12-20T16:34:14Z",
		IdemKey:    "idem:v1:sha256:abc",
		ContextID:  "sha256:ctx",
		DecisionID: "sha256:dec",
		Actor:      types.ReceiptActor{Kind: "workload"},
		Request:    types.ReceiptRequest{Action: "deploy", Resource: "res", Env: "prod"},
		Policy:     types.ReceiptPolicy{PolicyHash: "sha256:policy"},
		Outcome:    types.ReceiptOutcome{Status: types.OutcomeDenied},
	}

	receipt, err := MakeReceipt(input, signer)
	if err != nil {
		t.Fatalf("make receipt: %v", err)
	}

	receipt.Sig[0] ^= 0xff
	if err := VerifyReceipt(receipt, pub); err != ErrReceiptSignature {
		t.Fatalf("expected ErrReceiptSignature, got %v", err)
	}
}

func TestMakeReceiptFinalAndApproval(t *testing.T) {
	seed := bytes.Repeat([]byte{0x01}, 32)
	priv, _, err := crypto.KeyPairFromSeed(seed)
	if err != nil {
		t.Fatalf("keypair: %v", err)
	}

	signer := testSigner{keyID: "test-key", priv: priv}

	approval := &types.ReceiptApproval{
		Required:   true,
		ApprovalID: "appr-1",
		Status:     "approved",
		ApprovedAt: "2025-12-20T16:35:02Z",
	}

	outcome := types.ReceiptOutcome{
		Status:    types.OutcomeIssuedCredentials,
		IssuedAt:  "2025-12-20T16:35:03Z",
		ExpiresAt: "2025-12-20T16:50:03Z",
	}

	input := MakeReceiptInput{
		CreatedAt:  "2025-12-20T16:34:14Z",
		IdemKey:    "idem:v1:sha256:abc",
		ContextID:  "sha256:ctx",
		DecisionID: "sha256:dec",
		Actor:      types.ReceiptActor{Kind: "workload"},
		Request:    types.ReceiptRequest{Action: "deploy", Resource: "res", Env: "prod"},
		Policy:     types.ReceiptPolicy{PolicyHash: "sha256:policy"},
		Approval:   approval,
		Outcome:    outcome,
	}

	receipt, err := MakeReceipt(input, signer)
	if err != nil {
		t.Fatalf("make receipt: %v", err)
	}

	if !receipt.Final {
		t.Fatalf("expected final receipt")
	}
	if receipt.ApprovalID == nil || *receipt.ApprovalID != "appr-1" {
		t.Fatalf("expected approval id to be set")
	}
	if receipt.ExpiresAt == nil || *receipt.ExpiresAt != outcome.ExpiresAt {
		t.Fatalf("expected expires_at to be set")
	}
}

func TestMakeReceiptInvalidSchema(t *testing.T) {
	seed := bytes.Repeat([]byte{0x01}, 32)
	priv, _, err := crypto.KeyPairFromSeed(seed)
	if err != nil {
		t.Fatalf("keypair: %v", err)
	}

	signer := testSigner{keyID: "test-key", priv: priv}

	input := MakeReceiptInput{
		Schema:     "bad.schema",
		CreatedAt:  "2025-12-20T16:34:14Z",
		IdemKey:    "idem:v1:sha256:abc",
		ContextID:  "sha256:ctx",
		DecisionID: "sha256:dec",
		Actor:      types.ReceiptActor{Kind: "workload"},
		Request:    types.ReceiptRequest{Action: "deploy", Resource: "res", Env: "prod"},
		Policy:     types.ReceiptPolicy{PolicyHash: "sha256:policy"},
		Outcome:    types.ReceiptOutcome{Status: types.OutcomeDenied},
	}

	_, err = MakeReceipt(input, signer)
	if err == nil {
		t.Fatalf("expected error for invalid schema")
	}
}
