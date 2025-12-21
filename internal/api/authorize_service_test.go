package api

import "testing"

func TestAuthorizeRequireApproval(t *testing.T) {
	svc, err := NewAuthorizeService("../../policies/relia.yaml")
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-prod",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	req := AuthorizeRequest{
		Action:   "terraform.apply",
		Resource: "aws:account:123456789012:stack/prod",
		Env:      "prod",
	}

	resp, err := svc.Authorize(claims, req, "2025-12-20T16:34:14Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}

	if resp.Verdict != string(VerdictRequireApproval) {
		t.Fatalf("expected require_approval, got %s", resp.Verdict)
	}
	if resp.Approval == nil || resp.Approval.ApprovalID == "" {
		t.Fatalf("expected approval id")
	}
}

func TestAuthorizeAllow(t *testing.T) {
	svc, err := NewAuthorizeService("../../policies/relia.yaml")
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-dev",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	req := AuthorizeRequest{
		Action:   "terraform.apply",
		Resource: "aws:account:123456789012:stack/dev",
		Env:      "dev",
	}

	resp, err := svc.Authorize(claims, req, "2025-12-20T16:34:14Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}

	if resp.Verdict != string(VerdictAllow) {
		t.Fatalf("expected allow, got %s", resp.Verdict)
	}
	if resp.ReceiptID == "" {
		t.Fatalf("expected receipt id")
	}
	if resp.AWSCredentials == nil || resp.AWSCredentials.AccessKeyID == "" {
		t.Fatalf("expected aws credentials")
	}
}

func TestAuthorizeIdempotentPending(t *testing.T) {
	svc, err := NewAuthorizeService("../../policies/relia.yaml")
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-prod",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	req := AuthorizeRequest{
		Action:   "terraform.apply",
		Resource: "aws:account:123456789012:stack/prod",
		Env:      "prod",
	}

	first, err := svc.Authorize(claims, req, "2025-12-20T16:34:14Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	second, err := svc.Authorize(claims, req, "2025-12-20T16:34:15Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}

	if first.Approval == nil || second.Approval == nil {
		t.Fatalf("expected approval in both responses")
	}
	if first.Approval.ApprovalID != second.Approval.ApprovalID {
		t.Fatalf("expected same approval id for idempotent request")
	}
}

func TestAuthorizeIdempotentAllow(t *testing.T) {
	svc, err := NewAuthorizeService("../../policies/relia.yaml")
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-dev",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	req := AuthorizeRequest{
		Action:   "terraform.apply",
		Resource: "aws:account:123456789012:stack/dev",
		Env:      "dev",
	}

	first, err := svc.Authorize(claims, req, "2025-12-20T16:34:14Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	second, err := svc.Authorize(claims, req, "2025-12-20T16:34:15Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}

	if first.ReceiptID != second.ReceiptID {
		t.Fatalf("expected same receipt id for idempotent allow")
	}
}

func TestAuthorizeApproveFlow(t *testing.T) {
	svc, err := NewAuthorizeService("../../policies/relia.yaml")
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-prod",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	req := AuthorizeRequest{
		Action:   "terraform.apply",
		Resource: "aws:account:123456789012:stack/prod",
		Env:      "prod",
	}

	pending, err := svc.Authorize(claims, req, "2025-12-20T16:34:14Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if pending.Approval == nil {
		t.Fatalf("expected approval")
	}

	receiptID, err := svc.Approve(pending.Approval.ApprovalID, ApprovalApproved, "2025-12-20T16:35:00Z")
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if receiptID == "" {
		t.Fatalf("expected approval receipt id")
	}

	allowed, err := svc.Authorize(claims, req, "2025-12-20T16:35:02Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if allowed.Verdict != string(VerdictAllow) {
		t.Fatalf("expected allow, got %s", allowed.Verdict)
	}
	if allowed.ReceiptID == pending.ReceiptID {
		t.Fatalf("expected final receipt to differ from pending receipt")
	}
}

func TestAuthorizeApproveDenied(t *testing.T) {
	svc, err := NewAuthorizeService("../../policies/relia.yaml")
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-prod",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	req := AuthorizeRequest{
		Action:   "terraform.apply",
		Resource: "aws:account:123456789012:stack/prod",
		Env:      "prod",
	}

	pending, err := svc.Authorize(claims, req, "2025-12-20T16:34:14Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if pending.Approval == nil {
		t.Fatalf("expected approval")
	}

	_, err = svc.Approve(pending.Approval.ApprovalID, ApprovalDenied, "2025-12-20T16:35:00Z")
	if err != nil {
		t.Fatalf("approve: %v", err)
	}

	denied, err := svc.Authorize(claims, req, "2025-12-20T16:35:02Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if denied.Verdict != string(VerdictDeny) {
		t.Fatalf("expected deny, got %s", denied.Verdict)
	}
}

func TestApproveNotFound(t *testing.T) {
	svc, err := NewAuthorizeService("../../policies/relia.yaml")
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	_, err = svc.Approve("missing", ApprovalApproved, "2025-12-20T16:35:00Z")
	if err == nil {
		t.Fatalf("expected error for missing approval")
	}
}

func TestApproveInvalidStatus(t *testing.T) {
	svc, err := NewAuthorizeService("../../policies/relia.yaml")
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "terraform-prod",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	pending, err := svc.Authorize(claims, AuthorizeRequest{Action: "terraform.apply", Resource: "res", Env: "prod"}, "2025-12-20T16:34:14Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if pending.Approval == nil {
		t.Fatalf("expected approval")
	}

	_, err = svc.Approve(pending.Approval.ApprovalID, ApprovalStatus("bad"), "2025-12-20T16:35:00Z")
	if err == nil {
		t.Fatalf("expected error for invalid status")
	}
}

func TestAuthorizeDenyUnknownProdAction(t *testing.T) {
	svc, err := NewAuthorizeService("../../policies/relia.yaml")
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	claims := ActorContext{
		Subject:  "repo:org/repo:ref:refs/heads/main",
		Issuer:   "relia-dev",
		Repo:     "org/repo",
		Workflow: "deploy-prod",
		RunID:    "123456",
		SHA:      "abcdef123",
	}

	req := AuthorizeRequest{
		Action:   "deploy",
		Resource: "aws:account:123456789012:stack/prod",
		Env:      "prod",
	}

	resp, err := svc.Authorize(claims, req, "2025-12-20T16:34:14Z")
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if resp.Verdict != string(VerdictDeny) {
		t.Fatalf("expected deny, got %s", resp.Verdict)
	}
}

func TestAuthorizeMissingActorFields(t *testing.T) {
	svc, err := NewAuthorizeService("../../policies/relia.yaml")
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	_, err = svc.Authorize(ActorContext{}, AuthorizeRequest{Action: "a", Resource: "r", Env: "e"}, "2025-12-20T16:34:14Z")
	if err == nil {
		t.Fatalf("expected error for missing actor fields")
	}
}

func TestAuthorizeMissingPolicy(t *testing.T) {
	svc := &AuthorizeService{PolicyPath: "missing.yaml", Store: NewInMemoryIdemStore(), Signer: devSigner{keyID: "dev"}}

	_, err := svc.Authorize(ActorContext{Subject: "sub", Issuer: "iss", Repo: "repo", RunID: "run"}, AuthorizeRequest{Action: "a", Resource: "r", Env: "e"}, "2025-12-20T16:34:14Z")
	if err == nil {
		t.Fatalf("expected error for missing policy")
	}
}
