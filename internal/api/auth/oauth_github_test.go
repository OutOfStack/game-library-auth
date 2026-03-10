package auth_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/api/auth"
	mocks "github.com/OutOfStack/game-library-auth/internal/api/auth/mocks"
	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/client/githubapi"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestGitHubOAuthHandler_InvalidRequest(t *testing.T) {
	logger := zap.NewNop()

	t.Run("invalid request body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockGoogleTokenValidator := mocks.NewMockGoogleIDTokenClient(ctrl)
		mockGitHubOAuthClient := mocks.NewMockGitHubOAuthClient(ctrl)
		mockUserFacade := mocks.NewMockUserFacade(ctrl)
		authAPI, err := auth.NewAPI(logger, mockGoogleTokenValidator, mockGitHubOAuthClient, mockUserFacade, auth.APICfg{
			ContactEmail:               "contact@example.com",
			RefreshTokenCookieSameSite: "strict",
			RefreshTokenCookieSecure:   true,
		})
		require.NoError(t, err)
		app := fiber.New()
		app.Post("/oauth/github", authAPI.GitHubOAuthHandler)

		req := httptest.NewRequest(http.MethodPost, "/oauth/github", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestGitHubOAuthHandler_Success(t *testing.T) {
	cfg := &appconf.Cfg{
		OAuth: appconf.OAuth{
			GoogleClientID:     "test-client-id",
			GitHubClientID:     "test-github-client-id",
			GitHubClientSecret: "test-github-client-secret",
		},
		EmailSender: appconf.EmailSender{ContactEmail: "contact@example.com"},
		Web: appconf.Web{
			RefreshCookieSameSite: "strict",
			RefreshCookieSecure:   true,
		},
	}
	_, mockGitHubOAuthClient, authAPI, mockUserFacade, app, ctrl := setupTest(t, cfg)
	defer ctrl.Finish()

	app.Post("/oauth/github", authAPI.GitHubOAuthHandler)

	t.Run("successful new user creation", func(t *testing.T) {
		mockGitHubOAuthClient.EXPECT().
			ExchangeCodeForUser(gomock.Any(), "mock-github-code").
			Return(githubapi.UserInfo{ID: "12345", Login: "ghuser", Email: "ghuser@example.com"}, nil)

		u := model.User{ID: "uid-1", Username: "ghuser", Email: "ghuser@example.com", OAuthProvider: "github", OAuthID: "12345"}
		mockUserFacade.EXPECT().
			GitHubOAuth(gomock.Any(), "12345", "ghuser@example.com", "ghuser").
			Return(u, nil)

		mockUserFacade.EXPECT().
			CreateTokens(gomock.Any(), gomock.Any()).
			Return(facade.TokenPair{
				AccessToken:  "test-jwt-token",
				RefreshToken: facade.RefreshToken{Token: "refresh-token"},
			}, nil)

		reqBody := auth.GitHubOAuthRequest{Code: "mock-github-code"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/github", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var response auth.TokenResp
		responseBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		err = json.Unmarshal(responseBody, &response)
		require.NoError(t, err)
		require.Equal(t, "test-jwt-token", response.AccessToken)
	})

	t.Run("successful existing user login", func(t *testing.T) {
		mockGitHubOAuthClient.EXPECT().
			ExchangeCodeForUser(gomock.Any(), "mock-github-code").
			Return(githubapi.UserInfo{ID: "12345", Login: "ghuser", Email: "ghuser@example.com"}, nil)

		u := model.User{ID: "uid-2", Username: "ghuser", OAuthProvider: "github", OAuthID: "12345"}
		mockUserFacade.EXPECT().
			GitHubOAuth(gomock.Any(), "12345", "ghuser@example.com", "ghuser").
			Return(u, nil)

		mockUserFacade.EXPECT().
			CreateTokens(gomock.Any(), gomock.Any()).
			Return(facade.TokenPair{
				AccessToken:  "test-jwt-token",
				RefreshToken: facade.RefreshToken{Token: "refresh-token"},
			}, nil)

		reqBody := auth.GitHubOAuthRequest{Code: "mock-github-code"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/github", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var response auth.TokenResp
		responseBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		err = json.Unmarshal(responseBody, &response)
		require.NoError(t, err)
		require.Equal(t, "test-jwt-token", response.AccessToken)
	})

	t.Run("username conflict", func(t *testing.T) {
		mockGitHubOAuthClient.EXPECT().
			ExchangeCodeForUser(gomock.Any(), "mock-github-code").
			Return(githubapi.UserInfo{ID: "99999", Login: "conflictuser", Email: "conflict@example.com"}, nil)

		mockUserFacade.EXPECT().
			GitHubOAuth(gomock.Any(), "99999", "conflict@example.com", "conflictuser").
			Return(model.User{}, facade.ErrOAuthSignInConflict)

		reqBody := auth.GitHubOAuthRequest{Code: "mock-github-code"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/github", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("publisher email conflict", func(t *testing.T) {
		mockGitHubOAuthClient.EXPECT().
			ExchangeCodeForUser(gomock.Any(), "mock-github-code").
			Return(githubapi.UserInfo{ID: "12345", Login: "ghuser", Email: "publisher@example.com"}, nil)

		mockUserFacade.EXPECT().
			GitHubOAuth(gomock.Any(), "12345", "publisher@example.com", "ghuser").
			Return(model.User{}, facade.ErrOAuthPublisherConflict)

		reqBody := auth.GitHubOAuthRequest{Code: "mock-github-code"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/github", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusConflict, resp.StatusCode)

		var response struct {
			Error string `json:"error"`
		}
		responseBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		err = json.Unmarshal(responseBody, &response)
		require.NoError(t, err)
		require.Equal(t, "Publisher account found. Please sign in with your username and password.", response.Error)
	})

	t.Run("invalid authorization code", func(t *testing.T) {
		mockGitHubOAuthClient.EXPECT().
			ExchangeCodeForUser(gomock.Any(), "invalid-code").
			Return(githubapi.UserInfo{}, errors.New("invalid code"))

		reqBody := auth.GitHubOAuthRequest{Code: "invalid-code"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/github", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("database error", func(t *testing.T) {
		mockGitHubOAuthClient.EXPECT().
			ExchangeCodeForUser(gomock.Any(), "mock-github-code").
			Return(githubapi.UserInfo{ID: "12345", Login: "ghuser", Email: "ghuser@example.com"}, nil)

		mockUserFacade.EXPECT().
			GitHubOAuth(gomock.Any(), "12345", "ghuser@example.com", "ghuser").
			Return(model.User{}, errors.New("database connection failed"))

		reqBody := auth.GitHubOAuthRequest{Code: "mock-github-code"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/github", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("token generation failure", func(t *testing.T) {
		mockGitHubOAuthClient.EXPECT().
			ExchangeCodeForUser(gomock.Any(), "mock-github-code").
			Return(githubapi.UserInfo{ID: "12345", Login: "ghuser", Email: "ghuser@example.com"}, nil)

		u := model.User{ID: "uid-3", Username: "ghuser", OAuthProvider: "github", OAuthID: "12345"}
		mockUserFacade.EXPECT().
			GitHubOAuth(gomock.Any(), "12345", "ghuser@example.com", "ghuser").
			Return(u, nil)

		mockUserFacade.EXPECT().
			CreateTokens(gomock.Any(), gomock.Any()).
			Return(facade.TokenPair{}, errors.New("token generation failed"))

		reqBody := auth.GitHubOAuthRequest{Code: "mock-github-code"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/github", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
