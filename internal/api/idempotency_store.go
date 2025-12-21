package api

import "sync"

type IdemRecord struct {
	IdemKey         string
	Status          IdemStatus
	ApprovalID      string
	LatestReceiptID string
	FinalReceiptID  string
	ContextID       string
	DecisionID      string
	PolicyHash      string
	RoleARN         string
	TTLSeconds      int
}

type ApprovalRecord struct {
	ApprovalID string
	Status     ApprovalStatus
	ReceiptID  string
}

type InMemoryIdemStore struct {
	mu        sync.Mutex
	items     map[string]IdemRecord
	approvals map[string]ApprovalRecord
}

func NewInMemoryIdemStore() *InMemoryIdemStore {
	return &InMemoryIdemStore{
		items:     make(map[string]IdemRecord),
		approvals: make(map[string]ApprovalRecord),
	}
}

func (s *InMemoryIdemStore) Get(idemKey string) (IdemRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, ok := s.items[idemKey]
	return rec, ok
}

func (s *InMemoryIdemStore) Put(record IdemRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[record.IdemKey] = record
}

func (s *InMemoryIdemStore) GetApproval(approvalID string) (ApprovalRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, ok := s.approvals[approvalID]
	return rec, ok
}

func (s *InMemoryIdemStore) PutApproval(record ApprovalRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.approvals[record.ApprovalID] = record
}
