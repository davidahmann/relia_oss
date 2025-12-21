package ledger

// InMemoryStore is a placeholder for future implementations.
type InMemoryStore struct{}

func (s *InMemoryStore) WithTx(fn func(Tx) error) error {
	return fn(s)
}

func (s *InMemoryStore) PutContext(ctx ContextRecord) error {
	return nil
}

func (s *InMemoryStore) PutDecision(decision DecisionRecord) error {
	return nil
}

func (s *InMemoryStore) PutReceipt(receipt ReceiptRecord) error {
	return nil
}

func (s *InMemoryStore) PutApproval(approval ApprovalRecord) error {
	return nil
}

func (s *InMemoryStore) PutIdempotencyKey(key IdempotencyKey) error {
	return nil
}
