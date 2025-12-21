package api

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/davidahmann/relia/internal/aws"
	"github.com/davidahmann/relia/internal/context"
	"github.com/davidahmann/relia/internal/decision"
	"github.com/davidahmann/relia/internal/ledger"
	"github.com/davidahmann/relia/internal/policy"
	"github.com/davidahmann/relia/pkg/types"
)

type AuthorizeService struct {
	PolicyPath string
	Store      *InMemoryIdemStore
	Signer     ledger.Signer
	Broker     aws.CredentialBroker
}

type AuthorizeResponse struct {
	Verdict        string `json:"verdict"`
	ContextID      string `json:"context_id"`
	DecisionID     string `json:"decision_id"`
	ReceiptID      string `json:"receipt_id"`
	AWSCredentials *struct {
		AccessKeyID     string `json:"access_key_id"`
		SecretAccessKey string `json:"secret_access_key"`
		SessionToken    string `json:"session_token"`
		ExpiresAt       string `json:"expires_at"`
	} `json:"aws_credentials,omitempty"`
	Approval *struct {
		ApprovalID string `json:"approval_id"`
		Status     string `json:"status"`
	} `json:"approval,omitempty"`
	Error string `json:"error,omitempty"`
}

func NewAuthorizeService(policyPath string) (*AuthorizeService, error) {
	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		return nil, err
	}
	priv := ed25519.NewKeyFromSeed(seed)

	return &AuthorizeService{
		PolicyPath: policyPath,
		Store:      NewInMemoryIdemStore(),
		Signer:     devSigner{keyID: "dev", priv: priv},
		Broker:     aws.DevBroker{},
	}, nil
}

