package ledger

import (
	"fmt"

	"github.com/davidahmann/relia/internal/crypto"
	"github.com/davidahmann/relia/pkg/types"
)

const ReceiptSchema = "relia.receipt.v0.1"

type Signer interface {
	KeyID() string
	SignEd25519(message []byte) ([]byte, error)
}

type MakeReceiptInput struct {
	Schema    string
	CreatedAt string

	IdemKey             string
	SupersedesReceiptID *string

	ContextID  string
	DecisionID string

	Actor   types.ReceiptActor
	Request types.ReceiptRequest
	Policy  types.ReceiptPolicy

	InteractionRef  *types.InteractionRef
	Refs            *types.ReceiptRefs
	Approval        *types.ReceiptApproval
	CredentialGrant *types.ReceiptCredentialGrant
	Outcome         types.ReceiptOutcome
}

type StoredReceipt struct {
	ReceiptID  string
	BodyDigest string
	BodyJSON   []byte
	KeyID      string
	Sig        []byte

	IdemKey             string
	CreatedAt           string
	SupersedesReceiptID *string
	ContextID           string
	DecisionID          string
	OutcomeStatus       types.OutcomeStatus
	ApprovalID          *string
	PolicyHash          string
	Final               bool
	ExpiresAt           *string
}

// MakeReceipt canonicalizes + hashes + signs a receipt body.
func MakeReceipt(in MakeReceiptInput, signer Signer) (StoredReceipt, error) {
	if in.Schema == "" {
		in.Schema = ReceiptSchema
	}
	if in.Schema != ReceiptSchema {
		return StoredReceipt{}, fmt.Errorf("invalid schema: %s", in.Schema)
	}
	if in.IdemKey == "" || in.ContextID == "" || in.DecisionID == "" || in.Policy.PolicyHash == "" {
		return StoredReceipt{}, fmt.Errorf("missing required receipt fields")
	}
	if !validOutcome(in.Outcome.Status) {
		return StoredReceipt{}, fmt.Errorf("invalid outcome status: %s", in.Outcome.Status)
	}

	approval := approvalMap(in.Approval)
	credential := credentialMap(in.CredentialGrant)
	outcomeError := outcomeErrorMap(in.Outcome.Error)

	body := map[string]any{
		"schema":      in.Schema,
		"created_at":  in.CreatedAt,
		"context_id":  in.ContextID,
		"decision_id": in.DecisionID,
		"actor": map[string]any{
			"kind":     in.Actor.Kind,
			"subject":  in.Actor.Subject,
			"issuer":   in.Actor.Issuer,
			"repo":     in.Actor.Repo,
			"workflow": in.Actor.Workflow,
			"run_id":   in.Actor.RunID,
			"sha":      in.Actor.SHA,
		},
		"request": map[string]any{
			"request_id": in.Request.RequestID,
			"action":     in.Request.Action,
			"resource":   in.Request.Resource,
			"env":        in.Request.Env,
			"intent":     in.Request.Intent,
		},
		"policy": map[string]any{
			"policy_id":      in.Policy.PolicyID,
			"policy_version": in.Policy.PolicyVersion,
			"policy_hash":    in.Policy.PolicyHash,
		},
		"approval":         approval,
		"credential_grant": credential,
		"outcome": map[string]any{
			"status":     in.Outcome.Status,
			"issued_at":  in.Outcome.IssuedAt,
			"expires_at": in.Outcome.ExpiresAt,
			"error":      outcomeError,
		},
	}
	if ir := interactionRefMap(in.InteractionRef); ir != nil {
		body["interaction_ref"] = ir
	}
	if refs := refsMap(in.Refs); refs != nil {
		body["refs"] = refs
	}

	canonical, err := crypto.Canonicalize(body)
	if err != nil {
		return StoredReceipt{}, err
	}

	digestBytes := crypto.DigestBytes(canonical)
	bodyDigest := crypto.DigestWithPrefix(canonical)

	sig, err := signer.SignEd25519(digestBytes)
	if err != nil {
		return StoredReceipt{}, err
	}

	final := isFinalOutcome(in.Outcome.Status)

	var approvalID *string
	if in.Approval != nil && in.Approval.ApprovalID != "" {
		approvalID = &in.Approval.ApprovalID
	}

	var expiresAt *string
	if in.Outcome.ExpiresAt != "" {
		expiresAt = &in.Outcome.ExpiresAt
	}

	return StoredReceipt{
		ReceiptID:           bodyDigest,
		BodyDigest:          bodyDigest,
		BodyJSON:            canonical,
		KeyID:               signer.KeyID(),
		Sig:                 sig,
		IdemKey:             in.IdemKey,
		CreatedAt:           in.CreatedAt,
		SupersedesReceiptID: in.SupersedesReceiptID,
		ContextID:           in.ContextID,
		DecisionID:          in.DecisionID,
		OutcomeStatus:       in.Outcome.Status,
		ApprovalID:          approvalID,
		PolicyHash:          in.Policy.PolicyHash,
		Final:               final,
		ExpiresAt:           expiresAt,
	}, nil
}

