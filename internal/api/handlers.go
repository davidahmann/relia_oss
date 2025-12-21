package api

import (
	"encoding/json"
	"net/http"

	"github.com/davidahmann/relia_oss/internal/auth"
)

type Handler struct {
	Auth auth.Authenticator
}

func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	if !h.ensureAuth(w, r) {
		return
	}
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "authorize not implemented"})
}

func (h *Handler) Approvals(w http.ResponseWriter, r *http.Request) {
	if !h.ensureAuth(w, r) {
		return
	}
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "approvals not implemented"})
}

func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	if !h.ensureAuth(w, r) {
		return
	}
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "verify not implemented"})
}

func (h *Handler) Pack(w http.ResponseWriter, r *http.Request) {
	if !h.ensureAuth(w, r) {
		return
	}
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "pack not implemented"})
}

func (h *Handler) SlackInteractions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "slack interactions not implemented"})
}

func (h *Handler) ensureAuth(w http.ResponseWriter, r *http.Request) bool {
	_, err := h.Authenticate(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return false
	}
	return true
}

func (h *Handler) Authenticate(r *http.Request) (auth.Claims, error) {
	return h.Auth.Authenticate(r)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}
