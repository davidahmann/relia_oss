package slack

import "testing"

func TestBuildApprovalMessage(t *testing.T) {
	payload, err := BuildApprovalMessage(ApprovalMessageInput{
		ApprovalID: "appr-1",
		Action:     "terraform.apply",
		Resource:   "aws:stack/prod",
		Env:        "prod",
		Risk:       "high",
		DiffURL:    "https://example.com/diff",
	})
	if err != nil {
		t.Fatalf("message: %v", err)
	}
	if len(payload) == 0 {
		t.Fatalf("expected payload")
	}
}

func TestBuildApprovalMessageMinimal(t *testing.T) {
	payload, err := BuildApprovalMessage(ApprovalMessageInput{
		ApprovalID: "appr-2",
		Action:     "deploy",
		Resource:   "res",
		Env:        "prod",
	})
	if err != nil {
		t.Fatalf("message: %v", err)
	}
	if len(payload) == 0 {
		t.Fatalf("expected payload")
	}
}
