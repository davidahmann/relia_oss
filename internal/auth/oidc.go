package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const githubIssuer = "https://token.actions.githubusercontent.com"

type GitHubOIDCAuthenticator struct {
	Audience string
	Issuer   string
	JWKSURL  string

	http *http.Client

	mu   sync.Mutex
	keys map[string]*rsa.PublicKey
}

func NewGitHubOIDCAuthenticator(audience string) *GitHubOIDCAuthenticator {
	if audience == "" {
		audience = "relia"
	}
	return &GitHubOIDCAuthenticator{
		Audience: audience,
		Issuer:   githubIssuer,
		JWKSURL:  githubIssuer + "/.well-known/jwks",
		http:     &http.Client{Timeout: 5 * time.Second},
		keys:     make(map[string]*rsa.PublicKey),
	}
}

func (a *GitHubOIDCAuthenticator) AuthenticateBearer(token string) (Claims, error) {
	if token == "" {
		return Claims{}, ErrInvalidToken
	}

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithAudience(a.Audience),
		jwt.WithIssuer(a.Issuer),
	)
	claims := &githubClaims{}
	_, err := parser.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("missing kid")
		}
		return a.keyForKID(kid)
	})
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	repo := claims.Repository
	if repo == "" {
		repo = claims.RepositoryOwner + "/" + claims.RepositoryName
	}
	if repo == "/" {
		repo = ""
	}

	workflow := firstNonEmpty(claims.WorkflowRef, claims.JobWorkflowRef)
	if workflow == "" {
		workflow = claims.Workflow
	}

	if claims.Subject == "" || repo == "" || workflow == "" || claims.RunID == "" || claims.SHA == "" {
		return Claims{}, ErrInvalidToken
	}

	return Claims{
		Subject:  claims.Subject,
		Issuer:   claims.Issuer,
		Repo:     repo,
		Workflow: workflow,
		RunID:    claims.RunID,
		SHA:      claims.SHA,
	}, nil
}

func (a *GitHubOIDCAuthenticator) keyForKID(kid string) (*rsa.PublicKey, error) {
	a.mu.Lock()
	if key, ok := a.keys[kid]; ok {
		a.mu.Unlock()
		return key, nil
	}
	a.mu.Unlock()

	keys, err := a.fetchKeys()
	if err != nil {
		return nil, err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for k, v := range keys {
		a.keys[k] = v
	}
	if key, ok := a.keys[kid]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("kid not found")
}

func (a *GitHubOIDCAuthenticator) fetchKeys() (map[string]*rsa.PublicKey, error) {
	req, err := http.NewRequest(http.MethodGet, a.JWKSURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := a.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(res.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	out := make(map[string]*rsa.PublicKey)
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" || k.Alg != "RS256" {
			continue
		}
		pub, err := jwkToPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		if k.Kid != "" {
			out[k.Kid] = pub
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no jwk keys")
	}
	return out, nil
}

func jwkToPublicKey(nB64 string, eB64 string) (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eb, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nb)
	e := 0
	for _, b := range eb {
		e = e<<8 + int(b)
	}
	if e == 0 {
		return nil, fmt.Errorf("invalid exponent")
	}
	return &rsa.PublicKey{N: n, E: e}, nil
}

type githubClaims struct {
	jwt.RegisteredClaims

	Repository      string `json:"repository"`
	RepositoryOwner string `json:"repository_owner"`
	RepositoryName  string `json:"repository_name"`

	Workflow       string `json:"workflow"`
	WorkflowRef    string `json:"workflow_ref"`
	JobWorkflowRef string `json:"job_workflow_ref"`

	RunID string `json:"run_id"`
	SHA   string `json:"sha"`
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
