package api

import (
	"fmt"

	"github.com/davidahmann/relia/internal/crypto"
	"github.com/davidahmann/relia/pkg/types"
)

type AuthorizeRequest struct {
	RequestID      string                `json:"request_id,omitempty"`
	Action         string                `json:"action"`
	Resource       string                `json:"resource"`
	Env            string                `json:"env"`
	Intent         map[string]any        `json:"intent,omitempty"`
	Evidence       AuthorizeEvidence     `json:"evidence,omitempty"`
	AWS            *AuthorizeAWS         `json:"aws,omitempty"`
	InteractionRef *types.InteractionRef `json:"interaction_ref,omitempty"`
	ContextRef     *types.ContextRef     `json:"context_ref,omitempty"`
	DecisionRef    *types.DecisionRef    `json:"decision_ref,omitempty"`
}

type AuthorizeEvidence struct {
	PlanDigest string `json:"plan_digest,omitempty"`
	DiffURL    string `json:"diff_url,omitempty"`
}

type AuthorizeAWS struct {
	Region string `json:"region,omitempty"`
}

type ActorContext struct {
	Subject  string
	Issuer   string
	Repo     string
	Workflow string
	RunID    string
	SHA      string
	Token    string
}

// ComputeIdemKey derives a deterministic idempotency key from actor + request.
func ComputeIdemKey(actor ActorContext, req AuthorizeRequest) (string, error) {
	if req.Action == "" || req.Resource == "" || req.Env == "" {
		return "", fmt.Errorf("missing required action/resource/env")
	}
	if actor.Subject == "" || actor.Issuer == "" || actor.Repo == "" || actor.RunID == "" {
		return "", fmt.Errorf("missing required actor identity")
	}

	intent := req.Intent
	if intent == nil {
		intent = map[string]any{}
	}

	payload := map[string]any{
		"schema":      "relia.idem.v1",
		"iss":         actor.Issuer,
		"sub":         actor.Subject,
		"repo":        actor.Repo,
		"workflow":    actor.Workflow,
		"run_id":      actor.RunID,
		"sha":         actor.SHA,
		"action":      req.Action,
		"resource":    req.Resource,
		"env":         req.Env,
		"intent":      intent,
		"plan_digest": req.Evidence.PlanDigest,
	}
	if req.ContextRef != nil {
		payload["context_ref"] = map[string]any{
			"context_id":   req.ContextRef.ContextID,
			"record_hash":  req.ContextRef.RecordHash,
			"content_hash": req.ContextRef.ContentHash,
		}
	}
	if req.DecisionRef != nil {
		payload["decision_ref"] = map[string]any{
			"decision_id":    req.DecisionRef.DecisionID,
			"inputs_digest":  req.DecisionRef.InputsDigest,
			"record_hash":    req.DecisionRef.RecordHash,
			"content_digest": req.DecisionRef.ContentDigest,
		}
	}
	if req.InteractionRef != nil {
		payload["interaction_ref"] = map[string]any{
			"mode":            req.InteractionRef.Mode,
			"session_id":      req.InteractionRef.SessionID,
			"call_id":         req.InteractionRef.CallID,
			"turn_id":         req.InteractionRef.TurnID,
			"turn_index":      req.InteractionRef.TurnIndex,
			"turn_started_at": req.InteractionRef.TurnStartedAt,
			"turn_ended_at":   req.InteractionRef.TurnEndedAt,
			"jurisdiction":    req.InteractionRef.Jurisdiction,
			"consent_state":   req.InteractionRef.ConsentState,
			"redaction_mode":  req.InteractionRef.RedactionMode,
		}
	}

	if req.RequestID != "" {
		payload["request_id"] = req.RequestID
	}

	canonical, err := crypto.Canonicalize(payload)
	if err != nil {
		return "", err
	}

	digest := crypto.DigestHex(canonical)
	return "idem:v1:sha256:" + digest, nil
}
