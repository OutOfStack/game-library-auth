package auth_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/api/auth"
	mocks "github.com/OutOfStack/game-library-auth/internal/api/auth/mocks"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSignInHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        auth.SignInReq
		setupMocks     func(*mocks.MockUserFacade)
		expectedStatus int
		expectedResp   any
	}{
		{
			name: "successful sign in",
			request: auth.SignInReq{
				Username: "testuser",
				Password: "password123",
			},
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				u := model.User{ID: "uid-1", Username: "testuser"}

				mockUserFacade.EXPECT().
					SignIn(gomock.Any(), "testuser", "password123").
					Return(u, nil)

				mockUserFacade.EXPECT().
					CreateTokens(gomock.Any(), u).
					Return(facade.TokenPair{
						AccessToken:  "valid.jwt.token",
						RefreshToken: facade.RefreshToken{Token: "valid.refresh.token"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: auth.TokenResp{
				AccessToken: "valid.jwt.token",
			},
		},
		{
			name: "user not found",
			request: auth.SignInReq{
				Username: "nonexistent",
				Password: "password123",
			},
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().
					SignIn(gomock.Any(), "nonexistent", "password123").
					Return(model.User{}, facade.ErrSignInInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: web.ErrResp{
				Error: authErrorMsg,
			},
		},
		{
			name: "invalid password",
			request: auth.SignInReq{
				Username: "testuser",
				Password: "wrongpassword",
			},
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().
					SignIn(gomock.Any(), "testuser", "wrongpassword").
					Return(model.User{}, facade.ErrSignInInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: web.ErrResp{
				Error: authErrorMsg,
			},
		},
		{
			name: "user repo error",
			request: auth.SignInReq{
				Username: "testuser",
				Password: "password123",
			},
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				mockUserFacade.EXPECT().
					SignIn(gomock.Any(), "testuser", "password123").
					Return(model.User{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp: web.ErrResp{
				Error: internalErrorMsg,
			},
		},
		{
			name: "token generation error",
			request: auth.SignInReq{
				Username: "testuser",
				Password: "password123",
			},
			setupMocks: func(mockUserFacade *mocks.MockUserFacade) {
				u := model.User{ID: "uid-1", Username: "testuser"}

				mockUserFacade.EXPECT().
					SignIn(gomock.Any(), "testuser", "password123").
					Return(u, nil)

				mockUserFacade.EXPECT().
					CreateTokens(gomock.Any(), u).
					Return(facade.TokenPair{}, errors.New("token generation error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp: web.ErrResp{
				Error: internalErrorMsg,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, authAPI, mockUserFacade, app, ctrl := setupTest(t, nil)
			defer ctrl.Finish()

			if tt.setupMocks != nil {
				tt.setupMocks(mockUserFacade)
			}

			app.Post("/signin", authAPI.SignInHandler)

			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/signin", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			switch v := tt.expectedResp.(type) {
			case auth.TokenResp:
				var actual auth.TokenResp
				err = json.Unmarshal(body, &actual)
				require.NoError(t, err)
				assert.Equal(t, v, actual)
			case web.ErrResp:
				var actual web.ErrResp
				err = json.Unmarshal(body, &actual)
				require.NoError(t, err)
				assert.Equal(t, v.Error, actual.Error)
			}
		})
	}
}
