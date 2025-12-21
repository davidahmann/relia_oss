package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidahmann/relia/internal/ledger"
)

func TestPublicVerifyDisabledReturns404(t *testing.T) {
	router := NewRouter(&Handler{PublicVerify: false, AuthorizeService: &AuthorizeService{Ledger: ledger.NewInMemoryStore()}})

	req := httptest.NewRequest(http.MethodGet, "/verify/anything", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/pack/anything", nil)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr2.Code)
	}
}

func TestPublicPackMissingAuthorizeServiceReturns404(t *testing.T) {
	router := NewRouter(&Handler{PublicVerify: true})
	req := httptest.NewRequest(http.MethodGet, "/pack/anything", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