func (s *AuthorizeService) Authorize(claims ActorContext, req AuthorizeRequest, createdAt string) (AuthorizeResponse, error) {
	idemKey, err := ComputeIdemKey(claims, req)
	if err != nil {
		return AuthorizeResponse{}, err
	}

	if rec, ok := s.Store.Get(idemKey); ok {
		switch rec.Status {
		case IdemAllowed:
			return AuthorizeResponse{Verdict: string(VerdictAllow), ContextID: rec.ContextID, DecisionID: rec.DecisionID, ReceiptID: rec.FinalReceiptID}, nil
		case IdemDenied:
			return AuthorizeResponse{Verdict: string(VerdictDeny), ContextID: rec.ContextID, DecisionID: rec.DecisionID, ReceiptID: rec.FinalReceiptID}, nil
		case IdemPendingApproval:
			return AuthorizeResponse{Verdict: string(VerdictRequireApproval), ContextID: rec.ContextID, DecisionID: rec.DecisionID, ReceiptID: rec.LatestReceiptID, Approval: &struct {
				ApprovalID string `json:"approval_id"`
				Status     string `json:"status"`
			}{ApprovalID: rec.ApprovalID, Status: string(ApprovalPending)}}, nil
		case IdemApprovedReady:
			return s.issueCredentials(rec, claims, req, createdAt, true, rec.RoleARN, rec.TTLSeconds)
		case IdemIssuing:
			return s.issueCredentials(rec, claims, req, createdAt, false, rec.RoleARN, rec.TTLSeconds)
		case IdemErrored:
			return AuthorizeResponse{Verdict: string(VerdictDeny), Error: "previous error"}, nil
		}
	}

	loaded, err := policy.LoadPolicy(s.PolicyPath)
	if err != nil {
		return AuthorizeResponse{}, err
	}

	input := policy.Input{Action: req.Action, Resource: req.Resource, Env: req.Env}
	decisionResult := policy.Evaluate(loaded.Policy, loaded.Hash, input)

	source := types.ContextSource{
		Kind:     "github_actions",
		Repo:     claims.Repo,
		Workflow: claims.Workflow,
		RunID:    claims.RunID,
		Actor:    claims.Subject,
		Ref:      "",
		SHA:      claims.SHA,
	}
	inputs := types.ContextInputs{Action: req.Action, Resource: req.Resource, Env: req.Env, Intent: req.Intent}
	evidence := types.ContextEvidence{PlanDigest: req.Evidence.PlanDigest, DiffURL: req.Evidence.DiffURL}

	ctxRecord, err := context.BuildContext(source, inputs, evidence, createdAt)
	if err != nil {
		return AuthorizeResponse{}, err
	}

	policyMeta := types.DecisionPolicy{PolicyID: loaded.Policy.PolicyID, PolicyVersion: loaded.Policy.PolicyVersion, PolicyHash: loaded.Hash}
	decRecord, err := decision.BuildDecision(ctxRecord.ContextID, policyMeta, decisionResult.Verdict, decisionResult.ReasonCodes, decisionResult.RequireApproval, decisionResult.Risk, createdAt)
	if err != nil {
		return AuthorizeResponse{}, err
	}

	verdict := DecisionVerdict(decisionResult.Verdict)
	status, action := TransitionFromDecision(verdict)

	approvalID := ""
	var approval *types.ReceiptApproval
	if action == ActionReturnPending {
		approvalID = newApprovalID()
		approval = &types.ReceiptApproval{Required: true, ApprovalID: approvalID, Status: string(ApprovalPending)}
	}

	receiptPolicy := types.ReceiptPolicy(policyMeta)

	outcome := types.ReceiptOutcome{Status: types.OutcomeDenied}
	switch action {
	case ActionReturnDenied:
		outcome.Status = types.OutcomeDenied
	case ActionReturnPending:
		outcome.Status = types.OutcomeApprovalPending
	case ActionIssueCredentials:
		outcome.Status = types.OutcomeIssuingCredentials
		status = IdemIssuing
	default:
		outcome.Status = types.OutcomeIssueFailed
		status = IdemErrored
	}

	receipt, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
		CreatedAt:  createdAt,
		IdemKey:    idemKey,
		ContextID:  ctxRecord.ContextID,
		DecisionID: decRecord.DecisionID,
		Actor: types.ReceiptActor{
			Kind:     "workload",
			Subject:  claims.Subject,
			Issuer:   claims.Issuer,
			Repo:     claims.Repo,
			Workflow: claims.Workflow,
			RunID:    claims.RunID,
			SHA:      claims.SHA,
		},
		Request: types.ReceiptRequest{
			RequestID: req.RequestID,
			Action:    req.Action,
			Resource:  req.Resource,
			Env:       req.Env,
			Intent:    req.Intent,
		},
		Policy:   receiptPolicy,
		Approval: approval,
		Outcome:  outcome,
	}, s.Signer)
	if err != nil {
		return AuthorizeResponse{}, err
	}

	rec := IdemRecord{
		IdemKey:         idemKey,
		Status:          status,
		ApprovalID:      approvalID,
		LatestReceiptID: receipt.ReceiptID,
		FinalReceiptID:  receipt.ReceiptID,
		ContextID:       ctxRecord.ContextID,
		DecisionID:      decRecord.DecisionID,
		PolicyHash:      policyMeta.PolicyHash,
		RoleARN:         decisionResult.AWSRoleARN,
		TTLSeconds:      decisionResult.TTLSeconds,
	}

	if action == ActionReturnPending {
		rec.FinalReceiptID = ""
		s.Store.PutApproval(ApprovalRecord{ApprovalID: approvalID, Status: ApprovalPending, ReceiptID: receipt.ReceiptID})
	}

	s.Store.Put(rec)

	resp := AuthorizeResponse{
		Verdict:    string(verdict),
		ContextID:  ctxRecord.ContextID,
		DecisionID: decRecord.DecisionID,
		ReceiptID:  receipt.ReceiptID,
	}
	if approvalID != "" {
		resp.Verdict = string(VerdictRequireApproval)
		resp.Approval = &struct {
			ApprovalID string `json:"approval_id"`
			Status     string `json:"status"`
		}{ApprovalID: approvalID, Status: string(ApprovalPending)}
	}

	if action == ActionReturnDenied {
		resp.Verdict = string(VerdictDeny)
	}
	if action == ActionIssueCredentials {
		return s.issueCredentials(rec, claims, req, createdAt, false, decisionResult.AWSRoleARN, decisionResult.TTLSeconds)
	}

	return resp, nil
}

