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
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
)

func TestGoogleOAuthHandler_InvalidRequest(t *testing.T) {
	logger := zap.NewNop()

	t.Run("invalid request body", func(t *testing.T) {
		cfg := &appconf.Cfg{Auth: appconf.Auth{GoogleClientID: "test-client-id"}}
		mockAuth := mocks.NewMockAuth(gomock.NewController(t))
		api, err := handlers.NewAuthAPI(logger, cfg, mockAuth, nil, nil)
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
		cfg := &appconf.Cfg{Auth: appconf.Auth{GoogleClientID: ""}}
		_, err := handlers.NewAuthAPI(logger, cfg, nil, nil, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "google client id is empty")
	})
}

func TestGoogleOAuthHandler_Success(t *testing.T) {
	cfg := &appconf.Cfg{Auth: appconf.Auth{GoogleClientID: "test-client-id"}}
	mockAuth, mockGoogleTokenValidator, authAPI, mockUserFacade, app, ctrl := setupTest(t, cfg)
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

		// Mock facade Google OAuth
		u := model.User{ID: "uid-1", Username: "test", Email: "test@example.com", OAuthProvider: "google", OAuthID: "google-sub-id"}
		mockUserFacade.EXPECT().
			GoogleOAuth(gomock.Any(), "google-sub-id", "test@example.com").
			Return(u, nil)

		mockAuth.EXPECT().
			CreateUserClaims(gomock.Eq(u)).
			Return(jwt.MapClaims{"sub": u.ID})

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
		u := model.User{ID: "uid-2", Username: "existing", OAuthProvider: "google", OAuthID: "google-sub-id"}

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

		// Mock facade - user found
		mockUserFacade.EXPECT().
			GoogleOAuth(gomock.Any(), "google-sub-id", "existing@example.com").
			Return(u, nil)

		mockAuth.EXPECT().
			CreateUserClaims(gomock.Eq(u)).
			Return(jwt.MapClaims{"sub": u.ID})

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

		// Facade returns name conflict
		mockUserFacade.EXPECT().
			GoogleOAuth(gomock.Any(), "new-google-sub-id", "conflict@example.com").
			Return(model.User{}, facade.ErrOAuthSignInConflict)

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
		require.Equal(t, "Account setup incomplete. Please complete registration manually.", response.Error)
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

	t.Run("invalid email format in token", func(t *testing.T) {
		// Mock Google token validation with invalid email
		mockPayload := &idtoken.Payload{
			Subject: "google-sub-id",
			Claims: map[string]interface{}{
				"email":          "invalid-email",
				"name":           "Test User",
				"email_verified": true,
			},
		}

		mockGoogleTokenValidator.EXPECT().
			Validate(gomock.Any(), "mock-google-id-token", "test-client-id").
			Return(mockPayload, nil)

		// Mock facade returns invalid email error
		mockUserFacade.EXPECT().
			GoogleOAuth(gomock.Any(), "google-sub-id", "invalid-email").
			Return(model.User{}, facade.ErrInvalidEmail)

		reqBody := handlers.GoogleOAuthRequest{
			IDToken: "mock-google-id-token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response struct {
			Error string `json:"error"`
		}
		responseBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		err = json.Unmarshal(responseBody, &response)
		require.NoError(t, err)
		require.Equal(t, "Invalid email", response.Error)
	})

	t.Run("database error on GetUserByOAuth", func(t *testing.T) {
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

		// Mock facade error
		mockUserFacade.EXPECT().
			GoogleOAuth(gomock.Any(), "google-sub-id", "test@example.com").
			Return(model.User{}, errors.New("database connection failed"))

		reqBody := handlers.GoogleOAuthRequest{
			IDToken: "mock-google-id-token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("JWT generation failure", func(t *testing.T) {
		u2 := model.User{ID: "uid-3", Username: "existing", OAuthProvider: "google", OAuthID: "google-sub-id"}

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

		mockUserFacade.EXPECT().
			GoogleOAuth(gomock.Any(), "google-sub-id", "existing@example.com").
			Return(u2, nil)

		// Mock JWT generation failure
		mockAuth.EXPECT().
			CreateUserClaims(gomock.Eq(u2)).
			Return(jwt.MapClaims{"sub": u2.ID})

		mockAuth.EXPECT().
			GenerateToken(gomock.Any()).
			Return("", errors.New("token generation failed"))

		reqBody := handlers.GoogleOAuthRequest{
			IDToken: "mock-google-id-token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/oauth/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
