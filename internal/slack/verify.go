package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"
)

var (
	ErrMissingSignature = errors.New("missing slack signature")
	ErrInvalidSignature = errors.New("invalid slack signature")
	ErrStaleTimestamp   = errors.New("stale slack timestamp")
)

// VerifySignature validates the Slack signing secret against the request.
func VerifySignature(signingSecret, signature, timestamp string, body []byte, now time.Time) error {
	if signature == "" || timestamp == "" {
		return ErrMissingSignature
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return ErrInvalidSignature
	}

	requestTime := time.Unix(ts, 0)
	if now.Sub(requestTime) > 5*time.Minute || requestTime.Sub(now) > 5*time.Minute {
		return ErrStaleTimestamp
	}

	base := []byte("v0:" + timestamp + ":" + string(body))
	mac := hmac.New(sha256.New, []byte(signingSecret))
	_, _ = mac.Write(base)
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return ErrInvalidSignature
	}

	return nil
}
