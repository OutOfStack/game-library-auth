package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
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
		setupMocks               func(*mocks.MockAuth, *mocks.MockUserRepo, *mocks.MockEmailSender)
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
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo, _ *mocks.MockEmailSender) {
				mockUserRepo.EXPECT().
					GetUserByUsername(gomock.Any(), "newuser").
					Return(database.User{}, database.ErrNotFound)

				mockUserRepo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, user database.User) error {
						assert.Equal(t, "newuser", user.Username)
						assert.Equal(t, "New User", user.DisplayName)
						assert.Equal(t, database.UserRoleName, user.Role)
						return nil
					})
				mockAuth.EXPECT().CreateClaims(gomock.Any()).Return(nil)
				mockAuth.EXPECT().GenerateToken(gomock.Any()).Return("test-token", nil)
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
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo, _ *mocks.MockEmailSender) {
				mockUserRepo.EXPECT().
					GetUserByUsername(gomock.Any(), "newpublisher").
					Return(database.User{}, database.ErrNotFound)

				mockUserRepo.EXPECT().
					CheckUserExists(gomock.Any(), "Publisher Co", database.PublisherRoleName).
					Return(false, nil)

				mockUserRepo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, user database.User) error {
						assert.Equal(t, "newpublisher", user.Username)
						assert.Equal(t, "Publisher Co", user.DisplayName)
						assert.Equal(t, database.PublisherRoleName, user.Role)
						return nil
					})
				mockAuth.EXPECT().CreateClaims(gomock.Any()).Return(nil)
				mockAuth.EXPECT().GenerateToken(gomock.Any()).Return("test-token", nil)
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
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo, mockEmailSender *mocks.MockEmailSender) {
				mockUserRepo.EXPECT().
					GetUserByUsername(gomock.Any(), "newuser_verify").
					Return(database.User{}, database.ErrNotFound)

				mockUserRepo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, user database.User) error {
						assert.Equal(t, "newuser_verify", user.Username)
						assert.Equal(t, "New User Verify", user.DisplayName)
						assert.Equal(t, database.UserRoleName, user.Role)
						assert.Equal(t, "verify@example.com", user.Email.String)
						assert.False(t, user.EmailVerified)
						return nil
					})

				// Mock the verification flow
				mockUserRepo.EXPECT().GetEmailVerificationByUserID(gomock.Any(), gomock.Any()).Return(database.EmailVerification{}, database.ErrNotFound)
				mockUserRepo.EXPECT().CreateEmailVerification(gomock.Any(), gomock.Any()).Return(nil)
				mockEmailSender.EXPECT().SendEmailVerification(gomock.Any(), gomock.Any()).Return("msg-456", nil)
				mockUserRepo.EXPECT().SetEmailVerificationMessageID(gomock.Any(), gomock.Any(), "msg-456").Return(nil)

				mockAuth.EXPECT().CreateClaims(gomock.Any()).Return(nil)
				mockAuth.EXPECT().GenerateToken(gomock.Any()).Return("test-token", nil)
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
			setupMocks: func(_ *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo, _ *mocks.MockEmailSender) {
				mockUserRepo.EXPECT().
					GetUserByUsername(gomock.Any(), "existinguser").
					Return(database.User{Username: "existinguser"}, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedResp: web.ErrResp{
				Error: "This username is already taken",
			},
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
			setupMocks: func(_ *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo, _ *mocks.MockEmailSender) {
				mockUserRepo.EXPECT().
					GetUserByUsername(gomock.Any(), "newpublisher").
					Return(database.User{}, database.ErrNotFound)

				mockUserRepo.EXPECT().
					CheckUserExists(gomock.Any(), "Existing Publisher", database.PublisherRoleName).
					Return(true, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedResp: web.ErrResp{
				Error: "Publisher with this name already exists",
			},
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
			setupMocks: func(_ *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo, _ *mocks.MockEmailSender) {
				mockUserRepo.EXPECT().
					GetUserByUsername(gomock.Any(), "newuser").
					Return(database.User{}, database.ErrNotFound)

				mockUserRepo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp: web.ErrResp{
				Error: internalErrorMsg,
			},
		},
		{
			name:           "invalid request body",
			request:        "invalid json",
			setupMocks:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedResp: web.ErrResp{
				Error: "Error parsing data",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &appconf.Cfg{
				EmailSender: appconf.EmailSender{
					EmailVerificationEnabled: tt.emailVerificationEnabled,
				},
				Auth: appconf.Auth{
					GoogleClientID: "test-client-id",
				},
			}
			mockAuth, mockUserRepo, _, mockEmailSender, authAPI, app, ctrl := setupTest(t, cfg)
			defer ctrl.Finish()

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockUserRepo, mockEmailSender)
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
