package types

// InteractionRef is a stable, pass-through handle for a conversational timeline (voice/chat).
//
// It is intended to be an append-only event log reference with a canonical turn ordering
// (turn_index is 1-based). Downstream systems should carry it unchanged when provided.
type InteractionRef struct {
	Mode          string `json:"mode,omitempty"` // "voice" | "chat" | other (caller-defined)
	SessionID     string `json:"session_id,omitempty"`
	CallID        string `json:"call_id,omitempty"`
	TurnID        string `json:"turn_id,omitempty"`
	TurnIndex     int    `json:"turn_index,omitempty"` // 1-based
	TurnStartedAt string `json:"turn_started_at,omitempty"`
	TurnEndedAt   string `json:"turn_ended_at,omitempty"`

	Jurisdiction  string `json:"jurisdiction,omitempty"`
	ConsentState  string `json:"consent_state,omitempty"`
	RedactionMode string `json:"redaction_mode,omitempty"`
}
