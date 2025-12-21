package slack

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientPostApproval(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat.postMessage" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer xoxb-test" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"channel":"C123"`) {
			t.Fatalf("missing channel in request: %s", string(body))
		}
		_, _ = w.Write([]byte(`{"ok":true,"ts":"123.456"}`))
	}))
	defer srv.Close()

	c := &Client{Token: "xoxb-test", BaseURL: srv.URL, HTTP: srv.Client()}
	ts, err := c.PostApproval("C123", ApprovalMessageInput{ApprovalID: "a1", Action: "x", Env: "dev", Resource: "r", Risk: "low"})
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	if ts != "123.456" {
		t.Fatalf("unexpected ts: %s", ts)
	}
}

func TestClientPostApprovalErrors(t *testing.T) {
	c := &Client{Token: "", BaseURL: "https://example.test", HTTP: http.DefaultClient}
	if _, err := c.PostApproval("C1", ApprovalMessageInput{ApprovalID: "a1", Action: "x", Env: "dev", Resource: "r"}); err == nil {
		t.Fatalf("expected missing token error")
	}
	c.Token = "x"
	if _, err := c.PostApproval("", ApprovalMessageInput{ApprovalID: "a1", Action: "x", Env: "dev", Resource: "r"}); err == nil {
		t.Fatalf("expected missing channel error")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":false,"error":"nope"}`))
	}))
	defer srv.Close()

	c = &Client{Token: "x", BaseURL: srv.URL, HTTP: srv.Client()}
	if _, err := c.PostApproval("C1", ApprovalMessageInput{ApprovalID: "a1", Action: "x", Env: "dev", Resource: "r"}); err == nil {
		t.Fatalf("expected slack api error")
	}

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":false}`))
	}))
	defer srv2.Close()

	c = &Client{Token: "x", BaseURL: srv2.URL, HTTP: srv2.Client()}
	if _, err := c.PostApproval("C1", ApprovalMessageInput{ApprovalID: "a1", Action: "x", Env: "dev", Resource: "r"}); err == nil {
		t.Fatalf("expected generic slack api error")
	}

	srv3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv3.Close()

	c = &Client{Token: "x", BaseURL: srv3.URL, HTTP: srv3.Client()}
	if _, err := c.PostApproval("C1", ApprovalMessageInput{ApprovalID: "a1", Action: "x", Env: "dev", Resource: "r"}); err == nil {
		t.Fatalf("expected missing ts error")
	}

	srv4 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv4.Close()

	c = &Client{Token: "x", BaseURL: srv4.URL, HTTP: srv4.Client()}
	if _, err := c.PostApproval("C1", ApprovalMessageInput{ApprovalID: "a1", Action: "x", Env: "dev", Resource: "r"}); err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestClientPostApprovalDefaultHTTPTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"ts":"123.456"}`))
	}))
	defer srv.Close()

	c := &Client{Token: "xoxb-test", BaseURL: srv.URL}
	if _, err := c.PostApproval("C123", ApprovalMessageInput{ApprovalID: "a1", Action: "x", Env: "dev", Resource: "r"}); err != nil {
		t.Fatalf("post: %v", err)
	}
	if c.HTTP == nil {
		t.Fatalf("expected http client to be initialized")
	}
	if c.HTTP.Timeout == 0 {
		t.Fatalf("expected non-zero timeout")
	}
}
