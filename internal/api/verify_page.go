package api

import (
	"crypto/ed25519"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/davidahmann/relia/internal/grade"
	"github.com/davidahmann/relia/internal/ledger"
	"github.com/davidahmann/relia/internal/pack"
	"github.com/davidahmann/relia/pkg/types"
)

var verifyPageTmpl = template.Must(template.New("verify").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>Relia Verify</title>
  <style>
    body{font-family:ui-sans-serif,system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial; margin:24px; color:#0f172a; background:#ffffff}
    .card{max-width:920px; border:1px solid #e2e8f0; border-radius:12px; padding:18px 18px; box-shadow:0 1px 2px rgba(0,0,0,.04)}
    .row{display:flex; flex-wrap:wrap; gap:12px}
    .pill{display:inline-block; padding:4px 10px; border-radius:999px; font-size:12px; background:#f1f5f9}
    .ok{background:#dcfce7}
    .bad{background:#fee2e2}
    .muted{color:#475569}
    code{background:#f1f5f9; padding:2px 6px; border-radius:6px}
    a{color:#2563eb; text-decoration:none}
    a:hover{text-decoration:underline}
    h1{font-size:18px; margin:0 0 12px 0}
    .k{width:220px; font-size:12px; color:#475569}
    .v{font-size:13px}
    .kv{display:flex; gap:12px; padding:6px 0; border-bottom:1px dashed #e2e8f0}
    .kv:last-child{border-bottom:none}
    .top{display:flex; justify-content:space-between; align-items:center; gap:12px}
    .actions{display:flex; gap:12px; flex-wrap:wrap}
  </style>
</head>
<body>
  <div class="card">
    <div class="top">
      <h1>Relia Verify</h1>
      <div class="actions">
        {{if .PackURL}}<a href="{{.PackURL}}">Download pack</a>{{end}}
        {{if .JSONURL}}<a href="{{.JSONURL}}">JSON</a>{{end}}
      </div>
    </div>

    <div class="row" style="margin:10px 0 12px 0">
      <span class="pill {{if .Valid}}ok{{else}}bad{{end}}">{{if .Valid}}VALID{{else}}INVALID{{end}}</span>
      {{if .Grade}}<span class="pill">Grade: {{.Grade}}</span>{{end}}
      <span class="pill">Receipt: <code>{{.ReceiptID}}</code></span>
    </div>

    {{if .Error}}
      <div class="pill bad">Error: {{.Error}}</div>
    {{end}}

	    {{if .Valid}}
	      <div class="kv"><div class="k">Verdict</div><div class="v">{{.Verdict}}</div></div>
	      <div class="kv"><div class="k">Approval</div><div class="v">{{.Approval}}</div></div>
	      <div class="kv"><div class="k">Policy</div><div class="v">{{.Policy}}</div></div>
	      <div class="kv"><div class="k">Interaction</div><div class="v">{{if .Interaction}}<code>{{.Interaction}}</code>{{else}}<span class="muted">n/a</span>{{end}}</div></div>
	      <div class="kv"><div class="k">Refs</div><div class="v">{{if .Refs}}<code>{{.Refs}}</code>{{else}}<span class="muted">n/a</span>{{end}}</div></div>
	      <div class="kv"><div class="k">Role / TTL</div><div class="v">{{.RoleTTL}}</div></div>
	      <div class="kv"><div class="k">GitHub Run</div><div class="v">{{if .RunURL}}<a href="{{.RunURL}}">{{.RunURL}}</a>{{else}}<span class="muted">n/a</span>{{end}}</div></div>
	      <div class="kv"><div class="k">Plan Digest</div><div class="v">{{if .PlanDigest}}<code>{{.PlanDigest}}</code>{{else}}<span class="muted">n/a</span>{{end}}</div></div>
	      <div class="kv"><div class="k">Diff URL</div><div class="v">{{if .DiffURL}}<a href="{{.DiffURL}}">{{.DiffURL}}</a>{{else}}<span class="muted">n/a</span>{{end}}</div></div>
	    {{end}}
	  </div>
	</body>
	</html>`))

func (h *Handler) VerifyPage(w http.ResponseWriter, r *http.Request) {
	if !h.PublicVerify {
		http.NotFound(w, r)
		return
	}
	if h.AuthorizeService == nil || h.AuthorizeService.Ledger == nil || h.AuthorizeService.PublicKey == nil {
		http.NotFound(w, r)
		return
	}

	receiptID := strings.TrimPrefix(r.URL.Path, "/verify/")
	if receiptID == "" {
		http.NotFound(w, r)
		return
	}

	receiptRec, ok := h.AuthorizeService.Ledger.GetReceipt(receiptID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	stored := ledger.StoredReceipt{
		ReceiptID:  receiptRec.ReceiptID,
		BodyDigest: receiptRec.BodyDigest,
		BodyJSON:   receiptRec.BodyJSON,
		KeyID:      receiptRec.KeyID,
		Sig:        receiptRec.Sig,
	}
	pub, ok := h.AuthorizeService.Ledger.GetKey(receiptRec.KeyID)
	var verifyKey ed25519.PublicKey
	if ok {
		verifyKey = ed25519.PublicKey(pub.PublicKey)
	} else {
		verifyKey = h.AuthorizeService.PublicKey
	}
	err := ledger.VerifyReceipt(stored, verifyKey)
	valid := err == nil

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

	q := grade.Evaluate(grade.Input{Valid: valid, Receipt: stored, Context: ctx, Decision: dec})

	var rb struct {
		Policy          types.ReceiptPolicy           `json:"policy"`
		Approval        *types.ReceiptApproval        `json:"approval,omitempty"`
		InteractionRef  *types.InteractionRef         `json:"interaction_ref,omitempty"`
		Refs            *types.ReceiptRefs            `json:"refs,omitempty"`
		CredentialGrant *types.ReceiptCredentialGrant `json:"credential_grant,omitempty"`
	}
	_ = json.Unmarshal(receiptRec.BodyJSON, &rb)

	type view struct {
		Valid       bool
		Error       string
		ReceiptID   string
		Grade       string
		Verdict     string
		Approval    string
		Policy      string
		RoleTTL     string
		Interaction string
		Refs        string
		RunURL      string
		PlanDigest  string
		DiffURL     string
		PackURL     string
		JSONURL     string
	}

	v := view{
		Valid:     valid,
		ReceiptID: receiptID,
		Grade:     q.Grade,
		PackURL:   "/pack/" + receiptID,
		JSONURL:   "/v1/verify/" + receiptID,
	}
	if err != nil {
		v.Error = err.Error()
	}

	if dec != nil {
		v.Verdict = dec.Verdict
		v.Policy = dec.Policy.PolicyHash
		if dec.Policy.PolicyID != "" || dec.Policy.PolicyVersion != "" {
			v.Policy = dec.Policy.PolicyID + "@" + dec.Policy.PolicyVersion + " (" + dec.Policy.PolicyHash + ")"
		}
	} else if rb.Policy.PolicyHash != "" {
		v.Policy = rb.Policy.PolicyHash
	}

	v.Approval = "not required"
	if rb.Approval != nil && rb.Approval.Required {
		v.Approval = rb.Approval.Status
		if rb.Approval.Approver != nil && rb.Approval.Approver.Display != "" {
			v.Approval += " by " + rb.Approval.Approver.Display
		}
		if rb.Approval.ApprovedAt != "" {
			v.Approval += " at " + rb.Approval.ApprovedAt
		}
	}

	if ctx != nil {
		// Best-effort GitHub run URL from evidence diff_url (examples use that).
		if strings.Contains(ctx.Evidence.DiffURL, "github.com/") {
			v.RunURL = ctx.Evidence.DiffURL
		}
	}

	v.RoleTTL = "n/a"
	if rb.CredentialGrant != nil {
		role := rb.CredentialGrant.RoleARN
		ttl := rb.CredentialGrant.TTLSeconds
		if role != "" || ttl != 0 {
			v.RoleTTL = role
			if ttl != 0 {
				v.RoleTTL += " (ttl " + int64ToString(ttl) + "s)"
			}
		}
	}
	if receiptRec.ExpiresAt != nil && *receiptRec.ExpiresAt != "" {
		v.RoleTTL += " expires " + *receiptRec.ExpiresAt
	}

	if ctx != nil {
		v.PlanDigest = ctx.Evidence.PlanDigest
		v.DiffURL = ctx.Evidence.DiffURL
	}

	v.Interaction = ""
	if rb.InteractionRef != nil {
		ir := rb.InteractionRef
		parts := []string{}
		if ir.Mode != "" {
			parts = append(parts, "mode="+ir.Mode)
		}
		if ir.CallID != "" {
			parts = append(parts, "call_id="+ir.CallID)
		}
		if ir.TurnID != "" {
			parts = append(parts, "turn_id="+ir.TurnID)
		}
		if ir.TurnIndex != 0 {
			parts = append(parts, "turn_index="+int64ToString(int64(ir.TurnIndex)))
		}
		if ir.ConsentState != "" {
			parts = append(parts, "consent="+ir.ConsentState)
		}
		v.Interaction = strings.Join(parts, " ")
	}

	v.Refs = ""
	if rb.Refs != nil {
		if rb.Refs.Context != nil {
			if rb.Refs.Context.ContextID != "" {
				v.Refs += "context_id=" + rb.Refs.Context.ContextID + " "
			}
			if rb.Refs.Context.RecordHash != "" {
				v.Refs += "record_hash=" + rb.Refs.Context.RecordHash + " "
			}
			if rb.Refs.Context.ContentHash != "" {
				v.Refs += "content_hash=" + rb.Refs.Context.ContentHash + " "
			}
		}
		if rb.Refs.Decision != nil {
			if rb.Refs.Decision.DecisionID != "" {
				v.Refs += "decision_id=" + rb.Refs.Decision.DecisionID + " "
			}
			if rb.Refs.Decision.InputsDigest != "" {
				v.Refs += "inputs_digest=" + rb.Refs.Decision.InputsDigest + " "
			}
		}
		v.Refs = strings.TrimSpace(v.Refs)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = verifyPageTmpl.Execute(w, v)
}

func (h *Handler) PackPublic(w http.ResponseWriter, r *http.Request) {
	if !h.PublicVerify {
		http.NotFound(w, r)
		return
	}
	receiptID := strings.TrimPrefix(r.URL.Path, "/pack/")
	if receiptID == "" {
		http.NotFound(w, r)
		return
	}

	if h.AuthorizeService == nil || h.AuthorizeService.Ledger == nil {
		http.NotFound(w, r)
		return
	}

	receiptRec, ok := h.AuthorizeService.Ledger.GetReceipt(receiptID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	ctxRec, ok := h.AuthorizeService.Ledger.GetContext(receiptRec.ContextID)
	if !ok {
		http.NotFound(w, r)
		return
	}
	decisionRec, ok := h.AuthorizeService.Ledger.GetDecision(receiptRec.DecisionID)
	if !ok {
		http.NotFound(w, r)
		return
	}
	policyVersion, ok := h.AuthorizeService.Ledger.GetPolicyVersion(receiptRec.PolicyHash)
	if !ok {
		http.NotFound(w, r)
		return
	}

	var ctx types.ContextRecord
	if err := json.Unmarshal(ctxRec.BodyJSON, &ctx); err != nil {
		http.Error(w, "invalid stored context", http.StatusInternalServerError)
		return
	}
	var dec types.DecisionRecord
	if err := json.Unmarshal(decisionRec.BodyJSON, &dec); err != nil {
		http.Error(w, "invalid stored decision", http.StatusInternalServerError)
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

	storedReceipt := ledger.StoredReceipt{
		ReceiptID:  receiptRec.ReceiptID,
		BodyDigest: receiptRec.BodyDigest,
		BodyJSON:   receiptRec.BodyJSON,
		KeyID:      receiptRec.KeyID,
		Sig:        receiptRec.Sig,
		PolicyHash: receiptRec.PolicyHash,
		ApprovalID: receiptRec.ApprovalID,
		ExpiresAt:  receiptRec.ExpiresAt,
	}

	zipBytes, err := pack.BuildZip(pack.Input{
		Receipt:   storedReceipt,
		Context:   ctx,
		Decision:  dec,
		Policy:    []byte(policyVersion.PolicyYAML),
		Approvals: approvals,
		CreatedAt: receiptRec.CreatedAt,
	}, baseURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=relia-pack-"+receiptID+".zip")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(zipBytes)
}

func int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}
