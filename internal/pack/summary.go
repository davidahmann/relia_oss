package pack

import (
	"bytes"
	"encoding/json"
	"html/template"
	"strings"

	"github.com/davidahmann/relia/internal/grade"
	"github.com/davidahmann/relia/pkg/types"
)

type Summary struct {
	Schema         string                `json:"schema"`
	ReceiptID      string                `json:"receipt_id"`
	DecisionID     string                `json:"decision_id"`
	ContextID      string                `json:"context_id"`
	Verdict        string                `json:"verdict"`
	Grade          string                `json:"grade"`
	Refs           *types.ReceiptRefs    `json:"refs,omitempty"`
	InteractionRef *types.InteractionRef `json:"interaction_ref,omitempty"`
	ApprovalID     string                `json:"approval_id,omitempty"`
	ApprovalState  string                `json:"approval_status,omitempty"`
	PolicyID       string                `json:"policy_id,omitempty"`
	PolicyVersion  string                `json:"policy_version,omitempty"`
	PolicyHash     string                `json:"policy_hash"`
	RoleARN        string                `json:"role_arn,omitempty"`
	TTLSeconds     int64                 `json:"ttl_seconds,omitempty"`
	PlanDigest     string                `json:"plan_digest,omitempty"`
	DiffURL        string                `json:"diff_url,omitempty"`
	VerifyURL      string                `json:"verify_url,omitempty"`
	PackURL        string                `json:"pack_url,omitempty"`
}

const SummarySchema = "relia.pack_summary.v0.1"

func BuildSummary(input Input, baseURL string) (Summary, []byte, error) {
	quality := grade.Evaluate(grade.Input{
		Valid:    true,
		Receipt:  input.Receipt,
		Context:  &input.Context,
		Decision: &input.Decision,
	})

	var rb struct {
		Approval        *types.ReceiptApproval        `json:"approval,omitempty"`
		Refs            *types.ReceiptRefs            `json:"refs,omitempty"`
		InteractionRef  *types.InteractionRef         `json:"interaction_ref,omitempty"`
		CredentialGrant *types.ReceiptCredentialGrant `json:"credential_grant,omitempty"`
	}
	_ = json.Unmarshal(input.Receipt.BodyJSON, &rb)

	s := Summary{
		Schema:         SummarySchema,
		ReceiptID:      input.Receipt.ReceiptID,
		DecisionID:     input.Decision.DecisionID,
		ContextID:      input.Context.ContextID,
		Verdict:        input.Decision.Verdict,
		Grade:          quality.Grade,
		Refs:           rb.Refs,
		InteractionRef: rb.InteractionRef,
		PolicyID:       input.Decision.Policy.PolicyID,
		PolicyVersion:  input.Decision.Policy.PolicyVersion,
		PolicyHash:     input.Decision.Policy.PolicyHash,
		PlanDigest:     input.Context.Evidence.PlanDigest,
		DiffURL:        input.Context.Evidence.DiffURL,
	}

	if rb.CredentialGrant != nil {
		s.RoleARN = rb.CredentialGrant.RoleARN
		s.TTLSeconds = rb.CredentialGrant.TTLSeconds
	}
	if rb.Approval != nil && rb.Approval.Required {
		s.ApprovalID = rb.Approval.ApprovalID
		s.ApprovalState = rb.Approval.Status
	}

	if baseURL != "" {
		base := strings.TrimRight(baseURL, "/")
		s.VerifyURL = base + "/verify/" + input.Receipt.ReceiptID
		s.PackURL = base + "/pack/" + input.Receipt.ReceiptID
	}

	htmlBytes, err := renderSummaryHTML(s)
	if err != nil {
		return Summary{}, nil, err
	}
	return s, htmlBytes, nil
}

var summaryHTMLTmpl = template.Must(template.New("summary").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>Relia Pack Summary</title>
  <style>
    body{font-family:ui-sans-serif,system-ui,-apple-system,Segoe UI,Roboto,Helvetica,Arial; margin:24px; color:#0f172a}
    .card{max-width:920px; border:1px solid #e2e8f0; border-radius:12px; padding:18px 18px; box-shadow:0 1px 2px rgba(0,0,0,.04)}
    .row{display:flex; flex-wrap:wrap; gap:12px}
    .pill{display:inline-block; padding:4px 10px; border-radius:999px; font-size:12px; background:#f1f5f9}
    code{background:#f1f5f9; padding:2px 6px; border-radius:6px}
    .k{width:220px; font-size:12px; color:#475569}
    .v{font-size:13px}
    .kv{display:flex; gap:12px; padding:6px 0; border-bottom:1px dashed #e2e8f0}
    .kv:last-child{border-bottom:none}
    a{color:#2563eb; text-decoration:none}
    a:hover{text-decoration:underline}
  </style>
</head>
<body>
  <div class="card">
    <div class="row" style="margin:0 0 12px 0">
      <span class="pill">Grade: {{.Grade}}</span>
      <span class="pill">Verdict: {{.Verdict}}</span>
      <span class="pill">Receipt: <code>{{.ReceiptID}}</code></span>
    </div>
    <div class="kv"><div class="k">Policy</div><div class="v">{{.PolicyID}}@{{.PolicyVersion}} <code>{{.PolicyHash}}</code></div></div>
    <div class="kv"><div class="k">Approval</div><div class="v">{{if .ApprovalID}}{{.ApprovalState}} <code>{{.ApprovalID}}</code>{{else}}not required{{end}}</div></div>
    <div class="kv"><div class="k">Role / TTL</div><div class="v">{{if .RoleARN}}{{.RoleARN}} (ttl {{.TTLSeconds}}s){{else}}n/a{{end}}</div></div>
    <div class="kv"><div class="k">Interaction</div><div class="v">{{if .InteractionRef}}<code>{{.InteractionRef.Mode}}</code>{{if .InteractionRef.CallID}} call_id=<code>{{.InteractionRef.CallID}}</code>{{end}}{{if .InteractionRef.TurnID}} turn_id=<code>{{.InteractionRef.TurnID}}</code>{{end}}{{if .InteractionRef.TurnIndex}} turn_index=<code>{{.InteractionRef.TurnIndex}}</code>{{end}}{{else}}n/a{{end}}</div></div>
    <div class="kv"><div class="k">Plan Digest</div><div class="v">{{if .PlanDigest}}<code>{{.PlanDigest}}</code>{{else}}n/a{{end}}</div></div>
    <div class="kv"><div class="k">Diff URL</div><div class="v">{{if .DiffURL}}<a href="{{.DiffURL}}">{{.DiffURL}}</a>{{else}}n/a{{end}}</div></div>
    <div class="kv"><div class="k">Verify / Pack</div><div class="v">{{if .VerifyURL}}<a href="{{.VerifyURL}}">{{.VerifyURL}}</a>{{end}} {{if .PackURL}}<br/><a href="{{.PackURL}}">{{.PackURL}}</a>{{end}}</div></div>
  </div>
</body>
</html>`))

func renderSummaryHTML(s Summary) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := summaryHTMLTmpl.Execute(buf, s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
