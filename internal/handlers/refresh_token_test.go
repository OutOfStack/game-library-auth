package handlers_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRefreshTokenHandler(t *testing.T) {
	tests := []struct {
		name           string
		cookieValue    string
		setupMocks     func(*mocks.MockUserFacade)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name:        "successful refresh",
			cookieValue: "valid-refresh-token",
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().
					RefreshTokens(gomock.Any(), "valid-refresh-token").
					Return(facade.TokenPair{
						AccessToken:  "new-access-token",
						RefreshToken: facade.RefreshToken{Token: "new-refresh-token"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: handlers.TokenResp{
				AccessToken: "new-access-token",
			},
		},
		{
			name:        "missing cookie",
			cookieValue: "",
			setupMocks: func(_ *mocks.MockUserFacade) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: web.ErrResp{
				Error: "Refresh token not found",
			},
		},
		{
			name:        "token not found in database",
			cookieValue: "invalid-token",
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().
					RefreshTokens(gomock.Any(), "invalid-token").
					Return(facade.TokenPair{}, facade.ErrRefreshTokenNotFound)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: web.ErrResp{
				Error: "Invalid refresh token",
			},
		},
		{
			name:        "expired token",
			cookieValue: "expired-token",
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().
					RefreshTokens(gomock.Any(), "expired-token").
					Return(facade.TokenPair{}, facade.ErrRefreshTokenExpired)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: web.ErrResp{
				Error: "Refresh token expired",
			},
		},
		{
			name:        "internal server error",
			cookieValue: "some-token",
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().
					RefreshTokens(gomock.Any(), "some-token").
					Return(facade.TokenPair{}, errors.New("database error"))
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

			app.Post("/refresh", authAPI.RefreshTokenHandler)

			req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
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

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			switch expected := tt.expectedResp.(type) {
			case handlers.TokenResp:
				var tokenResp handlers.TokenResp
				err = json.Unmarshal(body, &tokenResp)
				require.NoError(t, err)
				assert.Equal(t, expected.AccessToken, tokenResp.AccessToken)
			case web.ErrResp:
				var errResp web.ErrResp
				err = json.Unmarshal(body, &errResp)
				require.NoError(t, err)
				assert.Equal(t, expected.Error, errResp.Error)
			}
		})
	}
}
