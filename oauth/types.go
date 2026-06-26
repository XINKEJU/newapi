package oauth

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

// getRedirectURI derives the OAuth redirect_uri from the actual request,
// ensuring it matches what the frontend sent during authorization.
// This is critical: OAuth providers require the redirect_uri to be identical
// in both the authorization and token-exchange steps.
//
// Priority: X-Forwarded-* headers (reverse proxy) → request scheme + host.
// Falls back to ServerAddress from OptionMap if the request scheme cannot be determined.
func getRedirectURI(c *gin.Context, provider string) string {
	scheme := "https"
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if c.Request.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}

	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}

	// If we have a usable host from the request, use it
	if host != "" {
		return fmt.Sprintf("%s://%s/oauth/%s", scheme, host, provider)
	}

	// Fallback to configured ServerAddress
	return fmt.Sprintf("%s/oauth/%s", common.OptionMap["ServerAddress"], provider)
}

// OAuthToken represents the token received from OAuth provider
type OAuthToken struct {
	AccessToken  string         `json:"access_token"`
	TokenType    string         `json:"token_type"`
	RefreshToken string         `json:"refresh_token,omitempty"`
	ExpiresIn    int            `json:"expires_in,omitempty"`
	Scope        string         `json:"scope,omitempty"`
	IDToken      string         `json:"id_token,omitempty"`
	Extra        map[string]any `json:"-"`
}

// OAuthUser represents the user info from OAuth provider
type OAuthUser struct {
	// ProviderUserID is the unique identifier from the OAuth provider
	ProviderUserID string
	// Username is the username from the OAuth provider (e.g., GitHub login)
	Username string
	// DisplayName is the display name from the OAuth provider
	DisplayName string
	// Email is the email from the OAuth provider
	Email string
	// Extra contains any additional provider-specific data
	Extra map[string]any
}

// OAuthError represents a translatable OAuth error
type OAuthError struct {
	// MsgKey is the i18n message key
	MsgKey string
	// Params contains optional parameters for the message template
	Params map[string]any
	// RawError is the underlying error for logging purposes
	RawError string
}

func (e *OAuthError) Error() string {
	if e.RawError != "" {
		return e.RawError
	}
	return e.MsgKey
}

// NewOAuthError creates a new OAuth error with the given message key
func NewOAuthError(msgKey string, params map[string]any) *OAuthError {
	return &OAuthError{
		MsgKey: msgKey,
		Params: params,
	}
}

// NewOAuthErrorWithRaw creates a new OAuth error with raw error message for logging
func NewOAuthErrorWithRaw(msgKey string, params map[string]any, rawError string) *OAuthError {
	return &OAuthError{
		MsgKey:   msgKey,
		Params:   params,
		RawError: rawError,
	}
}

// AccessDeniedError is a direct user-facing access denial message.
type AccessDeniedError struct {
	Message string
}

func (e *AccessDeniedError) Error() string {
	return e.Message
}
