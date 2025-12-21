package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davidahmann/relia/internal/api"
	"github.com/davidahmann/relia/internal/auth"
	"github.com/davidahmann/relia/internal/slack"
)

func main() {
	addr := os.Getenv("RELIA_LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	policyPath := os.Getenv("RELIA_POLICY_PATH")
	if policyPath == "" {
		policyPath = "policies/relia.yaml"
	}

	signingSecret := os.Getenv("RELIA_SLACK_SIGNING_SECRET")

	server := newServer(addr, policyPath, signingSecret)

	log.Printf("relia-gateway listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func newServer(addr string, policyPath string, signingSecret string) *http.Server {
	authorizeService, err := api.NewAuthorizeService(policyPath)
	if err != nil {
		log.Fatalf("authorize service error: %v", err)
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
	return &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(h),
		ReadHeaderTimeout: 5 * time.Second,
	}
}
