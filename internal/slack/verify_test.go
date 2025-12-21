package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func TestVerifySignature(t *testing.T) {
	secret := "secret"
	body := []byte("payload=test")
	ts := time.Now().Unix()
	timestamp := fmt.Sprintf("%d", ts)

	base := []byte("v0:" + timestamp + ":" + string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(base)
	sig := "v0=" + hex.EncodeToString(mac.Sum(nil))

	err := VerifySignature(secret, sig, timestamp, body, time.Unix(ts, 0))
	if err != nil {
		t.Fatalf("expected valid signature, got %v", err)
	}
}

func TestVerifySignatureInvalid(t *testing.T) {
	err := VerifySignature("secret", "v0=bad", "123", []byte("x"), time.Now())
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestVerifySignatureMissing(t *testing.T) {
	err := VerifySignature("secret", "", "", []byte("x"), time.Now())
	if err != ErrMissingSignature {
		t.Fatalf("expected ErrMissingSignature, got %v", err)
	}
}

func TestVerifySignatureStale(t *testing.T) {
	old := time.Now().Add(-10 * time.Minute)
	timestamp := fmt.Sprintf("%d", old.Unix())
	err := VerifySignature("secret", "v0=bad", timestamp, []byte("x"), time.Now())
	if err != ErrStaleTimestamp && err != ErrInvalidSignature {
		t.Fatalf("expected stale or invalid signature, got %v", err)
	}
}

func TestVerifySignatureBadTimestamp(t *testing.T) {
	err := VerifySignature("secret", "v0=bad", "not-a-time", []byte("x"), time.Now())
	if err != ErrInvalidSignature {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}
