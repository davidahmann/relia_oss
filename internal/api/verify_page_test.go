package api

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davidahmann/relia/internal/aws"
	"github.com/davidahmann/relia/internal/context"
	"github.com/davidahmann/relia/internal/decision"
	"github.com/davidahmann/relia/internal/ledger"
	"github.com/davidahmann/relia/pkg/types"
)

type fixedSigner struct {
	keyID string
	priv  ed25519.PrivateKey
}

func (s fixedSigner) KeyID() string { return s.keyID }
func (s fixedSigner) SignEd25519(message []byte) ([]byte, error) {
	return ed25519.Sign(s.priv, message), nil
}

func TestVerifyPageAndPublicPack(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)

	store := ledger.NewInMemoryStore()
	createdAt := time.Now().UTC().Format(time.RFC3339)

	policyHash := "sha256:policy"
	if err := store.PutPolicyVersion(ledger.PolicyVersionRecord{
		PolicyHash:    policyHash,
		PolicyID:      "relia-default",
		PolicyVersion: "2025-12-20",
		PolicyYAML:    "policy_id: relia-default\n",
		CreatedAt:     createdAt,
	}); err != nil {
		t.Fatalf("put policy: %v", err)
	}

	ctx, err := context.BuildContext(
		types.ContextSource{Kind: "github_actions", Repo: "org/repo", Workflow: "wf", RunID: "1", Actor: "dev", SHA: "abc"},
		types.ContextInputs{Action: "terraform.apply", Resource: "stack/prod", Env: "prod"},
		types.ContextEvidence{PlanDigest: "sha256:plan", DiffURL: "https://github.com/org/repo/actions/runs/1"},
		createdAt,
	)
	if err != nil {
		t.Fatalf("context: %v", err)
	}
	ctxJSON, _ := json.Marshal(ctx)
	if err := store.PutContext(ledger.ContextRecord{ContextID: ctx.ContextID, BodyJSON: ctxJSON, CreatedAt: createdAt}); err != nil {
		t.Fatalf("put context: %v", err)
	}

	dec, err := decision.BuildDecision(
		ctx.ContextID,
		types.DecisionPolicy{PolicyID: "relia-default", PolicyVersion: "2025-12-20", PolicyHash: policyHash},
		"allow",
		nil,
		false,
		"high",
		createdAt,
	)
	if err != nil {
		t.Fatalf("decision: %v", err)
	}
	decJSON, _ := json.Marshal(dec)
	if err := store.PutDecision(ledger.DecisionRecord{DecisionID: dec.DecisionID, ContextID: ctx.ContextID, PolicyHash: policyHash, Verdict: dec.Verdict, BodyJSON: decJSON, CreatedAt: createdAt}); err != nil {
		t.Fatalf("put decision: %v", err)
	}

	rec, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
		CreatedAt:      createdAt,
		IdemKey:        "idem",
		ContextID:      ctx.ContextID,
		DecisionID:     dec.DecisionID,
		Actor:          types.ReceiptActor{Kind: "workload", Subject: "dev", Issuer: "relia-dev", Repo: "org/repo", Workflow: "wf", RunID: "1", SHA: "abc"},
		Request:        types.ReceiptRequest{Action: "terraform.apply", Resource: "stack/prod", Env: "prod"},
		Policy:         types.ReceiptPolicy{PolicyID: "relia-default", PolicyVersion: "2025-12-20", PolicyHash: policyHash},
		InteractionRef: &types.InteractionRef{Mode: "voice", CallID: "call-1", TurnID: "turn-1", TurnIndex: 1, ConsentState: "recording_ok"},
		CredentialGrant: &types.ReceiptCredentialGrant{
			Provider:   "aws_sts",
			Method:     "AssumeRoleWithWebIdentity",
			RoleARN:    "arn:aws:iam::123456789012:role/test",
			TTLSeconds: 900,
		},
		Outcome: types.ReceiptOutcome{Status: types.OutcomeIssuedCredentials, ExpiresAt: createdAt},
	}, fixedSigner{keyID: "k1", priv: priv})
	if err != nil {
		t.Fatalf("receipt: %v", err)
	}
	if err := store.PutReceipt(receiptRecordFromStored(rec)); err != nil {
		t.Fatalf("put receipt: %v", err)
	}

	svc, err := NewAuthorizeService(NewAuthorizeServiceInput{
		PolicyPath: "policies/relia.yaml",
		Ledger:     store,
		Signer:     fixedSigner{keyID: "k1", priv: priv},
		PublicKey:  pub,
		Broker:     aws.DevBroker{},
	})
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	h := &Handler{AuthorizeService: svc, PublicVerify: true}
	router := NewRouter(h)

	r1 := httptest.NewRequest(http.MethodGet, "/verify/"+rec.ReceiptID, nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("verify page status=%d body=%s", w1.Code, w1.Body.String())
	}
	if !bytes.Contains(w1.Body.Bytes(), []byte("Relia Verify")) {
		t.Fatalf("expected html body")
	}
	if !bytes.Contains(w1.Body.Bytes(), []byte("call_id=call-1")) {
		t.Fatalf("expected interaction_ref to be rendered, got: %s", w1.Body.String())
	}

	r2 := httptest.NewRequest(http.MethodGet, "/pack/"+rec.ReceiptID, nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("pack status=%d body=%s", w2.Code, w2.Body.String())
	}
	if ct := w2.Header().Get("Content-Type"); ct != "application/zip" {
		t.Fatalf("expected zip content-type, got %q", ct)
	}
}

