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
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSignUpHandler(t *testing.T) {
	tests := []struct {
		name                     string
		request                  interface{}
		emailVerificationEnabled bool
		setupMocks               func(*mocks.MockUserFacade)
		expectedStatus           int
		expectedResp             interface{}
	}{
		{
			name: "successful user signup",
			request: handlers.SignUpReq{
				Username:        "newuser",
				DisplayName:     "New User",
				Password:        "password123",
				ConfirmPassword: "password123",
				IsPublisher:     false,
			},
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				u := model.User{ID: "uid-1", Username: "newuser", DisplayName: "New User", Role: "user"}
				mockUserFacade.EXPECT().SignUp(
					gomock.Any(), "newuser", "New User", "", "password123", false,
				).Return(u, nil)
				mockUserFacade.EXPECT().
					CreateTokens(gomock.Any(), gomock.Any()).
					Return(facade.TokenPair{
						AccessToken:  "test-token",
						RefreshToken: facade.RefreshToken{Token: "refresh-token"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp:   handlers.TokenResp{AccessToken: "test-token"},
		},
		{
			name: "successful publisher signup",
			request: handlers.SignUpReq{
				Username:        "newpublisher",
				DisplayName:     "Publisher Co",
				Password:        "password123",
				ConfirmPassword: "password123",
				IsPublisher:     true,
			},
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				u := model.User{ID: "uid-2", Username: "newpublisher", DisplayName: "Publisher Co", Role: "publisher"}
				mockUserFacade.EXPECT().SignUp(
					gomock.Any(), "newpublisher", "Publisher Co", "", "password123", true,
				).Return(u, nil)
				mockUserFacade.EXPECT().
					CreateTokens(gomock.Any(), gomock.Any()).
					Return(facade.TokenPair{
						AccessToken:  "test-token",
						RefreshToken: facade.RefreshToken{Token: "refresh-token"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp:   handlers.TokenResp{AccessToken: "test-token"},
		},
		{
			name: "successful user signup with email verification",
			request: handlers.SignUpReq{
				Username:        "newuser_verify",
				DisplayName:     "New User Verify",
				Email:           "verify@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
				IsPublisher:     false,
			},
			emailVerificationEnabled: true,
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				u := model.User{ID: "uid-3", Username: "newuser_verify", DisplayName: "New User Verify", Email: "verify@example.com", Role: "user"}
				mockUserFacade.EXPECT().SignUp(
					gomock.Any(), "newuser_verify", "New User Verify", "verify@example.com", "password123", false,
				).Return(u, nil)
				mockUserFacade.EXPECT().
					CreateTokens(gomock.Any(), gomock.Any()).
					Return(facade.TokenPair{
						AccessToken:  "test-token",
						RefreshToken: facade.RefreshToken{Token: "refresh-token"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp:   handlers.TokenResp{AccessToken: "test-token"},
		},
		{
			name: "username already exists",
			request: handlers.SignUpReq{
				Username:        "existinguser",
				DisplayName:     "Existing User",
				Password:        "password123",
				ConfirmPassword: "password123",
				IsPublisher:     false,
			},
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().SignUp(
					gomock.Any(), "existinguser", "Existing User", "", "password123", false,
				).Return(model.User{}, facade.ErrSignUpUsernameExists)
			},
			expectedStatus: http.StatusConflict,
			expectedResp:   web.ErrResp{Error: "This username is already taken"},
		},
		{
			name: "publisher name already exists",
			request: handlers.SignUpReq{
				Username:        "newpublisher",
				DisplayName:     "Existing Publisher",
				Password:        "password123",
				ConfirmPassword: "password123",
				IsPublisher:     true,
			},
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().SignUp(
					gomock.Any(), "newpublisher", "Existing Publisher", "", "password123", true,
				).Return(model.User{}, facade.ErrSignUpPublisherNameExists)
			},
			expectedStatus: http.StatusConflict,
			expectedResp:   web.ErrResp{Error: "Publisher with this name already exists"},
		},
		{
			name: "database error on create",
			request: handlers.SignUpReq{
				Username:        "newuser",
				DisplayName:     "New User",
				Password:        "password123",
				ConfirmPassword: "password123",
				IsPublisher:     false,
			},
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().SignUp(
					gomock.Any(), "newuser", "New User", "", "password123", false,
				).Return(model.User{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   web.ErrResp{Error: internalErrorMsg},
		},
		{
			name:           "invalid request body",
			request:        "invalid json",
			setupMocks:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedResp:   web.ErrResp{Error: "Error parsing data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &appconf.Cfg{
				EmailSender: appconf.EmailSender{},
				Auth:        appconf.Auth{GoogleClientID: "test-client-id"},
			}
			_, authAPI, mockUserFacade, app, ctrl := setupTest(t, cfg)
			defer ctrl.Finish()

			if tt.setupMocks != nil {
				tt.setupMocks(mockUserFacade)
			}

			app.Post("/signup", authAPI.SignUpHandler)

			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, 5000)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			switch v := tt.expectedResp.(type) {
			case handlers.TokenResp:
				var actual handlers.TokenResp
				err = json.Unmarshal(body, &actual)
				require.NoError(t, err)
				if v.AccessToken != "" {
					assert.Equal(t, v.AccessToken, actual.AccessToken)
				}
			case web.ErrResp:
				var actual web.ErrResp
				err = json.Unmarshal(body, &actual)
				require.NoError(t, err)
				assert.Equal(t, v.Error, actual.Error)
			}
		})
	}
}
