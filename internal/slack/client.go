package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	Token   string
	BaseURL string
	HTTP    *http.Client
}

func (c *Client) PostApproval(channel string, message ApprovalMessageInput) (string, error) {
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 10 * time.Second}
	}
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = "https://slack.com/api"
	}
	if c.Token == "" {
		return "", fmt.Errorf("missing slack token")
	}
	if channel == "" {
		return "", fmt.Errorf("missing slack channel")
	}

	msgBytes, err := BuildApprovalMessage(message)
	if err != nil {
		return "", err
	}

	var payload map[string]any
	if err := json.Unmarshal(msgBytes, &payload); err != nil {
		return "", err
	}
	payload["channel"] = channel

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/chat.postMessage", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
		TS    string `json:"ts,omitempty"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return "", err
	}
	if !resp.OK {
		if resp.Error == "" {
			resp.Error = "slack api error"
		}
		return "", fmt.Errorf("%s", resp.Error)
	}
	if resp.TS == "" {
		return "", fmt.Errorf("missing slack message ts")
	}
	return resp.TS, nil
}
