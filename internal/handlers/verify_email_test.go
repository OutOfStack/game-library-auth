package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestVerifyEmailHandler(t *testing.T) {
	userID := uuid.New().String()
	code := "123456"

	tests := []struct {
		name           string
		request        interface{}
		authHeader     string
		setupMocks     func(*mocks.MockUserFacade)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name: "successful verification",
			request: handlers.VerifyEmailReq{
				Code: code,
			},
			authHeader: "Bearer valid-token",
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				u := model.User{ID: userID, Username: "testuser", Email: "test@example.com", EmailVerified: true}
				mockUserFacade.EXPECT().VerifyEmail(gomock.Any(), userID, code).Return(u, nil)
				mockUserFacade.EXPECT().CreateTokens(gomock.Any(), u).Return(facade.TokenPair{
					AccessToken:  "new.jwt.token",
					RefreshToken: facade.RefreshToken{Token: "new.refresh.token"},
				}, nil)
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
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().VerifyEmail(gomock.Any(), userID, code).Return(model.User{}, facade.ErrVerifyEmailInvalidOrExpired)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   web.ErrResp{Error: "Invalid or expired verification code"},
		},
		{
			name: "user email already verified",
			request: handlers.VerifyEmailReq{
				Code: code,
			},
			authHeader: "Bearer valid-token",
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().VerifyEmail(gomock.Any(), userID, code).Return(model.User{}, facade.ErrVerifyEmailAlreadyVerified)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   web.ErrResp{Error: "Email is already verified"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, authAPI, mockUserFacade, app, ctrl := setupTest(t, nil)
			defer ctrl.Finish()

			if tt.authHeader == "Bearer valid-token" {
				mockUserFacade.EXPECT().
					GetClaimsFromAccessToken("valid-token").
					Return(auth.Claims{UserID: userID}, nil).
					AnyTimes()
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockUserFacade)
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
