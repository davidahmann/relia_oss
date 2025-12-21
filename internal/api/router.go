package api

import (
	"net/http"
)

func NewRouter(handler *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/authorize", handler.Authorize)
	mux.HandleFunc("/v1/approvals/", handler.Approvals)
	mux.HandleFunc("/v1/verify/", handler.Verify)
	mux.HandleFunc("/v1/pack/", handler.Pack)
	mux.HandleFunc("/v1/slack/interactions", handler.SlackInteractions)

	return mux
}
