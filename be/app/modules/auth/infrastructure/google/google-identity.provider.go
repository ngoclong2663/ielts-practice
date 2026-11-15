package google

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	authout "ielts-practice/app/modules/auth/ports/out"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"ielts-practice/app/modules/auth/domain/entity"
	"ielts-practice/configs"
)

const googleIssuer = "https://accounts.google.com"

var _ authout.IGoogleIdentityProvider = &GoogleIdentityProvider{}

type GoogleIdentityProvider struct {
	oauth    *oauth2.Config
	clientID string

	mu       sync.Mutex
	verifier *oidc.IDTokenVerifier
}

func NewGoogleIdentityProvider(cfg *configs.Config) authout.IGoogleIdentityProvider {
	return &GoogleIdentityProvider{
		clientID: cfg.Google.ClientID,
		oauth: &oauth2.Config{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
			RedirectURL:  cfg.Google.RedirectURL,
			Endpoint:     google.Endpoint,
			Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
		},
	}
}

// getVerifier lazily runs OIDC discovery so the API can boot even if Google is unreachable.
// A failed discovery is retried on the next sign-in instead of being cached.
func (p *GoogleIdentityProvider) getVerifier() (*oidc.IDTokenVerifier, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.verifier != nil {
		return p.verifier, nil
	}
	// Use a long-lived context: the provider keeps it for later JWKS key refreshes,
	// so it must not be the (short-lived) request context.
	bg := oidc.ClientContext(context.Background(), &http.Client{Timeout: 10 * time.Second})
	provider, err := oidc.NewProvider(bg, googleIssuer)
	if err != nil {
		return nil, err
	}
	p.verifier = provider.Verifier(&oidc.Config{ClientID: p.clientID})
	return p.verifier, nil
}

func (p *GoogleIdentityProvider) Exchange(ctx context.Context, code string, codeVerifier string) (*entity.GoogleIdentity, error) {
	if p.oauth.ClientID == "" || p.oauth.ClientSecret == "" {
		return nil, authout.ErrGoogleNotConfigured
	}

	token, err := p.oauth.Exchange(ctx, code, oauth2.VerifierOption(codeVerifier))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", authout.ErrGoogleExchange, err)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, fmt.Errorf("%w: missing id_token", authout.ErrGoogleIDToken)
	}

	verifier, err := p.getVerifier()
	if err != nil {
		return nil, fmt.Errorf("%w: discovery: %v", authout.ErrGoogleExchange, err)
	}
	// Verify checks signature (Google JWKS), issuer, audience (= our client ID) and expiry.
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", authout.ErrGoogleIDToken, err)
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("%w: claims: %v", authout.ErrGoogleIDToken, err)
	}

	return &entity.GoogleIdentity{
		Subject:       idToken.Subject,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		Picture:       claims.Picture,
		Nonce:         idToken.Nonce,
	}, nil
}
