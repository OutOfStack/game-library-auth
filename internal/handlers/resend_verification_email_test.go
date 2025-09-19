package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const invalidAuthToken = "Invalid or missing authorization token"

func TestResendVerificationEmailHandler(t *testing.T) {
	userID := uuid.New().String()

	tests := []struct {
		name                     string
		authHeader               string
		emailVerificationEnabled bool
		setupMocks               func(*mocks.MockAuth, *mocks.MockUserFacade)
		expectedStatus           int
		expectedResp             interface{}
	}{
		{
			name:                     "successful resend",
			authHeader:               "Bearer valid-token",
			emailVerificationEnabled: true,
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserFacade *mocks.MockUserFacade) {
				claims := auth.Claims{UserID: userID}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)
				mockUserFacade.EXPECT().ResendVerificationEmail(gomock.Any(), userID).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedResp:   nil,
		},
		{
			name:                     "email already verified",
			authHeader:               "Bearer valid-token",
			emailVerificationEnabled: true,
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserFacade *mocks.MockUserFacade) {
				claims := auth.Claims{UserID: userID}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)
				mockUserFacade.EXPECT().ResendVerificationEmail(gomock.Any(), userID).Return(facade.ErrVerifyEmailAlreadyVerified)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   web.ErrResp{Error: "Email is already verified"},
		},
		{
			name:           "missing auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedResp:   web.ErrResp{Error: invalidAuthToken},
		},
		{
			name:                     "user not found",
			authHeader:               "Bearer valid-token",
			emailVerificationEnabled: true,
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserFacade *mocks.MockUserFacade) {
				claims := auth.Claims{UserID: userID}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)
				mockUserFacade.EXPECT().ResendVerificationEmail(gomock.Any(), userID).Return(errors.New("not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   web.ErrResp{Error: internalErrorMsg},
		},
		{
			name:                     "user has no email address",
			authHeader:               "Bearer valid-token",
			emailVerificationEnabled: true,
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserFacade *mocks.MockUserFacade) {
				claims := auth.Claims{UserID: userID}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)
				mockUserFacade.EXPECT().ResendVerificationEmail(gomock.Any(), userID).Return(facade.ErrResendVerificationNoEmail)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   web.ErrResp{Error: "User does not have an email address"},
		},
		{
			name:                     "email sending failure",
			authHeader:               "Bearer valid-token",
			emailVerificationEnabled: true,
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserFacade *mocks.MockUserFacade) {
				claims := auth.Claims{UserID: userID}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)
				mockUserFacade.EXPECT().ResendVerificationEmail(gomock.Any(), userID).Return(errors.New("email service unavailable"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   web.ErrResp{Error: internalErrorMsg},
		},
		{
			name:                     "database error on GetUserByID",
			authHeader:               "Bearer valid-token",
			emailVerificationEnabled: true,
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserFacade *mocks.MockUserFacade) {
				claims := auth.Claims{UserID: userID}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)
				mockUserFacade.EXPECT().ResendVerificationEmail(gomock.Any(), userID).Return(errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   web.ErrResp{Error: internalErrorMsg},
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
			mockAuth, _, authAPI, mockUserFacade, app, ctrl := setupTest(t, cfg)
			defer ctrl.Finish()

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockUserFacade)
			}

			app.Post("/resend-verification", authAPI.ResendVerificationEmailHandler)

			req := httptest.NewRequest(http.MethodPost, "/resend-verification", nil)
			req.Header.Set("Authorization", tt.authHeader)

			resp, err := app.Test(req, 5000)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedResp != nil {
				var actual web.ErrResp
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&actual))

				expected, ok := tt.expectedResp.(web.ErrResp)
				require.True(t, ok)
				assert.Equal(t, expected.Error, actual.Error)
			}
		})
	}
}