func (s *AuthorizeService) Approve(approvalID string, status string, createdAt string) (string, error) {
	approval, ok := s.Store.GetApproval(approvalID)
	if !ok {
		return "", fmt.Errorf("approval not found")
	}
	if approval.Status == ApprovalApproved || approval.Status == ApprovalDenied {
		return approval.ReceiptID, nil
	}
	approvalStatus := ApprovalStatus(status)
	if approvalStatus != ApprovalApproved && approvalStatus != ApprovalDenied {
		return "", fmt.Errorf("invalid approval status")
	}

	idem, ok := s.findIdemByApproval(approvalID)
	if !ok {
		return "", fmt.Errorf("idempotency not found for approval")
	}

	outcome := types.ReceiptOutcome{Status: types.OutcomeApprovalDenied}
	if approvalStatus == ApprovalApproved {
		outcome.Status = types.OutcomeApprovalApproved
	}

	approvalReceipt, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
		CreatedAt:           createdAt,
		IdemKey:             idem.IdemKey,
		SupersedesReceiptID: &idem.LatestReceiptID,
		ContextID:           idem.ContextID,
		DecisionID:          idem.DecisionID,
		Actor:               types.ReceiptActor{Kind: "approval", Subject: "slack"},
		Request:             types.ReceiptRequest{RequestID: "approval", Action: "approve", Resource: idem.IdemKey, Env: ""},
		Policy:              types.ReceiptPolicy{PolicyHash: idem.PolicyHash},
		Approval: &types.ReceiptApproval{
			Required:   true,
			ApprovalID: approvalID,
			Status:     string(approvalStatus),
		},
		Outcome: outcome,
	}, s.Signer)
	if err != nil {
		return "", err
	}

	approval.Status = approvalStatus
	approval.ReceiptID = approvalReceipt.ReceiptID
	s.Store.PutApproval(approval)

	idem.LatestReceiptID = approvalReceipt.ReceiptID
	if approvalStatus == ApprovalDenied {
		idem.Status = IdemDenied
		idem.FinalReceiptID = approvalReceipt.ReceiptID
		s.Store.Put(idem)
		return approvalReceipt.ReceiptID, nil
	}

	idem.Status = IdemApprovedReady
	idem.FinalReceiptID = ""
	s.Store.Put(idem)
	return approvalReceipt.ReceiptID, nil
}

func (s *AuthorizeService) GetApproval(approvalID string) (ApprovalRecord, bool) {
	return s.Store.GetApproval(approvalID)
}

