package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyToken(t *testing.T) {
	tests := []struct {
		name           string
		request        handlers.VerifyTokenReq
		setupMocks     func(*mocks.MockAuth, *mocks.MockStorage)
		expectedStatus int
		expectedResp   handlers.VerifyTokenResp
	}{
		{
			name: "valid token",
			request: handlers.VerifyTokenReq{
				Token: "valid.jwt.token",
			},
			setupMocks: func(mockAuth *mocks.MockAuth, _ *mocks.MockStorage) {
				mockAuth.EXPECT().
					ValidateToken("valid.jwt.token").
					Return(auth.Claims{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: handlers.VerifyTokenResp{
				Valid: true,
			},
		},
		{
			name: "invalid token",
			request: handlers.VerifyTokenReq{
				Token: "invalid.jwt.token",
			},
			setupMocks: func(mockAuth *mocks.MockAuth, _ *mocks.MockStorage) {
				mockAuth.EXPECT().
					ValidateToken("invalid.jwt.token").
					Return(auth.Claims{}, errors.New("token validation error"))
			},
			expectedStatus: http.StatusOK,
			expectedResp: handlers.VerifyTokenResp{
				Valid: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth, mockStorage, _, authAPI, app, ctrl := setupTest(t, nil)
			defer ctrl.Finish()

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockStorage)
			}

			app.Post("/token/verify", authAPI.VerifyTokenHandler)

			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/token/verify", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var actual handlers.VerifyTokenResp
			err = json.Unmarshal(body, &actual)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResp, actual)
		})
	}
}
