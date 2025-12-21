package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davidahmann/relia/internal/aws"
	"github.com/davidahmann/relia/internal/config"
)

func TestNewServer(t *testing.T) {
	cfg := config.Config{
		ListenAddr: ":9999",
		PolicyPath: "policies/relia.yaml",
		DB:         config.DBConfig{Driver: "sqlite", DSN: "file::memory:?cache=shared"},
	}
	srv, err := newServer(cfg, func(string) string { return "" })
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	if srv.Addr != ":9999" {
		t.Fatalf("expected addr %s, got %s", ":9999", srv.Addr)
	}
	if srv.Handler == nil {
		t.Fatalf("expected handler to be set")
	}
}

func TestNewServerUnsupportedDBDriver(t *testing.T) {
	cfg := config.Config{
		ListenAddr: ":9999",
		PolicyPath: "policies/relia.yaml",
		DB:         config.DBConfig{Driver: "postgres", DSN: "ignored"},
	}
	_, err := newServer(cfg, func(string) string { return "" })
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestNewServerSigningKeyAndSlack(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "ed25519.key")
	seed := make([]byte, ed25519.SeedSize)
	if err := os.WriteFile(keyPath, seed, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	cfg := config.Config{
		ListenAddr: ":9999",
		PolicyPath: "policies/relia.yaml",
		DB:         config.DBConfig{Driver: "sqlite", DSN: "file::memory:?cache=shared"},
		SigningKey: config.SigningKeyConfig{PrivateKeyPath: keyPath},
		Slack: config.SlackConfig{
			Enabled:         true,
			BotToken:        "xoxb-test",
			SigningSecret:   "signing-secret",
			ApprovalChannel: "C123",
		},
	}
	srv, err := newServer(cfg, func(key string) string {
		if key == "RELIA_SLACK_OUTBOX_WORKER" {
			return "0"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	if srv.Handler == nil {
		t.Fatalf("expected handler")
	}
}

func TestNewServerStartsSlackOutboxWorker(t *testing.T) {
	cfg := config.Config{
		ListenAddr: ":9999",
		PolicyPath: "policies/relia.yaml",
		DB:         config.DBConfig{Driver: "sqlite", DSN: "file::memory:?cache=shared"},
		Slack: config.SlackConfig{
			Enabled:         true,
			BotToken:        "xoxb-test",
			SigningSecret:   "signing-secret",
			ApprovalChannel: "C123",
		},
	}
	srv, err := newServer(cfg, func(key string) string {
		if key == "RELIA_SLACK_OUTBOX_WORKER" {
			return "1"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestRunDefaults(t *testing.T) {
	factory := func(cfg config.Config, getenv envFn) (*http.Server, error) {
		return &http.Server{Addr: ":8080"}, nil
	}

	listen := func(_ *http.Server) error {
		return http.ErrServerClosed
	}

	getenv := func(string) string { return "" }
	if err := run(nil, getenv, listen, factory); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLoadsConfigFile(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "relia.yaml")
	cfgYAML := `listen_addr: ":8088"
policy_path: "policies/relia.yaml"
db:
  driver: "sqlite"
  dsn: "file::memory:?cache=shared"
`
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	called := false
	factory := func(cfg config.Config, _ envFn) (*http.Server, error) {
		called = true
		if cfg.ListenAddr != ":8088" {
			t.Fatalf("expected addr %q, got %q", ":8088", cfg.ListenAddr)
		}
		if cfg.DB.Driver != "sqlite" {
			t.Fatalf("expected sqlite driver, got %q", cfg.DB.Driver)
		}
		return &http.Server{Addr: cfg.ListenAddr}, nil
	}

	listen := func(_ *http.Server) error { return http.ErrServerClosed }
	getenv := func(string) string { return "" }
	if err := run([]string{"-config", cfgPath}, getenv, listen, factory); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected server factory to be called")
	}
}

func TestRunError(t *testing.T) {
	listenErr := errors.New("listen failed")
	listen := func(_ *http.Server) error {
		return listenErr
	}

	factory := func(cfg config.Config, getenv envFn) (*http.Server, error) {
		return &http.Server{Addr: ":8080"}, nil
	}

	getenv := func(string) string { return "" }

	if err := run(nil, getenv, listen, factory); err == nil {
		t.Fatalf("expected error")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "a", "b"); got != "a" {
		t.Fatalf("expected a, got %s", got)
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Fatalf("expected empty, got %s", got)
	}
}

func TestListenAndServeInvalidAddr(t *testing.T) {
	err := listenAndServe(&http.Server{Addr: "127.0.0.1"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestMainNoError(t *testing.T) {
	oldRun := runFn
	oldFatal := fatalf
	defer func() {
		runFn = oldRun
		fatalf = oldFatal
	}()

	runFn = func(args []string, envFn envFn, listenFn listenFn, serverFactory serverFactory) error {
		return nil
	}

	called := false
	fatalf = func(string, ...any) {
		called = true
	}

	main()
	if called {
		t.Fatalf("unexpected fatal call")
	}
}

func TestMainError(t *testing.T) {
	oldRun := runFn
	oldFatal := fatalf
	defer func() {
		runFn = oldRun
		fatalf = oldFatal
	}()

	runFn = func(args []string, envFn envFn, listenFn listenFn, serverFactory serverFactory) error {
		return errors.New("boom")
	}

	called := false
	fatalf = func(string, ...any) {
		called = true
	}

	main()
	if !called {
		t.Fatalf("expected fatal call")
	}
}

func TestAPIDevSigner(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		t.Fatalf("rand: %v", err)
	}
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)

	signer := apiDevSigner{keyID: "k", priv: priv}
	if signer.KeyID() != "k" {
		t.Fatalf("expected key id")
	}
	msg := []byte("hello")
	sig, err := signer.SignEd25519(msg)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if !ed25519.Verify(pub, msg, sig) {
		t.Fatalf("signature did not verify")
	}
}

func TestLogErrorf(t *testing.T) {
	err := logErrorf("hello %s", "world")
	if err == nil || err.Error() != "hello world" {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestAWSBrokerFromEnv(t *testing.T) {
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")

	cfg := config.Config{}
	getenv := func(key string) string {
		if key == "RELIA_AWS_MODE" {
			return "real"
		}
		return ""
	}
	if b := awsBrokerFromEnv(getenv, cfg); func() bool { _, ok := b.(aws.DevBroker); return ok }() == false {
		t.Fatalf("expected fallback to dev broker when region missing")
	}

	cfg.AWS.STSRegionDefault = "us-east-1"
	if b := awsBrokerFromEnv(getenv, cfg); func() bool { _, ok := b.(aws.DevBroker); return ok }() {
		t.Fatalf("expected real broker when region present")
	}
}
