package main

import "testing"

func TestNewServer(t *testing.T) {
	addr := "127.0.0.1:9999"
	srv := newServer(addr, "policies/relia.yaml", "")
	if srv.Addr != addr {
		t.Fatalf("expected addr %s, got %s", addr, srv.Addr)
	}
	if srv.Handler == nil {
		t.Fatalf("expected handler to be set")
	}
}
