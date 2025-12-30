package pack

import (
	"archive/zip"
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/davidahmann/relia/internal/context"
	"github.com/davidahmann/relia/internal/decision"
	"github.com/davidahmann/relia/internal/ledger"
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
	return ed25519.Sign(s.priv, message), nil
}

func TestBuildZipIncludesArtifacts(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	priv := ed25519.NewKeyFromSeed(seed)

	createdAt := time.Now().UTC().Format(time.RFC3339)
	source := types.ContextSource{Kind: "github_actions", Repo: "org/repo", Workflow: "wf", RunID: "1", Actor: "dev", Ref: "refs/heads/main", SHA: "abc"}
	inputs := types.ContextInputs{Action: "terraform.apply", Resource: "res", Env: "prod"}
	evidence := types.ContextEvidence{PlanDigest: "sha256:abc"}
	ctx, err := context.BuildContext(source, inputs, evidence, createdAt)
	if err != nil {
		t.Fatalf("context: %v", err)
	}

	policy := types.DecisionPolicy{PolicyID: "relia-default", PolicyVersion: "2025-12-20", PolicyHash: "sha256:policy"}
	dec, err := decision.BuildDecision(ctx.ContextID, policy, "allow", nil, false, "high", createdAt)
	if err != nil {
		t.Fatalf("decision: %v", err)
	}

	receipt, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
		CreatedAt:  createdAt,
		IdemKey:    "idem",
		ContextID:  ctx.ContextID,
		DecisionID: dec.DecisionID,
		Actor:      types.ReceiptActor{Kind: "workload", Subject: "dev"},
		Request:    types.ReceiptRequest{RequestID: "req", Action: "terraform.apply", Resource: "res", Env: "prod"},
		Policy:     types.ReceiptPolicy{PolicyHash: "sha256:policy"},
		Outcome:    types.ReceiptOutcome{Status: types.OutcomeDenied},
	}, testSigner{keyID: "test", priv: priv})
	if err != nil {
		t.Fatalf("receipt: %v", err)
	}

	zipBytes, err := BuildZip(Input{
		Receipt:   receipt,
		Context:   ctx,
		Decision:  dec,
		Policy:    []byte("policy_id: relia-default\n"),
		Approvals: []ApprovalRecord{{ApprovalID: "approval-1", Status: "approved", ReceiptID: receipt.ReceiptID}},
	}, "http://localhost:8080")
	if err != nil {
		t.Fatalf("build zip: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip reader: %v", err)
	}

	expected := map[string]bool{
		"receipt.json":   false,
		"context.json":   false,
		"decision.json":  false,
		"policy.yaml":    false,
		"approvals.json": false,
		"summary.json":   false,
		"summary.html":   false,
		"manifest.json":  false,
		"sha256sums.txt": false,
	}

	for _, file := range reader.File {
		if _, ok := expected[file.Name]; ok {
			expected[file.Name] = true
		}
	}

	for name, seen := range expected {
		if !seen {
			t.Fatalf("missing %s", name)
		}
	}
}

func TestBuildZipManifestIncludesRefs(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	priv := ed25519.NewKeyFromSeed(seed)

	createdAt := time.Now().UTC().Format(time.RFC3339)
	source := types.ContextSource{Kind: "github_actions", Repo: "org/repo", Workflow: "wf", RunID: "1", Actor: "dev", Ref: "refs/heads/main", SHA: "abc"}
	inputs := types.ContextInputs{Action: "terraform.apply", Resource: "res", Env: "prod"}
	ctx, err := context.BuildContext(source, inputs, types.ContextEvidence{}, createdAt)
	if err != nil {
		t.Fatalf("context: %v", err)
	}
	policy := types.DecisionPolicy{PolicyID: "relia-default", PolicyVersion: "2025-12-20", PolicyHash: "sha256:policy"}
	dec, err := decision.BuildDecision(ctx.ContextID, policy, "deny", nil, false, "low", createdAt)
	if err != nil {
		t.Fatalf("decision: %v", err)
	}

	receipt, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
		CreatedAt:      createdAt,
		IdemKey:        "idem",
		ContextID:      ctx.ContextID,
		DecisionID:     dec.DecisionID,
		Actor:          types.ReceiptActor{Kind: "workload", Subject: "dev"},
		Request:        types.ReceiptRequest{RequestID: "req", Action: "terraform.apply", Resource: "res", Env: "prod"},
		Policy:         types.ReceiptPolicy{PolicyHash: "sha256:policy"},
		InteractionRef: &types.InteractionRef{Mode: "voice", CallID: "call-1", TurnID: "turn-1", TurnIndex: 1},
		Refs: &types.ReceiptRefs{
			Context:  &types.ContextRef{ContextID: "context-1", RecordHash: "sha256:ctxrecord"},
			Decision: &types.DecisionRef{DecisionID: "decision-1", InputsDigest: "sha256:decinputs"},
		},
		Outcome: types.ReceiptOutcome{Status: types.OutcomeDenied},
	}, testSigner{keyID: "test", priv: priv})
	if err != nil {
		t.Fatalf("receipt: %v", err)
	}

	zipBytes, err := BuildZip(Input{
		Receipt:  receipt,
		Context:  ctx,
		Decision: dec,
		Policy:   []byte("policy_id: relia-default\n"),
	}, "http://localhost:8080")
	if err != nil {
		t.Fatalf("build zip: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip reader: %v", err)
	}

	var manifestBytes []byte
	for _, f := range reader.File {
		if f.Name != "manifest.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open manifest: %v", err)
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read manifest: %v", err)
		}
		manifestBytes = b
		break
	}
	if len(manifestBytes) == 0 {
		t.Fatalf("missing manifest.json")
	}

	var manifest types.PackManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	if manifest.Refs == nil || manifest.Refs.Context == nil || manifest.Refs.Decision == nil {
		t.Fatalf("expected refs in manifest")
	}
	if manifest.Refs.Context.RecordHash != "sha256:ctxrecord" {
		t.Fatalf("unexpected context refs: %+v", manifest.Refs.Context)
	}
	if manifest.Refs.Decision.InputsDigest != "sha256:decinputs" {
		t.Fatalf("unexpected decision refs: %+v", manifest.Refs.Decision)
	}
	if manifest.InteractionRef == nil || manifest.InteractionRef.CallID != "call-1" {
		t.Fatalf("unexpected interaction_ref: %+v", manifest.InteractionRef)
	}
}

