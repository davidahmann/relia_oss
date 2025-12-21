package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/davidahmann/relia/internal/auth"
	"github.com/davidahmann/relia/internal/grade"
	"github.com/davidahmann/relia/internal/ledger"
	"github.com/davidahmann/relia/internal/pack"
	"github.com/davidahmann/relia/internal/slack"
	"github.com/davidahmann/relia/pkg/types"
)

type Handler struct {
	Auth             auth.Authenticator
	AuthorizeService *AuthorizeService
	SlackHandler     *slack.InteractionHandler
	PublicVerify     bool
}

func (h *Handler) Healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
		Token:    claims.Token,
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
		"status":      approval.Status,
	})
}

func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	if !h.ensureAuth(w, r) {
		return
	}
	if h.AuthorizeService == nil || h.AuthorizeService.Ledger == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "verify not implemented"})
		return
	}

	receiptID := strings.TrimPrefix(r.URL.Path, "/v1/verify/")
	if receiptID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing receipt_id"})
		return
	}

	receiptRec, ok := h.AuthorizeService.Ledger.GetReceipt(receiptID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "receipt not found"})
		return
	}

	if h.AuthorizeService.PublicKey == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "public key not configured"})
		return
	}

	stored := ledger.StoredReceipt{
		ReceiptID:  receiptRec.ReceiptID,
		BodyDigest: receiptRec.BodyDigest,
		BodyJSON:   receiptRec.BodyJSON,
		KeyID:      receiptRec.KeyID,
		Sig:        receiptRec.Sig,
	}
	err := ledger.VerifyReceipt(stored, h.AuthorizeService.PublicKey)
	if err != nil {
		quality := grade.Evaluate(grade.Input{Valid: false, Receipt: stored})
		writeJSON(w, http.StatusOK, map[string]any{
			"receipt_id": receiptID,
			"valid":      false,
			"grade":      quality.Grade,
			"grade_info": map[string]any{"reasons": quality.Reasons},
			"error":      err.Error(),
		})
		return
	}

	var ctxRec types.ContextRecord
	var decRec types.DecisionRecord
	var ctx *types.ContextRecord
	var dec *types.DecisionRecord
	if rec, ok := h.AuthorizeService.Ledger.GetContext(receiptRec.ContextID); ok {
		if err := json.Unmarshal(rec.BodyJSON, &ctxRec); err == nil {
			ctx = &ctxRec
		}
	}
	if rec, ok := h.AuthorizeService.Ledger.GetDecision(receiptRec.DecisionID); ok {
		if err := json.Unmarshal(rec.BodyJSON, &decRec); err == nil {
			dec = &decRec
		}
	}
	quality := grade.Evaluate(grade.Input{
		Valid:    true,
		Receipt:  stored,
		Context:  ctx,
		Decision: dec,
	})

	var body map[string]any
	if err := json.Unmarshal(receiptRec.BodyJSON, &body); err == nil {
		body["integrity"] = map[string]any{
			"body_digest": receiptRec.BodyDigest,
			"signatures": []map[string]any{
				{
					"alg":    "Ed25519",
					"key_id": receiptRec.KeyID,
					"sig":    "base64:" + base64.StdEncoding.EncodeToString(receiptRec.Sig),
				},
			},
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"receipt_id": receiptID,
		"valid":      true,
		"grade":      quality.Grade,
		"grade_info": map[string]any{"reasons": quality.Reasons},
		"receipt":    body,
	})
}

func (h *Handler) Pack(w http.ResponseWriter, r *http.Request) {
	if !h.ensureAuth(w, r) {
		return
	}
	if h.AuthorizeService == nil || h.AuthorizeService.Ledger == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "pack not implemented"})
		return
	}

	receiptID := strings.TrimPrefix(r.URL.Path, "/v1/pack/")
	if receiptID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing receipt_id"})
		return
	}

	receiptRec, ok := h.AuthorizeService.Ledger.GetReceipt(receiptID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "receipt not found"})
		return
	}

	ctxRec, ok := h.AuthorizeService.Ledger.GetContext(receiptRec.ContextID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "context not found"})
		return
	}

	decisionRec, ok := h.AuthorizeService.Ledger.GetDecision(receiptRec.DecisionID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "decision not found"})
		return
	}

	policyVersion, ok := h.AuthorizeService.Ledger.GetPolicyVersion(receiptRec.PolicyHash)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "policy not found"})
		return
	}

	approvals := []pack.ApprovalRecord{}
	if receiptRec.ApprovalID != nil {
		if approval, ok := h.AuthorizeService.GetApproval(*receiptRec.ApprovalID); ok {
			approvals = append(approvals, pack.ApprovalRecord{
				ApprovalID: approval.ApprovalID,
				Status:     approval.Status,
				ReceiptID:  receiptRec.ReceiptID,
			})
		}
	}

	baseURL := ""
	if r.Host != "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		baseURL = scheme + "://" + r.Host
	}

	var ctx types.ContextRecord
	if err := json.Unmarshal(ctxRec.BodyJSON, &ctx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "invalid stored context"})
		return
	}
	var dec types.DecisionRecord
	if err := json.Unmarshal(decisionRec.BodyJSON, &dec); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "invalid stored decision"})
		return
	}

	zipBytes, err := pack.BuildZip(pack.Input{
		Receipt: ledger.StoredReceipt{
			ReceiptID:           receiptRec.ReceiptID,
			BodyDigest:          receiptRec.BodyDigest,
			BodyJSON:            receiptRec.BodyJSON,
			KeyID:               receiptRec.KeyID,
			Sig:                 receiptRec.Sig,
			IdemKey:             receiptRec.IdemKey,
			CreatedAt:           receiptRec.CreatedAt,
			SupersedesReceiptID: receiptRec.SupersedesReceiptID,
			ContextID:           receiptRec.ContextID,
			DecisionID:          receiptRec.DecisionID,
			OutcomeStatus:       types.OutcomeStatus(receiptRec.OutcomeStatus),
			ApprovalID:          receiptRec.ApprovalID,
			PolicyHash:          receiptRec.PolicyHash,
			Final:               receiptRec.Final,
			ExpiresAt:           receiptRec.ExpiresAt,
		},
		Context:   ctx,
		Decision:  dec,
		Policy:    []byte(policyVersion.PolicyYAML),
		Approvals: approvals,
	}, baseURL)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=relia-pack.zip")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(zipBytes)
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
