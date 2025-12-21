package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGitHubOIDCAuthenticator_SuccessAndCache(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("gen rsa: %v", err)
	}

	kid := "k1"
	jwksCalls := 0
	jwksSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwksCalls++
		n := base64.RawURLEncoding.EncodeToString(priv.PublicKey.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString([]byte{1, 0, 1}) // 65537

		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{
				{"kid": kid, "kty": "RSA", "alg": "RS256", "use": "sig", "n": n, "e": e},
			},
		})
	}))
	defer jwksSrv.Close()

	a := NewGitHubOIDCAuthenticator("relia")
	a.JWKSURL = jwksSrv.URL
	a.http = jwksSrv.Client()

	now := time.Now().UTC()
	claims := githubClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    githubIssuer,
			Subject:   "repo:org/repo:ref:refs/heads/main",
			Audience:  jwt.ClaimStrings{"relia"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
		Repository:     "org/repo",
		WorkflowRef:    "terraform-prod",
		RunID:          "123",
		SHA:            "abcdef",
		Workflow:       "",
		JobWorkflowRef: "",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	out, err := a.AuthenticateBearer(signed)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if out.Repo != "org/repo" || out.Workflow != "terraform-prod" || out.RunID != "123" || out.SHA != "abcdef" {
		t.Fatalf("unexpected claims: %+v", out)
	}
	if jwksCalls != 1 {
		t.Fatalf("expected 1 jwks call, got %d", jwksCalls)
	}

	// Second call should use cached KID (no jwks fetch).
	out2, err := a.AuthenticateBearer(signed)
	if err != nil {
		t.Fatalf("authenticate cached: %v", err)
	}
	if out2.Repo != "org/repo" {
		t.Fatalf("unexpected claims from cache: %+v", out2)
	}
	if jwksCalls != 1 {
		t.Fatalf("expected jwks not to be called again, got %d", jwksCalls)
	}
}

func TestOIDCHelpers(t *testing.T) {
	if got := firstNonEmpty("", "a", "b"); got != "a" {
		t.Fatalf("expected a, got %s", got)
	}
	if _, err := jwkToPublicKey("!!!", "AQAB"); err == nil {
		t.Fatalf("expected error for bad modulus")
	}
	// exponent 0
	nb := base64.RawURLEncoding.EncodeToString([]byte{1})
	eb := base64.RawURLEncoding.EncodeToString([]byte{0})
	if _, err := jwkToPublicKey(nb, eb); err == nil {
		t.Fatalf("expected error for exponent")
	}
}

func TestMultiAuthenticator_OIDCPath(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("gen rsa: %v", err)
	}

	kid := "k2"
	jwksSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := base64.RawURLEncoding.EncodeToString(priv.PublicKey.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString([]byte{1, 0, 1})
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{
				{"kid": kid, "kty": "RSA", "alg": "RS256", "use": "sig", "n": n, "e": e},
			},
		})
	}))
	defer jwksSrv.Close()

	oidc := NewGitHubOIDCAuthenticator("relia")
	oidc.JWKSURL = jwksSrv.URL
	oidc.http = jwksSrv.Client()

	now := time.Now().UTC()
	claims := githubClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    githubIssuer,
			Subject:   "repo:org/repo:ref:refs/heads/main",
			Audience:  jwt.ClaimStrings{"relia"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
		Repository:  "org/repo",
		WorkflowRef: "wf",
		RunID:       "1",
		SHA:         "sha",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	a := &MultiAuthenticator{DevToken: "not-it", OIDC: oidc}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	out, err := a.Authenticate(req)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if out.Token != signed {
		t.Fatalf("expected token to be carried through")
	}
}

func TestGitHubOIDCAuthenticator_NoJWKKeys(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("gen rsa: %v", err)
	}

	kid := "k3"
	jwksSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return no usable RSA/RS256 keys.
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{
				{"kid": kid, "kty": "EC", "alg": "ES256", "use": "sig"},
			},
		})
	}))
	defer jwksSrv.Close()

	a := NewGitHubOIDCAuthenticator("relia")
	a.JWKSURL = jwksSrv.URL
	a.http = jwksSrv.Client()

	now := time.Now().UTC()
	claims := githubClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    githubIssuer,
			Subject:   "repo:org/repo:ref:refs/heads/main",
			Audience:  jwt.ClaimStrings{"relia"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
		Repository:  "org/repo",
		WorkflowRef: "wf",
		RunID:       "1",
		SHA:         "sha",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := a.AuthenticateBearer(signed); err == nil {
		t.Fatalf("expected invalid token")
	}
}
