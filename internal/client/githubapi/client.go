package githubapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/OutOfStack/game-library-auth/pkg/observability"
	"go.opentelemetry.io/otel"
)

const (
	accessTokenURL = "https://github.com/login/oauth/access_token" //nolint:gosec // not a credential
	userAPIURL     = "https://api.github.com/user"
)

var tracer = otel.Tracer("githubapi")

// Config represents GitHub OAuth client configuration
type Config struct {
	ClientID     string
	ClientSecret string
	Timeout      time.Duration
}

// Client represents GitHub OAuth API client
type Client struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
}

// NewClient creates a new GitHub OAuth client
func NewClient(cfg Config) *Client {
	httpClient := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: observability.NewTransport("github_api_client", observability.WithOtel()),
	}
	return &Client{
		httpClient:   httpClient,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
	}
}

// ExchangeCodeForUser exchanges an authorization code for user info
func (c *Client) ExchangeCodeForUser(ctx context.Context, code string) (UserInfo, error) {
	ctx, span := tracer.Start(ctx, "exchangeCodeForUser")
	defer span.End()

	// exchange code for access token
	accessToken, err := c.exchangeCode(ctx, code)
	if err != nil {
		return UserInfo{}, fmt.Errorf("exchange code: %w", err)
	}

	// fetch user info
	user, err := c.fetchUser(ctx, accessToken)
	if err != nil {
		return UserInfo{}, fmt.Errorf("fetch user: %w", err)
	}

	return user, nil
}

// exchangeCode exchanges an authorization code for an access token
func (c *Client) exchangeCode(ctx context.Context, code string) (string, error) {
	body := url.Values{
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"code":          {code},
	}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, accessTokenURL, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req) //nolint:gosec // URL is from config
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tokenResp accessTokenResponse
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf("token error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token")
	}

	return tokenResp.AccessToken, nil
}

// fetchUser fetches user info from GitHub API using an access token
func (c *Client) fetchUser(ctx context.Context, accessToken string) (UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userAPIURL, nil)
	if err != nil {
		return UserInfo{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req) //nolint:gosec // URL is from config
	if err != nil {
		return UserInfo{}, fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return UserInfo{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var user githubUser
	if err = json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return UserInfo{}, fmt.Errorf("decode response: %w", err)
	}

	if user.ID == 0 {
		return UserInfo{}, fmt.Errorf("invalid user: missing id")
	}

	return UserInfo{
		ID:    strconv.FormatInt(user.ID, 10),
		Login: user.Login,
		Email: user.Email,
	}, nil
}