func (s *AuthorizeService) issueCredentials(idem IdemRecord, claims ActorContext, req AuthorizeRequest, createdAt string, createIssuing bool, roleARN string, ttlSeconds int) (AuthorizeResponse, error) {
	latest := idem.LatestReceiptID
	if createIssuing {
		issuingReceipt, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
			CreatedAt:           createdAt,
			IdemKey:             idem.IdemKey,
			SupersedesReceiptID: &latest,
			ContextID:           idem.ContextID,
			DecisionID:          idem.DecisionID,
			Actor: types.ReceiptActor{
				Kind:     "workload",
				Subject:  claims.Subject,
				Issuer:   claims.Issuer,
				Repo:     claims.Repo,
				Workflow: claims.Workflow,
				RunID:    claims.RunID,
				SHA:      claims.SHA,
			},
			Request: types.ReceiptRequest{
				RequestID: req.RequestID,
				Action:    req.Action,
				Resource:  req.Resource,
				Env:       req.Env,
				Intent:    req.Intent,
			},
			Policy:  types.ReceiptPolicy{PolicyHash: idem.PolicyHash},
			Outcome: types.ReceiptOutcome{Status: types.OutcomeIssuingCredentials},
		}, s.Signer)
		if err != nil {
			return AuthorizeResponse{}, err
		}
		latest = issuingReceipt.ReceiptID
	}

	region := ""
	if req.AWS != nil {
		region = req.AWS.Region
	}

	credentialGrant := &types.ReceiptCredentialGrant{
		Provider:   "aws_sts",
		Method:     "AssumeRoleWithWebIdentity",
		RoleARN:    roleARN,
		Region:     region,
		TTLSeconds: int64(ttlSeconds),
	}

	creds, err := s.Broker.AssumeRoleWithWebIdentity(aws.AssumeRoleInput{
		RoleARN:    roleARN,
		Region:     region,
		TTLSeconds: ttlSeconds,
		Subject:    claims.Subject,
	})
	if err != nil {
		return AuthorizeResponse{}, err
	}

	finalReceipt, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
		CreatedAt:           createdAt,
		IdemKey:             idem.IdemKey,
		SupersedesReceiptID: &latest,
		ContextID:           idem.ContextID,
		DecisionID:          idem.DecisionID,
		Actor: types.ReceiptActor{
			Kind:     "workload",
			Subject:  claims.Subject,
			Issuer:   claims.Issuer,
			Repo:     claims.Repo,
			Workflow: claims.Workflow,
			RunID:    claims.RunID,
			SHA:      claims.SHA,
		},
		Request: types.ReceiptRequest{
			RequestID: req.RequestID,
			Action:    req.Action,
			Resource:  req.Resource,
			Env:       req.Env,
			Intent:    req.Intent,
		},
		Policy:          types.ReceiptPolicy{PolicyHash: idem.PolicyHash},
		CredentialGrant: credentialGrant,
		Outcome:         types.ReceiptOutcome{Status: types.OutcomeIssuedCredentials, ExpiresAt: creds.ExpiresAt.UTC().Format(time.RFC3339)},
	}, s.Signer)
	if err != nil {
		return AuthorizeResponse{}, err
	}

	idem.Status = IdemAllowed
	idem.LatestReceiptID = finalReceipt.ReceiptID
	idem.FinalReceiptID = finalReceipt.ReceiptID
	s.Store.Put(idem)

	return AuthorizeResponse{
		Verdict:    string(VerdictAllow),
		ContextID:  idem.ContextID,
		DecisionID: idem.DecisionID,
		ReceiptID:  finalReceipt.ReceiptID,
		AWSCredentials: &struct {
			AccessKeyID     string `json:"access_key_id"`
			SecretAccessKey string `json:"secret_access_key"`
			SessionToken    string `json:"session_token"`
			ExpiresAt       string `json:"expires_at"`
		}{
			AccessKeyID:     creds.AccessKeyID,
			SecretAccessKey: creds.SecretAccessKey,
			SessionToken:    creds.SessionToken,
			ExpiresAt:       creds.ExpiresAt.UTC().Format(time.RFC3339),
		},
	}, nil
}

func (s *AuthorizeService) findIdemByApproval(approvalID string) (IdemRecord, bool) {
	for _, rec := range s.Store.items {
		if rec.ApprovalID == approvalID {
			return rec, true
		}
	}
	return IdemRecord{}, false
}

type devSigner struct {
	keyID string
	priv  ed25519.PrivateKey
}

func (s devSigner) KeyID() string {
	return s.keyID
}

func (s devSigner) SignEd25519(message []byte) ([]byte, error) {
	return ed25519.Sign(s.priv, message), nil
}

func newApprovalID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("approval-%d", time.Now().UnixNano())
	}
	return "approval-" + hex.EncodeToString(buf)
}
