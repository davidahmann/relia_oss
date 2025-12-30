package types

type PackManifest struct {
	Schema         string          `json:"schema"`
	CreatedAt      string          `json:"created_at"`
	ReceiptID      string          `json:"receipt_id"`
	ContextID      string          `json:"context_id"`
	DecisionID     string          `json:"decision_id"`
	PolicyHash     string          `json:"policy_hash"`
	ApprovalID     string          `json:"approval_id,omitempty"`
	Refs           *ReceiptRefs    `json:"refs,omitempty"`
	InteractionRef *InteractionRef `json:"interaction_ref,omitempty"`
	Schemas        PackSchemas     `json:"schemas"`
	Files          []PackFile      `json:"files"`
}

type PackSchemas struct {
	Context  string `json:"context"`
	Decision string `json:"decision"`
	Receipt  string `json:"receipt"`
}

type PackFile struct {
	Name        string `json:"name"`
	SHA256      string `json:"sha256"`
	SizeBytes   int64  `json:"size_bytes"`
	ContentType string `json:"content_type,omitempty"`
}
