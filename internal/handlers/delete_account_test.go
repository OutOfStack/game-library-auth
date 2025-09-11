package handlers_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	auth_ "github.com/OutOfStack/game-library-auth/internal/auth"
	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDeleteAccountHandler(t *testing.T) {
	userID := uuid.New().String()

	tests := []struct {
		name           string
		authHeader     string
		setupMocks     func(*mocks.MockUserRepo)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name:       "successful delete",
			authHeader: "Bearer valid-token",
			setupMocks: func(mockUserRepo *mocks.MockUserRepo) {
				mockUserRepo.EXPECT().
					DeleteUser(gomock.Any(), userID).
					Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:       "user not found - still succeeds",
			authHeader: "Bearer valid-token",
			setupMocks: func(mockUserRepo *mocks.MockUserRepo) {
				mockUserRepo.EXPECT().
					DeleteUser(gomock.Any(), userID).
					Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedResp:   nil,
		},
		{
			name:       "user repo error on delete",
			authHeader: "Bearer valid-token",
			setupMocks: func(mockUserRepo *mocks.MockUserRepo) {
				mockUserRepo.EXPECT().
					DeleteUser(gomock.Any(), userID).
					Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp: web.ErrResp{
				Error: internalErrorMsg,
			},
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			setupMocks:     func(*mocks.MockUserRepo) {},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: web.ErrResp{
				Error: "Invalid or missing authorization token",
			},
		},
		{
			name:           "invalid authorization header format",
			authHeader:     "InvalidFormat",
			setupMocks:     func(*mocks.MockUserRepo) {},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: web.ErrResp{
				Error: "Invalid or missing authorization token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth, mockUserRepo, _, _, authAPI, app, ctrl := setupTest(t, nil)
			defer ctrl.Finish()

			if tt.authHeader == "Bearer valid-token" {
				claims := auth_.Claims{UserID: userID}
				mockAuth.EXPECT().
					ValidateToken("valid-token").
					Return(claims, nil).
					AnyTimes()
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo)
			}

			app.Post("/delete_account", authAPI.DeleteAccountHandler)

			req := httptest.NewRequest(http.MethodPost, "/delete_account", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedResp != nil {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var actual web.ErrResp
				err = json.Unmarshal(body, &actual)
				require.NoError(t, err)
				if expected, ok := tt.expectedResp.(web.ErrResp); ok {
					assert.Equal(t, expected.Error, actual.Error)
				}
			}
		})
	}
}
