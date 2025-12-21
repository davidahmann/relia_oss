package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/davidahmann/relia/internal/auth"
	"github.com/davidahmann/relia/internal/slack"
)

type Handler struct {
	Auth             auth.Authenticator
	AuthorizeService *AuthorizeService
	SlackHandler     *slack.InteractionHandler
}

func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	if !h.ensureAuth(w, r) {
		return
	}

	if h.AuthorizeService == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "authorize service not configured"})
		return
	}

	var req AuthorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	claims, err := h.Authenticate(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	actor := ActorContext{
		Subject:  claims.Subject,
		Issuer:   claims.Issuer,
		Repo:     claims.Repo,
		Workflow: claims.Workflow,
		RunID:    claims.RunID,
		SHA:      claims.SHA,
	}

	resp, err := h.AuthorizeService.Authorize(actor, req, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Approvals(w http.ResponseWriter, r *http.Request) {
	if !h.ensureAuth(w, r) {
		return
	}
	if h.AuthorizeService == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "authorize service not configured"})
		return
	}

	approvalID := strings.TrimPrefix(r.URL.Path, "/v1/approvals/")
	if approvalID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing approval_id"})
		return
	}

	approval, ok := h.AuthorizeService.GetApproval(approvalID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "approval not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"approval_id": approval.ApprovalID,
		"status":      string(approval.Status),
		"receipt_id":  approval.ReceiptID,
	})
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
	if h.SlackHandler == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "slack handler not configured"})
		return
	}
	h.SlackHandler.HandleInteractions(w, r)
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
