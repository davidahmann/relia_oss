package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_UsageAndUnknown(t *testing.T) {
	var out, errOut bytes.Buffer

	if code := run([]string{"relia"}, &out, &errOut); code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}

	out.Reset()
	errOut.Reset()
	if code := run([]string{"relia", "nope"}, &out, &errOut); code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "Usage:") {
		t.Fatalf("expected usage output")
	}
}

func TestHandleVerify(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/verify/r1" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Fatalf("unexpected auth header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"receipt_id":"r1","valid":true,"grade":"A"}`))
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	var out, errOut bytes.Buffer
	code := handleVerify([]string{"--addr", srv.URL, "--token", "tok", "r1"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "valid=true") {
		t.Fatalf("unexpected stdout: %s", out.String())
	}
	if !strings.Contains(out.String(), "grade=A") {
		t.Fatalf("expected grade output, got: %s", out.String())
	}

	out.Reset()
	errOut.Reset()
	code = handleVerify([]string{"--addr", srv.URL, "--token", "tok", "--json", "r1"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
	if strings.TrimSpace(out.String()) != `{"receipt_id":"r1","valid":true,"grade":"A"}` {
		t.Fatalf("unexpected json stdout: %s", out.String())
	}
}

func TestHandleVerify_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"nope"}`))
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	code := handleVerify([]string{"--addr", srv.URL, "--token", "tok", "missing"}, &out, &errOut)
	if code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
}

func TestHandleVerify_Errors(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := handleVerify([]string{}, &out, &errOut); code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}

	out.Reset()
	errOut.Reset()
	if code := handleVerify([]string{"--nope"}, &out, &errOut); code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}

	invalidJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	defer invalidJSON.Close()

	out.Reset()
	errOut.Reset()
	if code := handleVerify([]string{"--addr", invalidJSON.URL, "r1"}, &out, &errOut); code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}

	validFalse := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"receipt_id":"r1","valid":false,"error":"bad"}`))
	}))
	defer validFalse.Close()

	out.Reset()
	errOut.Reset()
	if code := handleVerify([]string{"--addr", validFalse.URL, "r1"}, &out, &errOut); code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
	if !strings.Contains(out.String(), "valid=false") {
		t.Fatalf("unexpected stdout: %s", out.String())
	}
}

func TestHandlePack(t *testing.T) {
	payload := []byte("zip-bytes")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/pack/r1" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	tmp := t.TempDir()
	outPath := filepath.Join(tmp, "out", "pack.zip")

	var out, errOut bytes.Buffer
	code := handlePack([]string{"--addr", srv.URL, "--token", "tok", "--out", outPath, "r1"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d stderr=%s", code, errOut.String())
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read out: %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("output mismatch")
	}
}

func TestHandlePack_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad"))
	}))
	defer srv.Close()

	var out, errOut bytes.Buffer
	code := handlePack([]string{"--addr", srv.URL, "--token", "tok", "--out", filepath.Join(t.TempDir(), "p.zip"), "r1"}, &out, &errOut)
	if code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
}

func TestHandlePack_Errors(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := handlePack([]string{}, &out, &errOut); code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}

	okSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("x"))
	}))
	defer okSrv.Close()

	tmp := t.TempDir()
	parentFile := filepath.Join(tmp, "file")
	if err := os.WriteFile(parentFile, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	outPath := filepath.Join(parentFile, "pack.zip") // parent is a file, mkdir should fail

	out.Reset()
	errOut.Reset()
	if code := handlePack([]string{"--addr", okSrv.URL, "--out", outPath, "r1"}, &out, &errOut); code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}

	out.Reset()
	errOut.Reset()
	if code := handlePack([]string{"--addr", okSrv.URL, "--out", tmp, "r1"}, &out, &errOut); code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
}

func TestHandlePolicyLint(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	policyYAML := `
policy_id: test
policy_version: "1"
defaults:
  ttl_seconds: 900
  require_approval: false
  deny: false
rules:
  - id: allow
    match:
      action: "terraform.apply"
      env: "dev"
    effect:
      ttl_seconds: 900
      aws_role_arn: "arn:aws:iam::123456789012:role/test"
`
	if err := os.WriteFile(path, []byte(policyYAML), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	var out, errOut bytes.Buffer
	code := run([]string{"relia", "policy", "lint", path}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "ok policy_id=") {
		t.Fatalf("unexpected stdout: %s", out.String())
	}
}

func TestHandlePolicy_Errors(t *testing.T) {
	var out, errOut bytes.Buffer

	if code := run([]string{"relia", "policy"}, &out, &errOut); code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}
	out.Reset()
	errOut.Reset()
	if code := run([]string{"relia", "policy", "nope"}, &out, &errOut); code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}

	out.Reset()
	errOut.Reset()
	if code := run([]string{"relia", "policy", "lint"}, &out, &errOut); code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}

	out.Reset()
	errOut.Reset()
	if code := run([]string{"relia", "policy", "lint", "missing.yaml"}, &out, &errOut); code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
}

func TestHandlePolicyTest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	policyYAML := `
policy_id: test
policy_version: "1"
defaults:
  ttl_seconds: 900
  require_approval: false
  deny: false
