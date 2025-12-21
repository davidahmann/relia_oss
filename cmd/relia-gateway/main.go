package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davidahmann/relia_oss/internal/api"
	"github.com/davidahmann/relia_oss/internal/auth"
)

func main() {
	addr := os.Getenv("RELIA_LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	server := newServer(addr)

	log.Printf("relia-gateway listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func newServer(addr string) *http.Server {
	h := &api.Handler{Auth: auth.NewAuthenticatorFromEnv()}
	return &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(h),
		ReadHeaderTimeout: 5 * time.Second,
	}
}
