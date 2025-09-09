package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/database"
	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const invalidAuthToken = "Invalid or missing authorization token"

func TestResendVerificationEmailHandler(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                     string
		authHeader               string
		emailVerificationEnabled bool
		setupMocks               func(*mocks.MockAuth, *mocks.MockUserRepo, *mocks.MockEmailSender)
		expectedStatus           int
		expectedResp             interface{}
	}{
		{
			name:                     "successful resend",
			authHeader:               "Bearer valid-token",
			emailVerificationEnabled: true,
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo, mockEmailSender *mocks.MockEmailSender) {
				claims := auth.Claims{UserID: userID.String()}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)

				user := database.User{ID: userID, Username: "testuser"}
				user.SetEmail("test@example.com", false)
				mockUserRepo.EXPECT().GetUserByID(gomock.Any(), userID.String()).Return(user, nil)

				// Mock the verification flow
				verification := database.EmailVerification{ID: uuid.New()}
				mockUserRepo.EXPECT().GetEmailVerificationByUserID(gomock.Any(), userID).Return(verification, nil).AnyTimes()
				mockUserRepo.EXPECT().SetEmailVerificationUsed(gomock.Any(), verification.ID, false).Return(nil).AnyTimes()
				mockUserRepo.EXPECT().CreateEmailVerification(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				mockEmailSender.EXPECT().SendEmailVerification(gomock.Any(), gomock.Any()).Return("msg-123", nil).AnyTimes()
				mockUserRepo.EXPECT().SetEmailVerificationMessageID(gomock.Any(), gomock.Any(), "msg-123").Return(nil).AnyTimes()
			},
			expectedStatus: http.StatusNoContent,
			expectedResp:   nil,
		},
		{
			name:                     "email already verified",
			authHeader:               "Bearer valid-token",
			emailVerificationEnabled: true,
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo, _ *mocks.MockEmailSender) {
				claims := auth.Claims{UserID: userID.String()}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)

				user := database.User{ID: userID, Username: "testuser"}
				user.SetEmail("test@example.com", true)
				mockUserRepo.EXPECT().GetUserByID(gomock.Any(), userID.String()).Return(user, nil)
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
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo, _ *mocks.MockEmailSender) {
				claims := auth.Claims{UserID: userID.String()}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)

				mockUserRepo.EXPECT().GetUserByID(gomock.Any(), userID.String()).Return(database.User{}, database.ErrNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   web.ErrResp{Error: internalErrorMsg},
		},
		{
			name:                     "user has no email address",
			authHeader:               "Bearer valid-token",
			emailVerificationEnabled: true,
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo, _ *mocks.MockEmailSender) {
				claims := auth.Claims{UserID: userID.String()}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)

				user := database.User{ID: userID, Username: "testuser"}
				// User has no email set (Email.Valid is false)
				mockUserRepo.EXPECT().GetUserByID(gomock.Any(), userID.String()).Return(user, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   web.ErrResp{Error: "User does not have an email address"},
		},
		{
			name:                     "email sending failure",
			authHeader:               "Bearer valid-token",
			emailVerificationEnabled: true,
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo, mockEmailSender *mocks.MockEmailSender) {
				claims := auth.Claims{UserID: userID.String()}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)

				user := database.User{ID: userID, Username: "testuser"}
				user.SetEmail("test@example.com", false)
				mockUserRepo.EXPECT().GetUserByID(gomock.Any(), userID.String()).Return(user, nil)

				verification := database.EmailVerification{ID: uuid.New()}
				mockUserRepo.EXPECT().GetEmailVerificationByUserID(gomock.Any(), userID).Return(verification, nil).AnyTimes()
				mockUserRepo.EXPECT().SetEmailVerificationUsed(gomock.Any(), verification.ID, false).Return(nil).AnyTimes()
				mockUserRepo.EXPECT().CreateEmailVerification(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				mockEmailSender.EXPECT().SendEmailVerification(gomock.Any(), gomock.Any()).Return("", errors.New("email service unavailable")).AnyTimes()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   web.ErrResp{Error: internalErrorMsg},
		},
		{
			name:                     "database error on GetUserByID",
			authHeader:               "Bearer valid-token",
			emailVerificationEnabled: true,
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo, _ *mocks.MockEmailSender) {
				claims := auth.Claims{UserID: userID.String()}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)

				mockUserRepo.EXPECT().GetUserByID(gomock.Any(), userID.String()).Return(database.User{}, errors.New("database connection failed"))
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
			mockAuth, mockUserRepo, _, mockEmailSender, authAPI, app, ctrl := setupTest(t, cfg)
			defer ctrl.Finish()

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockUserRepo, mockEmailSender)
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