func TestBuildFilesRequiresPolicy(t *testing.T) {
	_, err := BuildFiles(Input{}, "")
	if err == nil {
		t.Fatalf("expected error for missing policy")
	}
}

func TestBuildZipRequiresPolicy(t *testing.T) {
	_, err := BuildZip(Input{}, "")
	if err == nil {
		t.Fatalf("expected error for missing policy")
	}
}

func TestWriteZip(t *testing.T) {
	files := map[string][]byte{
		"a.txt": []byte("alpha"),
		"b.txt": []byte("bravo"),
	}
	buf := bytes.NewBuffer(nil)
	if err := WriteZip(buf, files); err != nil {
		t.Fatalf("write zip: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("zip reader: %v", err)
	}
	if len(reader.File) != 2 {
		t.Fatalf("expected 2 files, got %d", len(reader.File))
	}
}

type failingWriter struct{}

func (f failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestWriteZipWriterError(t *testing.T) {
	files := map[string][]byte{"a.txt": []byte("alpha")}
	if err := WriteZip(failingWriter{}, files); err == nil {
		t.Fatalf("expected error")
	}
}

func TestBuildFilesInvalidReceiptJSON(t *testing.T) {
	_, err := BuildFiles(Input{
		Receipt: ledger.StoredReceipt{
			ReceiptID:  "r",
			BodyDigest: "digest",
			BodyJSON:   []byte("not-json"),
			KeyID:      "kid",
			Sig:        []byte("sig"),
			PolicyHash: "ph",
		},
		Policy: []byte("policy_id: x\n"),
	}, "")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestExtractReceiptRefs(t *testing.T) {
	bodyWithRefs, err := json.Marshal(map[string]any{
		"refs": &types.ReceiptRefs{
			Context: &types.ContextRef{
				ContextID:  "sha256:ctx",
				RecordHash: "sha256:ctx_record",
			},
			Decision: &types.DecisionRef{
				DecisionID:   "sha256:dec",
				InputsDigest: "sha256:inputs",
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	cases := []struct {
		name    string
		body    []byte
		wantNil bool
	}{
		{name: "empty", body: nil, wantNil: true},
		{name: "invalid_json", body: []byte("not-json"), wantNil: true},
		{name: "missing_refs", body: []byte(`{"x":1}`), wantNil: true},
		{name: "null_refs", body: []byte(`{"refs":null}`), wantNil: true},
		{name: "empty_refs_obj", body: []byte(`{"refs":{}}`), wantNil: true},
		{name: "valid_refs", body: bodyWithRefs, wantNil: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractReceiptRefs(tc.body)
			if (got == nil) != tc.wantNil {
				t.Fatalf("got nil=%v, wantNil=%v", got == nil, tc.wantNil)
			}
		})
	}
}

func TestExtractReceiptInteractionRef(t *testing.T) {
	bodyWithRef, err := json.Marshal(map[string]any{
		"interaction_ref": &types.InteractionRef{Mode: "voice", CallID: "call-1", TurnID: "turn-1", TurnIndex: 1},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := extractReceiptInteractionRef(bodyWithRef); got == nil || got.CallID != "call-1" {
		t.Fatalf("unexpected interaction_ref: %+v", got)
	}
	if got := extractReceiptInteractionRef([]byte("not-json")); got != nil {
		t.Fatalf("expected nil for invalid json")
	}
}
