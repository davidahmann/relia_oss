package api

import (
	stdcontext "context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidahmann/relia/internal/aws"
	reliactx "github.com/davidahmann/relia/internal/context"
	"github.com/davidahmann/relia/internal/decision"
	"github.com/davidahmann/relia/internal/ledger"
	"github.com/davidahmann/relia/internal/policy"
	"github.com/davidahmann/relia/internal/slack"
	"github.com/davidahmann/relia/pkg/types"
)

type AuthorizeService struct {
	PolicyPath string
	Ledger     ledger.Store
	Signer     ledger.Signer
	Broker     aws.CredentialBroker
	PublicKey  ed25519.PublicKey
	Slack      SlackNotifier
	SlackChan  string
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

type SlackNotifier interface {
	PostApproval(channel string, message slack.ApprovalMessageInput) (msgTS string, err error)
}

type NewAuthorizeServiceInput struct {
	PolicyPath string
	Ledger     ledger.Store
	Signer     ledger.Signer
	PublicKey  ed25519.PublicKey
	Broker     aws.CredentialBroker
	Slack      SlackNotifier
	SlackChan  string
}

func NewAuthorizeService(in NewAuthorizeServiceInput) (*AuthorizeService, error) {
	if in.PolicyPath == "" {
		return nil, fmt.Errorf("missing policy path")
	}
	if in.Ledger == nil {
		in.Ledger = ledger.NewInMemoryStore()
	}
	if in.Signer == nil || in.PublicKey == nil {
		seed := make([]byte, ed25519.SeedSize)
		if _, err := rand.Read(seed); err != nil {
			return nil, err
		}
		priv := ed25519.NewKeyFromSeed(seed)
		in.PublicKey = priv.Public().(ed25519.PublicKey)
		in.Signer = devSigner{keyID: "dev", priv: priv}
	}
	if in.Broker == nil {
		in.Broker = aws.DevBroker{}
	}
	return &AuthorizeService{
		PolicyPath: in.PolicyPath,
		Ledger:     in.Ledger,
		Signer:     in.Signer,
		Broker:     in.Broker,
		PublicKey:  in.PublicKey,
		Slack:      in.Slack,
		SlackChan:  in.SlackChan,
	}, nil
}

func (s *AuthorizeService) Authorize(claims ActorContext, req AuthorizeRequest, createdAt string) (AuthorizeResponse, error) {
	idemKey, err := ComputeIdemKey(claims, req)
	if err != nil {
		return AuthorizeResponse{}, err
	}

	existing, ok := s.Ledger.GetIdempotencyKey(idemKey)
	if ok {
		switch IdemStatus(existing.Status) {
		case IdemAllowed, IdemDenied, IdemPendingApproval:
			return s.handleExisting(existing)
		case IdemApprovedReady:
			return s.issueApprovedReady(idemKey, existing, claims, req, createdAt)
		case IdemIssuing:
			return s.retryIssuing(idemKey, existing, claims, req, createdAt)
		case IdemErrored:
			return AuthorizeResponse{Verdict: string(VerdictDeny), Error: "previous error"}, nil
		default:
			return AuthorizeResponse{Verdict: string(VerdictDeny), Error: "unsupported state"}, nil
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

	ctxRecord, err := reliactx.BuildContext(source, inputs, evidence, createdAt)
	if err != nil {
		return AuthorizeResponse{}, err
	}

	policyMeta := types.DecisionPolicy{PolicyID: loaded.Policy.PolicyID, PolicyVersion: loaded.Policy.PolicyVersion, PolicyHash: loaded.Hash}
	decRecord, err := decision.BuildDecision(ctxRecord.ContextID, policyMeta, decisionResult.Verdict, decisionResult.ReasonCodes, decisionResult.RequireApproval, decisionResult.Risk, createdAt)
	if err != nil {
		return AuthorizeResponse{}, err
	}

	verdict := DecisionVerdict(decisionResult.Verdict)
	_, action := TransitionFromDecision(verdict)

	approvalID := ""
	var approval *types.ReceiptApproval
	if action == ActionReturnPending {
		approvalID = newApprovalID()
		approval = &types.ReceiptApproval{Required: true, ApprovalID: approvalID, Status: string(ApprovalPending)}
	}

	receiptPolicy := types.ReceiptPolicy(policyMeta)

	outcome := outcomeForAction(action)

	baseReceipt, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
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

	ctxJSON, err := json.Marshal(ctxRecord)
	if err != nil {
		return AuthorizeResponse{}, err
	}
	decJSON, err := json.Marshal(decRecord)
	if err != nil {
		return AuthorizeResponse{}, err
	}

	policyRec := ledger.PolicyVersionRecord{
		PolicyHash:    loaded.Hash,
		PolicyID:      loaded.Policy.PolicyID,
		PolicyVersion: loaded.Policy.PolicyVersion,
		PolicyYAML:    string(loaded.Bytes),
		CreatedAt:     createdAt,
	}

	var postSlack bool
	var approvalRec ledger.ApprovalRecord
	var outboxRec *ledger.SlackOutboxRecord

	var initialStatus IdemStatus
	var finalReceiptID *string
	latestReceiptID := baseReceipt.ReceiptID

	switch action {
	case ActionReturnDenied:
		initialStatus = IdemDenied
		finalReceiptID = &latestReceiptID
	case ActionReturnPending:
		initialStatus = IdemPendingApproval
		finalReceiptID = nil
		approvalRec = ledger.ApprovalRecord{
			ApprovalID: approvalID,
			IdemKey:    idemKey,
			Status:     string(ApprovalPending),
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		}
		postSlack = s.Slack != nil && s.SlackChan != ""
		if postSlack {
			input := slack.ApprovalMessageInput{
				ApprovalID: approvalRec.ApprovalID,
				ReceiptID:  baseReceipt.ReceiptID,
				PolicyHash: loaded.Hash,
				ContextID:  ctxRecord.ContextID,
				DecisionID: decRecord.DecisionID,
				Action:     req.Action,
				Resource:   req.Resource,
				Env:        req.Env,
				Risk:       decisionResult.Risk,
				DiffURL:    req.Evidence.DiffURL,
			}
			msgBytes, err := json.Marshal(input)
			if err != nil {
				return AuthorizeResponse{}, err
			}
			outbox := ledger.SlackOutboxRecord{
				NotificationID: "slack:" + approvalRec.ApprovalID,
				ApprovalID:     approvalRec.ApprovalID,
				Channel:        s.SlackChan,
				MessageJSON:    msgBytes,
				Status:         slack.OutboxStatusPending,
				AttemptCount:   0,
				NextAttemptAt:  createdAt,
				CreatedAt:      createdAt,
				UpdatedAt:      createdAt,
			}
			outboxRec = &outbox
		}
	case ActionIssueCredentials:
		initialStatus = IdemIssuing
		finalReceiptID = nil
	default:
		initialStatus = IdemErrored
		finalReceiptID = &latestReceiptID
	}

	err = s.Ledger.WithTx(func(tx ledger.Tx) error {
		if err := s.putSigningKey(tx, createdAt); err != nil {
			return err
		}
		if err := tx.PutPolicyVersion(policyRec); err != nil {
			return err
		}
		if err := tx.PutContext(ledger.ContextRecord{ContextID: ctxRecord.ContextID, BodyJSON: ctxJSON, CreatedAt: createdAt}); err != nil {
			return err
		}
		if err := tx.PutDecision(ledger.DecisionRecord{DecisionID: decRecord.DecisionID, ContextID: decRecord.ContextID, PolicyHash: loaded.Hash, Verdict: decisionResult.Verdict, BodyJSON: decJSON, CreatedAt: createdAt}); err != nil {
			return err
		}

		idem := ledger.IdempotencyKey{
			IdemKey:   idemKey,
			Status:    string(initialStatus),
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		}
		if err := tx.PutIdempotencyKey(idem); err != nil {
			return err
		}

		if action == ActionReturnPending {
			if err := tx.PutApproval(approvalRec); err != nil {
				return err
			}
			if outboxRec != nil {
				if err := tx.PutSlackOutbox(*outboxRec); err != nil {
					return err
				}
			}
			idem.ApprovalID = ptrOrNil(approvalID)
			if err := tx.PutIdempotencyKey(idem); err != nil {
				return err
			}
		}

		if err := tx.PutReceipt(receiptRecordFromStored(baseReceipt)); err != nil {
			return err
		}

		idem.LatestReceiptID = &latestReceiptID
		if finalReceiptID != nil {
			idem.FinalReceiptID = finalReceiptID
		}
		return tx.PutIdempotencyKey(idem)
	})
	if err != nil {
		// If another request won the race, return its view.
		if existing, ok := s.Ledger.GetIdempotencyKey(idemKey); ok {
			return s.handleExisting(existing)
		}
		return AuthorizeResponse{}, err
	}

	if action == ActionReturnPending && postSlack {
		_, _ = slack.ProcessOutboxDue(stdcontext.Background(), s.Ledger, s.Slack, time.Now().UTC(), 1)
	}

	if action == ActionIssueCredentials {
		// Write issuing receipt already done; finalize with creds and final receipt.
		return s.finalizeIssuance(idemKey, baseReceipt, claims, req, createdAt, decisionResult.AWSRoleARN, decisionResult.TTLSeconds)
	}

	resp := AuthorizeResponse{
		Verdict:    string(verdict),
		ContextID:  ctxRecord.ContextID,
		DecisionID: decRecord.DecisionID,
		ReceiptID:  baseReceipt.ReceiptID,
	}
	if action == ActionReturnPending {
		resp.Verdict = string(VerdictRequireApproval)
		resp.Approval = &struct {
			ApprovalID string `json:"approval_id"`
			Status     string `json:"status"`
		}{ApprovalID: approvalID, Status: string(ApprovalPending)}
	}
	if action == ActionReturnDenied {
		resp.Verdict = string(VerdictDeny)
	}
	return resp, nil
}

func (s *AuthorizeService) retryIssuing(idemKey string, idem ledger.IdempotencyKey, claims ActorContext, req AuthorizeRequest, createdAt string) (AuthorizeResponse, error) {
	if idem.LatestReceiptID == nil || *idem.LatestReceiptID == "" {
		return AuthorizeResponse{Verdict: string(VerdictDeny), Error: "issuing in progress"}, nil
	}
	issuingRec, ok := s.Ledger.GetReceipt(*idem.LatestReceiptID)
	if !ok {
		return AuthorizeResponse{}, fmt.Errorf("issuing receipt not found")
	}

	// Reload policy snapshot for this intent and re-evaluate to get role/ttl (keeps behavior deterministic
	// across retries without holding DB locks during AWS calls).
	policyVersion, ok := s.Ledger.GetPolicyVersion(issuingRec.PolicyHash)
	if !ok {
		return AuthorizeResponse{}, fmt.Errorf("policy version not found")
	}
	loaded, err := policy.LoadPolicyFromBytes([]byte(policyVersion.PolicyYAML))
	if err != nil {
		return AuthorizeResponse{}, err
	}

	input := policy.Input{Action: req.Action, Resource: req.Resource, Env: req.Env}
	decisionResult := policy.Evaluate(loaded.Policy, loaded.Hash, input)
	if decisionResult.AWSRoleARN == "" {
		return AuthorizeResponse{}, fmt.Errorf("missing aws_role_arn in policy")
	}

	stored := ledger.StoredReceipt{
		ReceiptID:  issuingRec.ReceiptID,
		BodyDigest: issuingRec.BodyDigest,
		BodyJSON:   issuingRec.BodyJSON,
		KeyID:      issuingRec.KeyID,
		Sig:        issuingRec.Sig,

		IdemKey:       issuingRec.IdemKey,
		CreatedAt:     issuingRec.CreatedAt,
		ContextID:     issuingRec.ContextID,
		DecisionID:    issuingRec.DecisionID,
		OutcomeStatus: types.OutcomeStatus(issuingRec.OutcomeStatus),
		ApprovalID:    issuingRec.ApprovalID,
		PolicyHash:    issuingRec.PolicyHash,
		Final:         issuingRec.Final,
		ExpiresAt:     issuingRec.ExpiresAt,
	}

	return s.finalizeIssuance(idemKey, stored, claims, req, createdAt, decisionResult.AWSRoleARN, decisionResult.TTLSeconds)
}

func (s *AuthorizeService) Approve(approvalID string, status string, createdAt string) (string, error) {
	approvalStatus := ApprovalStatus(status)
	if approvalStatus != ApprovalApproved && approvalStatus != ApprovalDenied {
		return "", fmt.Errorf("invalid approval status")
	}

	var receiptID string
	err := s.Ledger.WithTx(func(tx ledger.Tx) error {
		if err := s.putSigningKey(tx, createdAt); err != nil {
			return err
		}
		approval, ok := tx.GetApproval(approvalID)
		if !ok {
			return fmt.Errorf("approval not found")
		}
		if approval.Status == string(ApprovalApproved) || approval.Status == string(ApprovalDenied) {
			// already finalized
			idem, ok := tx.GetIdempotencyKey(approval.IdemKey)
			if ok && idem.LatestReceiptID != nil {
				receiptID = *idem.LatestReceiptID
			}
			return nil
		}

		idem, ok := tx.GetIdempotencyKey(approval.IdemKey)
		if !ok || idem.LatestReceiptID == nil {
			return fmt.Errorf("idempotency not found for approval")
		}
		latestReceipt, ok := tx.GetReceipt(*idem.LatestReceiptID)
		if !ok {
			return fmt.Errorf("latest receipt not found")
		}

		outcome := types.ReceiptOutcome{Status: types.OutcomeApprovalDenied}
		if approvalStatus == ApprovalApproved {
			outcome.Status = types.OutcomeApprovalApproved
		}

		approvalReceipt, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
			CreatedAt:           createdAt,
			IdemKey:             approval.IdemKey,
			SupersedesReceiptID: idem.LatestReceiptID,
			ContextID:           latestReceipt.ContextID,
			DecisionID:          latestReceipt.DecisionID,
			Actor:               types.ReceiptActor{Kind: "approval", Subject: "slack"},
			Request:             types.ReceiptRequest{RequestID: "approval", Action: "approve", Resource: approval.IdemKey, Env: ""},
			Policy:              types.ReceiptPolicy{PolicyHash: latestReceipt.PolicyHash},
			Approval: &types.ReceiptApproval{
				Required:   true,
				ApprovalID: approvalID,
				Status:     string(approvalStatus),
			},
			Outcome: outcome,
		}, s.Signer)
		if err != nil {
			return err
		}

		if err := tx.PutReceipt(receiptRecordFromStored(approvalReceipt)); err != nil {
			return err
		}

		approval.Status = string(approvalStatus)
		approval.UpdatedAt = createdAt
		if err := tx.PutApproval(approval); err != nil {
			return err
		}

		idem.Status = string(IdemApprovedReady)
		idem.LatestReceiptID = &approvalReceipt.ReceiptID
		idem.FinalReceiptID = nil
		if approvalStatus == ApprovalDenied {
			idem.Status = string(IdemDenied)
			idem.FinalReceiptID = &approvalReceipt.ReceiptID
		}
		idem.UpdatedAt = createdAt
		if err := tx.PutIdempotencyKey(idem); err != nil {
			return err
		}

		receiptID = approvalReceipt.ReceiptID
		return nil
	})
	return receiptID, err
}

func (s *AuthorizeService) GetApproval(approvalID string) (ledger.ApprovalRecord, bool) {
	return s.Ledger.GetApproval(approvalID)
}

func (s *AuthorizeService) handleExisting(idem ledger.IdempotencyKey) (AuthorizeResponse, error) {
	switch IdemStatus(idem.Status) {
	case IdemAllowed, IdemDenied:
		final := ""
		if idem.FinalReceiptID != nil {
			final = *idem.FinalReceiptID
		}
		if final == "" && idem.LatestReceiptID != nil {
			final = *idem.LatestReceiptID
		}
		receipt, ok := s.Ledger.GetReceipt(final)
		if !ok {
			return AuthorizeResponse{}, fmt.Errorf("final receipt not found")
		}
		verdict := VerdictAllow
		if IdemStatus(idem.Status) == IdemDenied {
			verdict = VerdictDeny
		}
		return AuthorizeResponse{Verdict: string(verdict), ContextID: receipt.ContextID, DecisionID: receipt.DecisionID, ReceiptID: receipt.ReceiptID}, nil
	case IdemPendingApproval:
		var approvalID string
		if idem.ApprovalID != nil {
			approvalID = *idem.ApprovalID
		}
		latest := ""
		if idem.LatestReceiptID != nil {
			latest = *idem.LatestReceiptID
		}
		receipt, ok := s.Ledger.GetReceipt(latest)
		if !ok {
			return AuthorizeResponse{}, fmt.Errorf("latest receipt not found")
		}
		return AuthorizeResponse{
			Verdict:    string(VerdictRequireApproval),
			ContextID:  receipt.ContextID,
			DecisionID: receipt.DecisionID,
			ReceiptID:  receipt.ReceiptID,
			Approval: &struct {
				ApprovalID string `json:"approval_id"`
				Status     string `json:"status"`
			}{ApprovalID: approvalID, Status: string(ApprovalPending)},
		}, nil
	default:
		return AuthorizeResponse{Verdict: string(VerdictDeny), Error: "unsupported state"}, nil
	}
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
	return fmt.Sprintf("approval-%x", buf)
}

func receiptRecordFromStored(stored ledger.StoredReceipt) ledger.ReceiptRecord {
	return ledger.ReceiptRecord{
		ReceiptID:           stored.ReceiptID,
		IdemKey:             stored.IdemKey,
		CreatedAt:           stored.CreatedAt,
		SupersedesReceiptID: stored.SupersedesReceiptID,
		ContextID:           stored.ContextID,
		DecisionID:          stored.DecisionID,
		PolicyHash:          stored.PolicyHash,
		ApprovalID:          stored.ApprovalID,
		OutcomeStatus:       string(stored.OutcomeStatus),
		Final:               stored.Final,
		ExpiresAt:           stored.ExpiresAt,
		BodyJSON:            stored.BodyJSON,
		BodyDigest:          stored.BodyDigest,
		KeyID:               stored.KeyID,
		Sig:                 stored.Sig,
	}
}

func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func outcomeForAction(action NextAction) types.ReceiptOutcome {
	switch action {
	case ActionReturnDenied:
		return types.ReceiptOutcome{Status: types.OutcomeDenied}
	case ActionReturnPending:
		return types.ReceiptOutcome{Status: types.OutcomeApprovalPending}
	case ActionIssueCredentials:
		return types.ReceiptOutcome{Status: types.OutcomeIssuingCredentials}
	default:
		return types.ReceiptOutcome{Status: types.OutcomeIssueFailed}
	}
}

func (s *AuthorizeService) putSigningKey(tx ledger.Tx, createdAt string) error {
	if s.Signer == nil || s.PublicKey == nil {
		return nil
	}
	pub := make([]byte, len(s.PublicKey))
	copy(pub, s.PublicKey)
	return tx.PutKey(ledger.KeyRecord{
		KeyID:     s.Signer.KeyID(),
		PublicKey: pub,
		CreatedAt: createdAt,
	})
}

func (s *AuthorizeService) finalizeIssuance(idemKey string, issuingReceipt ledger.StoredReceipt, claims ActorContext, req AuthorizeRequest, createdAt string, roleARN string, ttlSeconds int) (AuthorizeResponse, error) {
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
		RoleARN:          roleARN,
		Region:           region,
		TTLSeconds:       ttlSeconds,
		Subject:          claims.Subject,
		WebIdentityToken: claims.Token,
	})
	if err != nil {
		return AuthorizeResponse{}, err
	}

	finalReceipt, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
		CreatedAt:           createdAt,
		IdemKey:             idemKey,
		SupersedesReceiptID: &issuingReceipt.ReceiptID,
		ContextID:           issuingReceipt.ContextID,
		DecisionID:          issuingReceipt.DecisionID,
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
		Policy:          types.ReceiptPolicy{PolicyHash: issuingReceipt.PolicyHash},
		CredentialGrant: credentialGrant,
		Outcome:         types.ReceiptOutcome{Status: types.OutcomeIssuedCredentials, ExpiresAt: creds.ExpiresAt.UTC().Format(time.RFC3339)},
	}, s.Signer)
	if err != nil {
		return AuthorizeResponse{}, err
	}

	err = s.Ledger.WithTx(func(tx ledger.Tx) error {
		if err := s.putSigningKey(tx, createdAt); err != nil {
			return err
		}
		if err := tx.PutReceipt(receiptRecordFromStored(finalReceipt)); err != nil {
			return err
		}
		idem, ok := tx.GetIdempotencyKey(idemKey)
		if !ok {
			return fmt.Errorf("idempotency key missing")
		}
		idem.Status = string(IdemAllowed)
		idem.LatestReceiptID = &finalReceipt.ReceiptID
		idem.FinalReceiptID = &finalReceipt.ReceiptID
		idem.UpdatedAt = createdAt
		return tx.PutIdempotencyKey(idem)
	})
	if err != nil {
		return AuthorizeResponse{}, err
	}

	return AuthorizeResponse{
		Verdict:    string(VerdictAllow),
		ContextID:  issuingReceipt.ContextID,
		DecisionID: issuingReceipt.DecisionID,
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

func (s *AuthorizeService) issueApprovedReady(idemKey string, idem ledger.IdempotencyKey, claims ActorContext, req AuthorizeRequest, createdAt string) (AuthorizeResponse, error) {
	if idem.LatestReceiptID == nil {
		return AuthorizeResponse{}, fmt.Errorf("missing latest receipt id")
	}
	latest, ok := s.Ledger.GetReceipt(*idem.LatestReceiptID)
	if !ok {
		return AuthorizeResponse{}, fmt.Errorf("latest receipt not found")
	}

	policyVersion, ok := s.Ledger.GetPolicyVersion(latest.PolicyHash)
	if !ok {
		return AuthorizeResponse{}, fmt.Errorf("policy version not found")
	}
	loaded, err := policy.LoadPolicyFromBytes([]byte(policyVersion.PolicyYAML))
	if err != nil {
		return AuthorizeResponse{}, err
	}

	input := policy.Input{Action: req.Action, Resource: req.Resource, Env: req.Env}
	decisionResult := policy.Evaluate(loaded.Policy, loaded.Hash, input)
	if decisionResult.Verdict != string(VerdictAllow) && decisionResult.Verdict != string(VerdictRequireApproval) {
		return AuthorizeResponse{}, fmt.Errorf("unexpected verdict for approved_ready: %s", decisionResult.Verdict)
	}
	if decisionResult.AWSRoleARN == "" {
		return AuthorizeResponse{}, fmt.Errorf("missing aws_role_arn in policy")
	}

	issuingReceipt, err := ledger.MakeReceipt(ledger.MakeReceiptInput{
		CreatedAt:           createdAt,
		IdemKey:             idemKey,
		SupersedesReceiptID: idem.LatestReceiptID,
		ContextID:           latest.ContextID,
		DecisionID:          latest.DecisionID,
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
		Policy:  types.ReceiptPolicy{PolicyHash: latest.PolicyHash},
		Outcome: types.ReceiptOutcome{Status: types.OutcomeIssuingCredentials},
	}, s.Signer)
	if err != nil {
		return AuthorizeResponse{}, err
	}

	if err := s.Ledger.WithTx(func(tx ledger.Tx) error {
		if err := s.putSigningKey(tx, createdAt); err != nil {
			return err
		}
		if err := tx.PutReceipt(receiptRecordFromStored(issuingReceipt)); err != nil {
			return err
		}
		current, ok := tx.GetIdempotencyKey(idemKey)
		if !ok {
			return fmt.Errorf("idempotency key missing")
		}
		if IdemStatus(current.Status) != IdemApprovedReady {
			return fmt.Errorf("unexpected state: %s", current.Status)
		}
		current.Status = string(IdemIssuing)
		current.LatestReceiptID = &issuingReceipt.ReceiptID
		current.UpdatedAt = createdAt
		return tx.PutIdempotencyKey(current)
	}); err != nil {
		return AuthorizeResponse{}, err
	}

	return s.finalizeIssuance(idemKey, issuingReceipt, claims, req, createdAt, decisionResult.AWSRoleARN, decisionResult.TTLSeconds)
}
