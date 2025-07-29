package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
)

func TestGoogleOAuthHandler_InvalidRequest(t *testing.T) {
	logger := zap.NewNop()

	t.Run("invalid request body", func(t *testing.T) {
		cfg := &appconf.Cfg{
			Auth: appconf.Auth{
				GoogleClientID: "test-client-id",
			},
		}
		api, err := handlers.NewAuthAPI(logger, cfg, nil, nil, nil)
		require.NoError(t, err)
		app := fiber.New()
		app.Post("/oauth/google", api.GoogleOAuthHandler)

		req := httptest.NewRequest(http.MethodPost, "/oauth/google", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing google client id", func(t *testing.T) {
		cfg := &appconf.Cfg{
			Auth: appconf.Auth{
				GoogleClientID: "",
			},
		}
		_, err := handlers.NewAuthAPI(logger, cfg, nil, nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "google client id is empty")
	})
}

func TestGoogleOAuthHandler_Success(t *testing.T) {
	cfg := &appconf.Cfg{
		Auth: appconf.Auth{
			GoogleClientID: "test-client-id",
		},
	}
	mockAuth, mockUserRepo, mockGoogleTokenValidator, authAPI, app, ctrl := setupTest(t, cfg)
	defer ctrl.Finish()

	app.Post("/oauth/google", authAPI.GoogleOAuthHandler)

	t.Run("successful new user creation", func(t *testing.T) {
		// Mock Google token validation
		mockPayload := &idtoken.Payload{
			Subject: "google-sub-id",
			Claims: map[string]interface{}{
				"email":          "test@example.com",
				"name":           "Test User",
				"email_verified": true,
			},
		}

		mockGoogleTokenValidator.EXPECT().
			Validate(gomock.Any(), "mock-google-id-token", "test-client-id").
			Return(mockPayload, nil)

		// Mock user repo - user not found (new user)
		mockUserRepo.EXPECT().
			GetUserByOAuth(gomock.Any(), "google", "google-sub-id").
			Return(database.User{}, database.ErrNotFound)

		// Mock user creation
		mockUserRepo.EXPECT().
			CreateUser(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ interface{}, user database.User) error {
				// Verify user data from Google OAuth
				require.Equal(t, "test@example.com", user.Username) // Username is full email
				require.Equal(t, "Test User", user.DisplayName)
				require.Equal(t, "google", user.OAuthProvider.String)
				require.Equal(t, "google-sub-id", user.OAuthID.String)
				return nil
			})

		mockAuth.EXPECT().
			CreateClaims(gomock.Any()).
			Return(nil)

		mockAuth.EXPECT().
			GenerateToken(gomock.Any()).
			Return("test-jwt-token", nil)

		reqBody := handlers.GoogleOAuthRequest{
			IDToken: "mock-google-id-token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var response handlers.TokenResp
		responseBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		err = json.Unmarshal(responseBody, &response)
		require.NoError(t, err)
		require.Equal(t, "test-jwt-token", response.AccessToken)
	})

	t.Run("successful existing user login", func(t *testing.T) {
		existingUser := database.NewUser(
			"existing",
			"Existing User",
			nil,
			database.UserRoleName)
		existingUser.SetOAuthID(auth.GoogleAuthTokenProvider, "google-sub-id")

		// Mock Google token validation
		mockPayload := &idtoken.Payload{
			Subject: "google-sub-id",
			Claims: map[string]interface{}{
				"email":          "existing@example.com",
				"name":           "Existing User",
				"email_verified": true,
			},
		}

		mockGoogleTokenValidator.EXPECT().
			Validate(gomock.Any(), "mock-google-id-token", "test-client-id").
			Return(mockPayload, nil)

		// Mock user repo - user found
		mockUserRepo.EXPECT().
			GetUserByOAuth(gomock.Any(), "google", "google-sub-id").
			Return(existingUser, nil)

		// Mock JWT generation
		mockAuth.EXPECT().
			CreateClaims(existingUser).
			Return(nil)

		mockAuth.EXPECT().
			GenerateToken(gomock.Any()).
			Return("test-jwt-token", nil)

		reqBody := handlers.GoogleOAuthRequest{
			IDToken: "mock-google-id-token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var response handlers.TokenResp
		responseBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		err = json.Unmarshal(responseBody, &response)
		require.NoError(t, err)
		require.Equal(t, "test-jwt-token", response.AccessToken)
	})

	t.Run("username conflict - user already exists", func(t *testing.T) {
		// Mock Google token validation
		mockPayload := &idtoken.Payload{
			Subject: "new-google-sub-id",
			Claims: map[string]interface{}{
				"email":          "conflict@example.com",
				"name":           "Conflict User",
				"email_verified": true,
			},
		}

		mockGoogleTokenValidator.EXPECT().
			Validate(gomock.Any(), "mock-google-id-token", "test-client-id").
			Return(mockPayload, nil)

		// Mock user repo - OAuth user not found (new OAuth user)
		mockUserRepo.EXPECT().
			GetUserByOAuth(gomock.Any(), "google", "new-google-sub-id").
			Return(database.User{}, database.ErrNotFound)

		// Mock user creation - username already exists (constraint violation)
		mockUserRepo.EXPECT().
			CreateUser(gomock.Any(), gomock.Any()).
			Return(database.ErrUsernameExists)

		reqBody := handlers.GoogleOAuthRequest{
			IDToken: "mock-google-id-token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/google", bytes.NewReader(body))
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
		require.Equal(t, "Username already exists, please sign up with registration form", response.Error)
	})

	t.Run("invalid google token", func(t *testing.T) {
		// Mock Google token validation failure
		mockGoogleTokenValidator.EXPECT().
			Validate(gomock.Any(), "invalid-token", "test-client-id").
			Return(nil, errors.New("invalid token"))

		reqBody := handlers.GoogleOAuthRequest{
			IDToken: "invalid-token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
