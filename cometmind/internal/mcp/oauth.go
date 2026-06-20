package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/oauth2"
)

const oauthDirName = "mcp-oauth"

// OAuthTokenDir returns ~/.cometmind/mcp-oauth (created if missing).
func OAuthTokenDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".cometmind", oauthDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

func oauthTokenPath(serverID string) (string, error) {
	dir, err := OAuthTokenDir()
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(serverID)
	if id == "" {
		return "", fmt.Errorf("empty MCP server id")
	}
	return filepath.Join(dir, id+".json"), nil
}

// LoadOAuthToken reads a stored OAuth token for one MCP server.
func LoadOAuthToken(serverID string) (*oauth2.Token, error) {
	path, err := oauthTokenPath(serverID)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tok oauth2.Token
	if err := json.Unmarshal(raw, &tok); err != nil {
		return nil, fmt.Errorf("parse oauth token: %w", err)
	}
	if strings.TrimSpace(tok.AccessToken) == "" {
		return nil, fmt.Errorf("oauth token missing access_token")
	}
	return &tok, nil
}

// SaveOAuthToken writes an OAuth token for one MCP server (mode 0600).
func SaveOAuthToken(serverID string, tok *oauth2.Token) error {
	if tok == nil || strings.TrimSpace(tok.AccessToken) == "" {
		return fmt.Errorf("empty oauth token")
	}
	path, err := oauthTokenPath(serverID)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// OAuthConnected reports whether a non-empty token file exists for the server.
func OAuthConnected(serverID string) bool {
	tok, err := LoadOAuthToken(serverID)
	return err == nil && tok != nil && strings.TrimSpace(tok.AccessToken) != ""
}

type fileOAuthHandler struct {
	serverID string
}

func (h fileOAuthHandler) TokenSource(_ context.Context) (oauth2.TokenSource, error) {
	tok, err := LoadOAuthToken(h.serverID)
	if err != nil {
		return nil, nil
	}
	return oauth2.StaticTokenSource(tok), nil
}

func (h fileOAuthHandler) Authorize(_ context.Context, _ *http.Request, resp *http.Response) error {
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	return fmt.Errorf("MCP OAuth authorization required; connect via Cometline Settings for server %q", h.serverID)
}

func oauthHandlerFor(serverID string, oauth *OAuthConfig) auth.OAuthHandler {
	if oauth == nil {
		return nil
	}
	if !OAuthConnected(serverID) {
		return nil
	}
	return fileOAuthHandler{serverID: serverID}
}

func httpClientWithHeaders(base *http.Client, headers map[string]string, oauth *OAuthConfig, serverID string) *http.Client {
	if base == nil {
		base = http.DefaultClient
	}
	transport := base.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	wrapped := &headerTransport{
		base:     transport,
		headers:  headers,
		serverID: serverID,
		oauth:    oauth,
	}
	client := &http.Client{
		Transport: wrapped,
		Timeout:   base.Timeout,
	}
	if base.Timeout == 0 {
		client.Timeout = 0
	}
	return client
}

type headerTransport struct {
	base     http.RoundTripper
	headers  map[string]string
	serverID string
	oauth    *OAuthConfig
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		if strings.TrimSpace(v) == "" {
			continue
		}
		req.Header.Set(k, v)
	}
	if t.oauth != nil && t.serverID != "" {
		if tok, err := LoadOAuthToken(t.serverID); err == nil && tok != nil {
			tokenType := strings.TrimSpace(tok.TokenType)
			if tokenType == "" {
				tokenType = "Bearer"
			}
			req.Header.Set("Authorization", tokenType+" "+tok.AccessToken)
		}
	}
	return t.base.RoundTrip(req)
}

func streamableTransport(cfg ServerConfig) *mcp.StreamableClientTransport {
	client := httpClientWithHeaders(nil, cfg.Headers, cfg.OAuth, cfg.ID)
	return &mcp.StreamableClientTransport{
		Endpoint:     cfg.URL,
		HTTPClient:   client,
		OAuthHandler: oauthHandlerFor(cfg.ID, cfg.OAuth),
	}
}

func sseTransport(cfg ServerConfig) *mcp.SSEClientTransport {
	return &mcp.SSEClientTransport{
		Endpoint:   cfg.URL,
		HTTPClient: httpClientWithHeaders(nil, cfg.Headers, cfg.OAuth, cfg.ID),
	}
}

// TokenExpiry returns the token expiry time when a token file exists.
func TokenExpiry(serverID string) *time.Time {
	tok, err := LoadOAuthToken(serverID)
	if err != nil || tok == nil {
		return nil
	}
	if tok.Expiry.IsZero() {
		return nil
	}
	exp := tok.Expiry
	return &exp
}
