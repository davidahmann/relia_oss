package ledger

type Store interface {
	WithTx(fn func(Tx) error) error

	PutContext(ctx ContextRecord) error
	PutDecision(decision DecisionRecord) error
	PutReceipt(receipt ReceiptRecord) error
	PutApproval(approval ApprovalRecord) error
	PutIdempotencyKey(key IdempotencyKey) error
}

type Tx interface {
	PutContext(ctx ContextRecord) error
	PutDecision(decision DecisionRecord) error
	PutReceipt(receipt ReceiptRecord) error
	PutApproval(approval ApprovalRecord) error
	PutIdempotencyKey(key IdempotencyKey) error
}

type ContextRecord struct {
	ContextID string
	BodyJSON  []byte
	CreatedAt string
}

type DecisionRecord struct {
	DecisionID string
	ContextID  string
	PolicyHash string
	Verdict    string
	BodyJSON   []byte
	CreatedAt  string
}

type ReceiptRecord struct {
	ReceiptID           string
	IdemKey             string
	CreatedAt           string
	SupersedesReceiptID *string
	ContextID           string
	DecisionID          string
	PolicyHash          string
	ApprovalID          *string
	OutcomeStatus       string
	Final               bool
	ExpiresAt           *string
	BodyJSON            []byte
	BodyDigest          string
	KeyID               string
	Sig                 []byte
}

type ApprovalRecord struct {
	ApprovalID   string
	IdemKey      string
	Status       string
	SlackChannel *string
	SlackMsgTS   *string
	ApprovedBy   *string
	ApprovedAt   *string
	CreatedAt    string
	UpdatedAt    string
}

type IdempotencyKey struct {
	IdemKey         string
	Status          string
	ApprovalID      *string
	LatestReceiptID *string
	FinalReceiptID  *string
	CreatedAt       string
	UpdatedAt       string
	TTLExpiresAt    *string
}
