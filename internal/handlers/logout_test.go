package handlers_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLogoutHandler(t *testing.T) {
	tests := []struct {
		name           string
		cookieValue    string
		setupMocks     func(*mocks.MockUserFacade)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name:        "successful logout with refresh token",
			cookieValue: "valid-refresh-token",
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().
					RevokeRefreshToken(gomock.Any(), "valid-refresh-token").
					Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedResp:   nil,
		},
		{
			name:        "successful logout without refresh token",
			cookieValue: "",
			setupMocks: func(_ *mocks.MockUserFacade) {
			},
			expectedStatus: http.StatusNoContent,
			expectedResp:   nil,
		},
		{
			name:        "internal server error on revoke",
			cookieValue: "some-token",
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().
					RevokeRefreshToken(gomock.Any(), "some-token").
					Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp: web.ErrResp{
				Error: internalErrorMsg,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, authAPI, mockUserFacade, app, ctrl := setupTest(t, nil)
			defer ctrl.Finish()

			tt.setupMocks(mockUserFacade)

			app.Post("/logout", authAPI.LogoutHandler)

			req := httptest.NewRequest(http.MethodPost, "/logout", nil)
			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: tt.cookieValue,
				})
			}

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedResp != nil {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var errResp web.ErrResp
				err = json.Unmarshal(body, &errResp)
				require.NoError(t, err)
				expected, ok := tt.expectedResp.(web.ErrResp)
				require.True(t, ok)
				assert.Equal(t, expected.Error, errResp.Error)
			}

			cookies := resp.Cookies()
			var refreshTokenCookie *http.Cookie
			for _, cookie := range cookies {
				if cookie.Name == "refresh_token" {
					refreshTokenCookie = cookie
					break
				}
			}

			if tt.expectedStatus == http.StatusNoContent {
				require.NotNil(t, refreshTokenCookie, "refresh_token cookie should be set to clear it")
				assert.Empty(t, refreshTokenCookie.Value, "cookie value should be empty")
				assert.True(t, refreshTokenCookie.Expires.Before(time.Now()), "cookie should be expired")
			}
		})
	}
}
