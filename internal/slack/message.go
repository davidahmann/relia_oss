package slack

import "encoding/json"

type ApprovalMessageInput struct {
	ApprovalID string
	ReceiptID  string
	PolicyHash string
	ContextID  string
	DecisionID string
	Action     string
	Resource   string
	Env        string
	Risk       string
	DiffURL    string
	RunURL     string
}

// BuildApprovalMessage returns Slack Block Kit JSON for an approval request.
func BuildApprovalMessage(input ApprovalMessageInput) ([]byte, error) {
	blocks := []map[string]any{
		{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": "*Relia approval required*",
			},
		},
		{
			"type": "section",
			"fields": []map[string]any{
				{"type": "mrkdwn", "text": "*Action*\n" + input.Action},
				{"type": "mrkdwn", "text": "*Env*\n" + input.Env},
				{"type": "mrkdwn", "text": "*Resource*\n" + input.Resource},
				{"type": "mrkdwn", "text": "*Risk*\n" + input.Risk},
			},
		},
	}

	if input.DiffURL != "" || input.RunURL != "" {
		links := ""
		if input.DiffURL != "" {
			links += "<" + input.DiffURL + "|Diff> "
		}
		if input.RunURL != "" {
			links += "<" + input.RunURL + "|Run>"
		}
		blocks = append(blocks, map[string]any{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": links,
			},
		})
	}

	blocks = append(blocks, map[string]any{
		"type": "actions",
		"elements": []map[string]any{
			{
				"type":      "button",
				"text":      map[string]any{"type": "plain_text", "text": "Approve"},
				"style":     "primary",
				"action_id": "approve",
				"value":     input.ApprovalID,
			},
			{
				"type":      "button",
				"text":      map[string]any{"type": "plain_text", "text": "Deny"},
				"style":     "danger",
				"action_id": "deny",
				"value":     input.ApprovalID,
			},
		},
	})

	payload := map[string]any{
		"blocks": blocks,
	}

	return json.Marshal(payload)
}
