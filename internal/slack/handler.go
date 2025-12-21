package slack

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Approver interface {
	Approve(approvalID string, status string, createdAt string) (string, error)
}

type InteractionHandler struct {
	SigningSecret string
	Approver      Approver
	Now           func() time.Time
}

func (h *InteractionHandler) HandleInteractions(w http.ResponseWriter, r *http.Request) {
	if h.Approver == nil {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("X-Slack-Signature")
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")
	if h.SigningSecret != "" {
		now := time.Now()
		if h.Now != nil {
			now = h.Now()
		}
		if err := VerifySignature(h.SigningSecret, sig, timestamp, body, now); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	payload, err := parsePayload(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	approvalID, action, err := parseAction(payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status := "denied"
	if action == "approve" {
		status = "approved"
	}

	now := time.Now()
	if h.Now != nil {
		now = h.Now()
	}

	_, err = h.Approver.Approve(approvalID, status, now.UTC().Format(time.RFC3339))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type slackPayload struct {
	Actions []struct {
		ActionID string `json:"action_id"`
		Value    string `json:"value"`
	} `json:"actions"`
}

func parsePayload(body []byte) (slackPayload, error) {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return slackPayload{}, err
	}
	payloadStr := values.Get("payload")
	if payloadStr == "" {
		return slackPayload{}, errors.New("missing payload")
	}

	var payload slackPayload
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		return slackPayload{}, err
	}
	if len(payload.Actions) == 0 {
		return slackPayload{}, errors.New("missing actions")
	}
	return payload, nil
}

func parseAction(payload slackPayload) (string, string, error) {
	action := payload.Actions[0]
	if action.ActionID != "" {
		return action.Value, action.ActionID, nil
	}
	if action.Value == "" {
		return "", "", errors.New("missing action value")
	}

	// Fallback: expect value as "approve:<id>" or "deny:<id>"
	parts := strings.SplitN(action.Value, ":", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid action value")
	}
	return parts[1], parts[0], nil
}