func TestVerifyPageDisabledReturns404(t *testing.T) {
	router := NewRouter(&Handler{PublicVerify: false})
	req := httptest.NewRequest(http.MethodGet, "/verify/anything", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestVerifyPageInvalidSignatureStillRenders(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)

	store := ledger.NewInMemoryStore()
	createdAt := time.Now().UTC().Format(time.RFC3339)

	policyHash := "sha256:policy"
	if err := store.PutPolicyVersion(ledger.PolicyVersionRecord{
		PolicyHash:    policyHash,
		PolicyID:      "relia-default",
		PolicyVersion: "2025-12-20",
		PolicyYAML:    "policy_id: relia-default\n",
		CreatedAt:     createdAt,
	}); err != nil {
		t.Fatalf("put policy: %v", err)
	}

	ctx, err := context.BuildContext(
		types.ContextSource{Kind: "github_actions", Repo: "org/repo", Workflow: "wf", RunID: "1", Actor: "dev", SHA: "abc"},
		types.ContextInputs{Action: "terraform.apply", Resource: "stack/prod", Env: "prod"},
		types.ContextEvidence{},
		createdAt,
	)
	if err != nil {
		t.Fatalf("context: %v", err)
	}
	ctxJSON, _ := json.Marshal(ctx)
	if err := store.PutContext(ledger.ContextRecord{ContextID: ctx.ContextID, BodyJSON: ctxJSON, CreatedAt: createdAt}); err != nil {
		t.Fatalf("put context: %v", err)
	}

	dec, err := decision.BuildDecision(
		ctx.ContextID,
		types.DecisionPolicy{PolicyID: "relia-default", PolicyVersion: "2025-12-20", PolicyHash: policyHash},
		"allow",
		nil,
		false,
		"high",
		createdAt,
	)
	if err != nil {
		t.Fatalf("decision: %v", err)
	}
	decJSON, _ := json.Marshal(dec)
	if err := store.PutDecision(ledger.DecisionRecord{DecisionID: dec.DecisionID, ContextID: ctx.ContextID, PolicyHash: policyHash, Verdict: dec.Verdict, BodyJSON: decJSON, CreatedAt: createdAt}); err != nil {
		t.Fatalf("put decision: %v", err)
	}

	rec, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
		CreatedAt:  createdAt,
		IdemKey:    "idem",
		ContextID:  ctx.ContextID,
		DecisionID: dec.DecisionID,
		Actor:      types.ReceiptActor{Kind: "workload", Subject: "dev", Issuer: "relia-dev", Repo: "org/repo", Workflow: "wf", RunID: "1", SHA: "abc"},
		Request:    types.ReceiptRequest{Action: "terraform.apply", Resource: "stack/prod", Env: "prod"},
		Policy:     types.ReceiptPolicy{PolicyID: "relia-default", PolicyVersion: "2025-12-20", PolicyHash: policyHash},
		Outcome:    types.ReceiptOutcome{Status: types.OutcomeDenied},
	}, fixedSigner{keyID: "k1", priv: priv})
	if err != nil {
		t.Fatalf("receipt: %v", err)
	}
	rr := receiptRecordFromStored(rec)
	rr.Sig = []byte("bad")
	if err := store.PutReceipt(rr); err != nil {
		t.Fatalf("put receipt: %v", err)
	}

	svc, err := NewAuthorizeService(NewAuthorizeServiceInput{
		PolicyPath: "policies/relia.yaml",
		Ledger:     store,
		Signer:     fixedSigner{keyID: "k1", priv: priv},
		PublicKey:  pub,
		Broker:     aws.DevBroker{},
	})
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	router := NewRouter(&Handler{AuthorizeService: svc, PublicVerify: true})
	req := httptest.NewRequest(http.MethodGet, "/verify/"+rec.ReceiptID, nil)
	rrr := httptest.NewRecorder()
	router.ServeHTTP(rrr, req)
	if rrr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrr.Code)
	}
	if !bytes.Contains(rrr.Body.Bytes(), []byte("INVALID")) {
		t.Fatalf("expected INVALID marker, got: %s", rrr.Body.String())
	}
}

