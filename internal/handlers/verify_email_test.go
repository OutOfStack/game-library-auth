package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestVerifyEmailHandler(t *testing.T) {
	userID := uuid.New()
	code := "123456"
	codeHash, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)

	tests := []struct {
		name           string
		request        interface{}
		authHeader     string
		setupMocks     func(*mocks.MockAuth, *mocks.MockUserRepo)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name: "successful verification",
			request: handlers.VerifyEmailReq{
				Code: code,
			},
			authHeader: "Bearer valid-token",
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo) {
				claims := auth.Claims{UserID: userID.String()}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)

				user := database.User{ID: userID, Username: "testuser"}
				user.SetEmail("test@example.com", false)
				mockUserRepo.EXPECT().GetUserByID(gomock.Any(), userID.String()).Return(user, nil)

				verification := database.NewEmailVerification(userID, "test@example.com", string(codeHash), time.Now().Add(1*time.Hour))
				mockUserRepo.EXPECT().GetEmailVerificationByUserID(gomock.Any(), userID).Return(verification, nil)

				mockUserRepo.EXPECT().UpdateUserEmail(gomock.Any(), userID, "test@example.com", true).Return(nil)
				mockUserRepo.EXPECT().SetEmailVerificationUsed(gomock.Any(), verification.ID, true).Return(nil)

				mockAuth.EXPECT().CreateClaims(gomock.Any()).Return(jwt.MapClaims{})
				mockAuth.EXPECT().GenerateToken(gomock.Any()).Return("new.jwt.token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp:   handlers.TokenResp{AccessToken: "new.jwt.token"},
		},
		{
			name: "missing auth header",
			request: handlers.VerifyEmailReq{
				Code: code,
			},
			authHeader:     "",
			setupMocks:     nil,
			expectedStatus: http.StatusUnauthorized,
			expectedResp:   web.ErrResp{Error: "Invalid or missing authorization token"},
		},
		{
			name: "expired verification code",
			request: handlers.VerifyEmailReq{
				Code: code,
			},
			authHeader: "Bearer valid-token",
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo) {
				claims := auth.Claims{UserID: userID.String()}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)

				user := database.User{ID: userID, Username: "testuser"}
				user.SetEmail("test@example.com", false)
				mockUserRepo.EXPECT().GetUserByID(gomock.Any(), userID.String()).Return(user, nil)

				verification := database.NewEmailVerification(userID, "test@example.com", string(codeHash), time.Now().Add(-1*time.Hour))
				mockUserRepo.EXPECT().GetEmailVerificationByUserID(gomock.Any(), userID).Return(verification, nil)
				mockUserRepo.EXPECT().SetEmailVerificationUsed(gomock.Any(), verification.ID, false).Return(nil)
			},
			expectedStatus: http.StatusGone,
			expectedResp:   web.ErrResp{Error: "Verification code has expired"},
		},
		{
			name: "user email already verified",
			request: handlers.VerifyEmailReq{
				Code: code,
			},
			authHeader: "Bearer valid-token",
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo) {
				claims := auth.Claims{UserID: userID.String()}
				mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)

				user := database.User{ID: userID, Username: "testuser"}
				user.SetEmail("test@example.com", true) // Already verified
				mockUserRepo.EXPECT().GetUserByID(gomock.Any(), userID.String()).Return(user, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   web.ErrResp{Error: "Email is already verified"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth, mockUserRepo, _, _, authAPI, app, ctrl := setupTest(t, nil)
			defer ctrl.Finish()

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockUserRepo)
			}

			app.Post("/verify-email", authAPI.VerifyEmailHandler)

			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/verify-email", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			resp, err := app.Test(req, 5000)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedResp != nil {
				var actual interface{}
				if _, ok := tt.expectedResp.(web.ErrResp); ok {
					actual = &web.ErrResp{}
				} else {
					actual = &handlers.TokenResp{}
				}
				require.NoError(t, json.NewDecoder(resp.Body).Decode(actual))

				switch expected := tt.expectedResp.(type) {
				case web.ErrResp:
					actual, ok := actual.(*web.ErrResp)
					require.True(t, ok)
					assert.Equal(t, expected.Error, actual.Error)
				case handlers.TokenResp:
					actual, ok := actual.(*handlers.TokenResp)
					require.True(t, ok)
					assert.Equal(t, expected.AccessToken, actual.AccessToken)
				}
			}
		})
	}
}
