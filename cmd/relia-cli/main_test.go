package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	_ = r.Close()

	return string(out)
}

func TestMainOutput(t *testing.T) {
	out := captureOutput(t, main)
	if !strings.Contains(out, "relia-cli") {
		t.Fatalf("unexpected output: %q", out)
	}
}
