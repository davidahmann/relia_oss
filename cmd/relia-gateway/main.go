package main

import (
	"context"
	"crypto/ed25519"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/davidahmann/relia/internal/api"
	"github.com/davidahmann/relia/internal/auth"
	"github.com/davidahmann/relia/internal/aws"
	"github.com/davidahmann/relia/internal/config"
	"github.com/davidahmann/relia/internal/crypto"
	"github.com/davidahmann/relia/internal/ledger"
	"github.com/davidahmann/relia/internal/ledger/pgstore"
	"github.com/davidahmann/relia/internal/ledger/sqlstore"
	"github.com/davidahmann/relia/internal/slack"
)

func main() {
	if err := runFn(os.Args[1:], os.Getenv, listenAndServe, newServer); err != nil {
		fatalf("server error: %v", err)
	}
}

var runFn = run
var fatalf = log.Fatalf

type envFn func(string) string
type listenFn func(*http.Server) error
type serverFactory func(cfg config.Config, getenv envFn) (*http.Server, error)

func run(args []string, getenv envFn, listen listenFn, factory serverFactory) error {
	fs := flag.NewFlagSet("relia-gateway", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to relia config file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfgFile := *configPath
	if cfgFile == "" {
		cfgFile = getenv("RELIA_CONFIG_PATH")
	}

	var cfg config.Config
	if cfgFile != "" {
		loaded, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		cfg = loaded
	}

	server, err := factory(cfg, getenv)
	if err != nil {
		return err
	}

	log.Printf("relia-gateway listening on %s", server.Addr)
	if err := listen(server); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func listenAndServe(server *http.Server) error {
	return server.ListenAndServe()
}

func newServer(cfg config.Config, getenv envFn) (*http.Server, error) {
	addr := firstNonEmpty(getenv("RELIA_LISTEN_ADDR"), cfg.ListenAddr, ":8080")
	policyPath := firstNonEmpty(getenv("RELIA_POLICY_PATH"), cfg.PolicyPath, "policies/relia.yaml")

	slackEnabled := cfg.Slack.Enabled
	if raw := getenv("RELIA_SLACK_ENABLED"); raw != "" {
		slackEnabled = envBool(raw)
	}

	signingSecret := firstNonEmpty(getenv("RELIA_SLACK_SIGNING_SECRET"), cfg.Slack.SigningSecret, "")
	slackToken := firstNonEmpty(getenv("RELIA_SLACK_BOT_TOKEN"), cfg.Slack.BotToken, "")
	slackChannel := firstNonEmpty(getenv("RELIA_SLACK_APPROVAL_CHANNEL"), cfg.Slack.ApprovalChannel, "")
	if slackEnabled && signingSecret == "" {
		return nil, logErrorf("missing slack signing secret")
	}

	dbDriver := firstNonEmpty(getenv("RELIA_DB_DRIVER"), cfg.DB.Driver, "sqlite")
	dbDSN := firstNonEmpty(getenv("RELIA_DB_DSN"), cfg.DB.DSN, "file:relia.db?_journal_mode=WAL")

	var store ledger.Store
	if dbDriver == "sqlite" {
		sql, err := sqlstore.OpenSQLite(dbDSN)
		if err != nil {
			return nil, err
		}
		if err := ledger.Migrate(sql.DB(), ledger.DBSQLite); err != nil {
			return nil, err
		}
		store = sql
	} else if dbDriver == "postgres" {
		pg, err := pgstore.OpenPostgres(dbDSN)
		if err != nil {
			return nil, err
		}
		if err := ledger.Migrate(pg.DB(), ledger.DBPostgres); err != nil {
			return nil, err
		}
		store = pg
	} else {
		return nil, logErrorf("unsupported db.driver: %s", dbDriver)
	}

	var signer ledger.Signer
	var pub ed25519.PublicKey

	if cfg.SigningKey.PrivateKeyPath != "" {
		priv, publicKey, err := crypto.LoadEd25519PrivateKey(cfg.SigningKey.PrivateKeyPath)
		if err != nil {
			return nil, err
		}
		pub = publicKey
		keyID := cfg.SigningKey.KeyID
		if keyID == "" {
			keyID = "relia"
		}
		signer = apiDevSigner{keyID: keyID, priv: priv}
	}

	var notifier api.SlackNotifier
	if slackEnabled && slackToken != "" {
		notifier = &slack.Client{Token: slackToken}
	}

	authorizeService, err := api.NewAuthorizeService(api.NewAuthorizeServiceInput{
		PolicyPath: policyPath,
		Ledger:     store,
		Signer:     signer,
		PublicKey:  pub,
		Broker:     awsBrokerFromEnv(getenv, cfg),
		Slack:      notifier,
		SlackChan:  slackChannel,
	})
	if err != nil {
		return nil, err
	}

	slackHandler := &slack.InteractionHandler{
		SigningSecret: signingSecret,
		Approver:      authorizeService,
	}

	h := &api.Handler{
		Auth:             auth.NewAuthenticatorFromEnv(),
		AuthorizeService: authorizeService,
		SlackHandler:     slackHandler,
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(h),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if notifier != nil && slackChannel != "" && getenv("RELIA_SLACK_OUTBOX_WORKER") != "0" {
		ctx, cancel := context.WithCancel(context.Background())
		server.RegisterOnShutdown(cancel)
		go slack.RunOutboxWorker(ctx, store, notifier, 2*time.Second)
	}

	return server, nil
}

type apiDevSigner struct {
	keyID string
	priv  ed25519.PrivateKey
}

func (s apiDevSigner) KeyID() string { return s.keyID }
func (s apiDevSigner) SignEd25519(message []byte) ([]byte, error) {
	return ed25519.Sign(s.priv, message), nil
}

func logErrorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func envBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func awsBrokerFromEnv(getenv envFn, cfg config.Config) aws.CredentialBroker {
	mode := getenv("RELIA_AWS_MODE")
	if mode == "" {
		mode = "dev"
	}
	if mode != "real" {
		return aws.DevBroker{}
	}
	region := cfg.AWS.STSRegionDefault
	if region == "" {
		region = getenv("AWS_REGION")
	}
	broker, err := aws.NewSTSBroker(region)
	if err != nil {
		log.Printf("aws broker init failed, falling back to dev: %v", err)
		return aws.DevBroker{}
	}
	return broker
}