rules:
  - id: prod_terraform
    match:
      action: "terraform.apply"
      env: "prod"
    effect:
      require_approval: true
      ttl_seconds: 900
      aws_role_arn: "arn:aws:iam::123456789012:role/test"
`
	if err := os.WriteFile(path, []byte(policyYAML), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	var out, errOut bytes.Buffer
	code := run([]string{"relia", "policy", "test", "--policy", path, "--action", "terraform.apply", "--resource", "stack/prod", "--env", "prod"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "matched_rule=prod_terraform") {
		t.Fatalf("unexpected output: %s", out.String())
	}

	out.Reset()
	errOut.Reset()
	code = run([]string{"relia", "policy", "test", "--policy", path, "--action", "terraform.apply", "--resource", "stack/prod", "--env", "prod", "--json"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
	if !strings.Contains(out.String(), "\"MatchedRuleID\"") {
		t.Fatalf("expected json decision output, got: %s", out.String())
	}
}

func TestHandlePolicyTest_Errors(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"relia", "policy", "test"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}

	out.Reset()
	errOut.Reset()
	code = run([]string{"relia", "policy", "test", "--policy", "missing.yaml", "--action", "a", "--resource", "r", "--env", "e"}, &out, &errOut)
	if code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
}

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("RELIA_TEST_ENV", "x")
	if got := envOrDefault("RELIA_TEST_ENV", "y"); got != "x" {
		t.Fatalf("expected x, got %s", got)
	}
	if got := envOrDefault("RELIA_MISSING_ENV", "y"); got != "y" {
		t.Fatalf("expected y, got %s", got)
	}
}

func TestMainCallsExit(t *testing.T) {
	oldExit := exitFn
	oldArgs := os.Args
	defer func() {
		exitFn = oldExit
		os.Args = oldArgs
	}()

	var got int
	exitFn = func(code int) { got = code }
	os.Args = []string{"relia"}
	main()

	if got != 2 {
		t.Fatalf("expected exit code 2, got %d", got)
	}
}

func TestKeysGenWritesFiles(t *testing.T) {
	tmp := t.TempDir()
	priv := filepath.Join(tmp, "keys", "ed25519.key")
	pub := filepath.Join(tmp, "keys", "ed25519.pub")

	var out, errOut bytes.Buffer
	code := run([]string{"relia", "keys", "gen", "--private", priv, "--public", pub}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d stderr=%s", code, errOut.String())
	}

	if _, err := os.Stat(priv); err != nil {
		t.Fatalf("expected private key file: %v", err)
	}
	if _, err := os.Stat(pub); err != nil {
		t.Fatalf("expected public key file: %v", err)
	}
}

func TestKeysGenDoesNotOverwriteByDefault(t *testing.T) {
	tmp := t.TempDir()
	priv := filepath.Join(tmp, "ed25519.key")
	if err := os.WriteFile(priv, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	var out, errOut bytes.Buffer
	code := run([]string{"relia", "keys", "gen", "--private", priv}, &out, &errOut)
	if code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
}

func TestKeysGenOverwriteAndFormats(t *testing.T) {
	tmp := t.TempDir()
	priv := filepath.Join(tmp, "ed25519.key")
	pub := filepath.Join(tmp, "ed25519.pub")

	if err := os.WriteFile(priv, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	var out, errOut bytes.Buffer
	code := run([]string{"relia", "keys", "gen", "--private", priv, "--public", pub, "--overwrite", "--format", "raw"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d stderr=%s", code, errOut.String())
	}

	privBytes, err := os.ReadFile(priv)
	if err != nil {
		t.Fatalf("read priv: %v", err)
	}
	if len(privBytes) != 32 {
		t.Fatalf("expected 32-byte seed, got %d", len(privBytes))
	}

	pubBytes, err := os.ReadFile(pub)
	if err != nil {
		t.Fatalf("read pub: %v", err)
	}
	if len(pubBytes) != 32 {
		t.Fatalf("expected 32-byte public key, got %d", len(pubBytes))
	}
}

func TestKeysGenInvalidFormat(t *testing.T) {
	tmp := t.TempDir()
	priv := filepath.Join(tmp, "ed25519.key")

	var out, errOut bytes.Buffer
	code := run([]string{"relia", "keys", "gen", "--private", priv, "--format", "nope"}, &out, &errOut)
	if code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
}

func TestKeysGenBase64Format(t *testing.T) {
	tmp := t.TempDir()
	priv := filepath.Join(tmp, "ed25519.key")

	var out, errOut bytes.Buffer
	code := run([]string{"relia", "keys", "gen", "--private", priv, "--format", "base64"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d stderr=%s", code, errOut.String())
	}
	b, err := os.ReadFile(priv)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.HasPrefix(string(b), "base64:") {
		t.Fatalf("expected base64 prefix, got %q", string(b))
	}
}

func TestKeysCommandUsage(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"relia", "keys"}, &out, &errOut); code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}

	out.Reset()
	errOut.Reset()
	if code := run([]string{"relia", "keys", "gen"}, &out, &errOut); code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}
}