func TestPublicPackMissingPolicyVersionIs404(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)

	store := ledger.NewInMemoryStore()
	createdAt := time.Now().UTC().Format(time.RFC3339)

	// Intentionally do NOT put policy version record.
	policyHash := "sha256:policy"

	ctx, err := context.BuildContext(
		types.ContextSource{Kind: "github_actions", Repo: "org/repo", Workflow: "wf", RunID: "1", Actor: "dev", SHA: "abc"},
		types.ContextInputs{Action: "terraform.apply", Resource: "stack/prod", Env: "prod"},
		types.ContextEvidence{},
		createdAt,
	)
	if err != nil {
		t.Fatalf("context: %v", err)
	}
	ctxJSON, _ := json.Marshal(ctx)
	if err := store.PutContext(ledger.ContextRecord{ContextID: ctx.ContextID, BodyJSON: ctxJSON, CreatedAt: createdAt}); err != nil {
		t.Fatalf("put context: %v", err)
	}

	dec, err := decision.BuildDecision(
		ctx.ContextID,
		types.DecisionPolicy{PolicyID: "relia-default", PolicyVersion: "2025-12-20", PolicyHash: policyHash},
		"allow",
		nil,
		false,
		"high",
		createdAt,
	)
	if err != nil {
		t.Fatalf("decision: %v", err)
	}
	decJSON, _ := json.Marshal(dec)
	if err := store.PutDecision(ledger.DecisionRecord{DecisionID: dec.DecisionID, ContextID: ctx.ContextID, PolicyHash: policyHash, Verdict: dec.Verdict, BodyJSON: decJSON, CreatedAt: createdAt}); err != nil {
		t.Fatalf("put decision: %v", err)
	}

	rec, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
		CreatedAt:  createdAt,
		IdemKey:    "idem",
		ContextID:  ctx.ContextID,
		DecisionID: dec.DecisionID,
		Actor:      types.ReceiptActor{Kind: "workload", Subject: "dev", Issuer: "relia-dev", Repo: "org/repo", Workflow: "wf", RunID: "1", SHA: "abc"},
		Request:    types.ReceiptRequest{Action: "terraform.apply", Resource: "stack/prod", Env: "prod"},
		Policy:     types.ReceiptPolicy{PolicyID: "relia-default", PolicyVersion: "2025-12-20", PolicyHash: policyHash},
		Outcome:    types.ReceiptOutcome{Status: types.OutcomeDenied},
	}, fixedSigner{keyID: "k1", priv: priv})
	if err != nil {
		t.Fatalf("receipt: %v", err)
	}
	if err := store.PutReceipt(receiptRecordFromStored(rec)); err != nil {
		t.Fatalf("put receipt: %v", err)
	}

	svc, err := NewAuthorizeService(NewAuthorizeServiceInput{
		PolicyPath: "policies/relia.yaml",
		Ledger:     store,
		Signer:     fixedSigner{keyID: "k1", priv: priv},
		PublicKey:  pub,
		Broker:     aws.DevBroker{},
	})
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	router := NewRouter(&Handler{AuthorizeService: svc, PublicVerify: true})
	req := httptest.NewRequest(http.MethodGet, "/pack/"+rec.ReceiptID, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestPublicPackInvalidContextIs500(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)

	store := ledger.NewInMemoryStore()
	createdAt := time.Now().UTC().Format(time.RFC3339)
	policyHash := "sha256:policy"
	if err := store.PutPolicyVersion(ledger.PolicyVersionRecord{
		PolicyHash:    policyHash,
		PolicyID:      "relia-default",
		PolicyVersion: "2025-12-20",
		PolicyYAML:    "policy_id: relia-default\n",
		CreatedAt:     createdAt,
	}); err != nil {
		t.Fatalf("put policy: %v", err)
	}

	if err := store.PutContext(ledger.ContextRecord{ContextID: "c1", BodyJSON: []byte("not-json"), CreatedAt: createdAt}); err != nil {
		t.Fatalf("put context: %v", err)
	}

	dec, err := decision.BuildDecision(
		"c1",
		types.DecisionPolicy{PolicyID: "relia-default", PolicyVersion: "2025-12-20", PolicyHash: policyHash},
		"allow",
		nil,
		false,
		"high",
		createdAt,
	)
	if err != nil {
		t.Fatalf("decision: %v", err)
	}
	decJSON, _ := json.Marshal(dec)
	if err := store.PutDecision(ledger.DecisionRecord{DecisionID: dec.DecisionID, ContextID: "c1", PolicyHash: policyHash, Verdict: dec.Verdict, BodyJSON: decJSON, CreatedAt: createdAt}); err != nil {
		t.Fatalf("put decision: %v", err)
	}

	rec, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
		CreatedAt:  createdAt,
		IdemKey:    "idem",
		ContextID:  "c1",
		DecisionID: dec.DecisionID,
		Actor:      types.ReceiptActor{Kind: "workload", Subject: "dev"},
		Request:    types.ReceiptRequest{Action: "terraform.apply", Resource: "stack/prod", Env: "prod"},
		Policy:     types.ReceiptPolicy{PolicyID: "relia-default", PolicyVersion: "2025-12-20", PolicyHash: policyHash},
		Outcome:    types.ReceiptOutcome{Status: types.OutcomeDenied},
	}, fixedSigner{keyID: "k1", priv: priv})
	if err != nil {
		t.Fatalf("receipt: %v", err)
	}
	if err := store.PutReceipt(receiptRecordFromStored(rec)); err != nil {
		t.Fatalf("put receipt: %v", err)
	}

	svc, err := NewAuthorizeService(NewAuthorizeServiceInput{
		PolicyPath: "policies/relia.yaml",
		Ledger:     store,
		Signer:     fixedSigner{keyID: "k1", priv: priv},
		PublicKey:  pub,
		Broker:     aws.DevBroker{},
	})
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	router := NewRouter(&Handler{AuthorizeService: svc, PublicVerify: true})
	req := httptest.NewRequest(http.MethodGet, "/pack/"+rec.ReceiptID, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestPublicPackMissingReceiptIs404(t *testing.T) {
	router := NewRouter(&Handler{PublicVerify: true, AuthorizeService: &AuthorizeService{Ledger: ledger.NewInMemoryStore()}})
	req := httptest.NewRequest(http.MethodGet, "/pack/missing", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