func approvalMap(approval *types.ReceiptApproval) map[string]any {
	if approval == nil {
		return nil
	}

	var approver map[string]any
	if approval.Approver != nil {
		approver = map[string]any{
			"kind":    approval.Approver.Kind,
			"id":      approval.Approver.ID,
			"display": approval.Approver.Display,
		}
	}

	return map[string]any{
		"required":    approval.Required,
		"approval_id": emptyToNil(approval.ApprovalID),
		"status":      emptyToNil(approval.Status),
		"approved_at": emptyToNil(approval.ApprovedAt),
		"approver":    approver,
	}
}

func refsMap(refs *types.ReceiptRefs) map[string]any {
	if refs == nil {
		return nil
	}

	var ctx map[string]any
	if refs.Context != nil {
		ctx = map[string]any{
			"context_id":   emptyToNil(refs.Context.ContextID),
			"record_hash":  emptyToNil(refs.Context.RecordHash),
			"content_hash": emptyToNil(refs.Context.ContentHash),
		}
	}

	var decision map[string]any
	if refs.Decision != nil {
		decision = map[string]any{
			"decision_id":    emptyToNil(refs.Decision.DecisionID),
			"inputs_digest":  emptyToNil(refs.Decision.InputsDigest),
			"record_hash":    emptyToNil(refs.Decision.RecordHash),
			"content_digest": emptyToNil(refs.Decision.ContentDigest),
		}
	}

	if ctx == nil && decision == nil {
		return nil
	}

	return map[string]any{
		"context":  ctx,
		"decision": decision,
	}
}

func interactionRefMap(ref *types.InteractionRef) map[string]any {
	if ref == nil {
		return nil
	}

	// Keep this minimal and pass-through: Relia does not validate semantics, but it
	// does avoid emitting an empty object.
	m := map[string]any{
		"mode":            emptyToNil(ref.Mode),
		"session_id":      emptyToNil(ref.SessionID),
		"call_id":         emptyToNil(ref.CallID),
		"turn_id":         emptyToNil(ref.TurnID),
		"turn_started_at": emptyToNil(ref.TurnStartedAt),
		"turn_ended_at":   emptyToNil(ref.TurnEndedAt),
		"jurisdiction":    emptyToNil(ref.Jurisdiction),
		"consent_state":   emptyToNil(ref.ConsentState),
		"redaction_mode":  emptyToNil(ref.RedactionMode),
	}
	if ref.TurnIndex != 0 {
		m["turn_index"] = ref.TurnIndex
	}

	empty := true
	for _, v := range m {
		if v != nil {
			empty = false
			break
		}
	}
	if empty {
		return nil
	}
	return m
}

func credentialMap(credential *types.ReceiptCredentialGrant) map[string]any {
	if credential == nil {
		return nil
	}

	return map[string]any{
		"provider":     emptyToNil(credential.Provider),
		"method":       emptyToNil(credential.Method),
		"role_arn":     emptyToNil(credential.RoleARN),
		"region":       emptyToNil(credential.Region),
		"ttl_seconds":  credential.TTLSeconds,
		"scope_digest": emptyToNil(credential.ScopeDigest),
	}
}

func outcomeErrorMap(errInfo *struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}) map[string]any {
	if errInfo == nil {
		return nil
	}
	return map[string]any{
		"code": emptyToNil(errInfo.Code),
		"msg":  emptyToNil(errInfo.Msg),
	}
}

func emptyToNil(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func validOutcome(status types.OutcomeStatus) bool {
	switch status {
	case types.OutcomeApprovalPending,
		types.OutcomeApprovalApproved,
		types.OutcomeApprovalDenied,
		types.OutcomeIssuingCredentials,
		types.OutcomeIssuedCredentials,
		types.OutcomeDenied,
		types.OutcomeIssueFailed:
		return true
	default:
		return false
	}
}

func isFinalOutcome(status types.OutcomeStatus) bool {
	switch status {
	case types.OutcomeIssuedCredentials, types.OutcomeDenied, types.OutcomeIssueFailed:
		return true
	default:
		return false
	}
}
