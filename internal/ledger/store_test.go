package ledger

import "testing"

func TestInMemoryStoreMethods(t *testing.T) {
	store := &InMemoryStore{}

	err := store.WithTx(func(tx Tx) error {
		if err := tx.PutContext(ContextRecord{ContextID: "ctx", CreatedAt: "now"}); err != nil {
			return err
		}
		if err := tx.PutDecision(DecisionRecord{DecisionID: "dec", ContextID: "ctx"}); err != nil {
			return err
		}
		if err := tx.PutReceipt(ReceiptRecord{ReceiptID: "rec", IdemKey: "idem"}); err != nil {
			return err
		}
		if err := tx.PutApproval(ApprovalRecord{ApprovalID: "app", IdemKey: "idem"}); err != nil {
			return err
		}
		if err := tx.PutIdempotencyKey(IdempotencyKey{IdemKey: "idem", Status: "pending_approval"}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("with tx: %v", err)
	}

	if err := store.PutContext(ContextRecord{ContextID: "ctx", CreatedAt: "now"}); err != nil {
		t.Fatalf("put context: %v", err)
	}
	if err := store.PutDecision(DecisionRecord{DecisionID: "dec", ContextID: "ctx"}); err != nil {
		t.Fatalf("put decision: %v", err)
	}
	if err := store.PutReceipt(ReceiptRecord{ReceiptID: "rec", IdemKey: "idem"}); err != nil {
		t.Fatalf("put receipt: %v", err)
	}
	if err := store.PutApproval(ApprovalRecord{ApprovalID: "app", IdemKey: "idem"}); err != nil {
		t.Fatalf("put approval: %v", err)
	}
	if err := store.PutIdempotencyKey(IdempotencyKey{IdemKey: "idem", Status: "pending_approval"}); err != nil {
		t.Fatalf("put idempotency: %v", err)
	}
}
