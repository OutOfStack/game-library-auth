package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestSignInHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        handlers.SignInReq
		setupMocks     func(*mocks.MockAuth, *mocks.MockUserRepo)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name: "successful sign in",
			request: handlers.SignInReq{
				Username: "testuser",
				Password: "password123",
			},
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo) {
				passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := database.NewUser("testuser", "", passwordHash, database.UserRoleName)

				mockUserRepo.EXPECT().
					GetUserByUsername(gomock.Any(), "testuser").
					Return(user, nil)

				mockAuth.EXPECT().
					CreateClaims(gomock.Eq(user)).
					Return(jwt.MapClaims{"sub": user.ID})

				mockAuth.EXPECT().
					GenerateToken(gomock.Any()).
					Return("valid.jwt.token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: handlers.TokenResp{
				AccessToken: "valid.jwt.token",
			},
		},
		{
			name: "user not found",
			request: handlers.SignInReq{
				Username: "nonexistent",
				Password: "password123",
			},
			setupMocks: func(_ *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo) {
				mockUserRepo.EXPECT().
					GetUserByUsername(gomock.Any(), "nonexistent").
					Return(database.User{}, database.ErrNotFound)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: web.ErrResp{
				Error: authErrorMsg,
			},
		},
		{
			name: "invalid password",
			request: handlers.SignInReq{
				Username: "testuser",
				Password: "wrongpassword",
			},
			setupMocks: func(_ *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo) {
				passwordHash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
				user := database.NewUser("testuser", "", passwordHash, database.UserRoleName)

				mockUserRepo.EXPECT().
					GetUserByUsername(gomock.Any(), "testuser").
					Return(user, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: web.ErrResp{
				Error: authErrorMsg,
			},
		},
		{
			name: "user repo error",
			request: handlers.SignInReq{
				Username: "testuser",
				Password: "password123",
			},
			setupMocks: func(_ *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo) {
				mockUserRepo.EXPECT().
					GetUserByUsername(gomock.Any(), "testuser").
					Return(database.User{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp: web.ErrResp{
				Error: internalErrorMsg,
			},
		},
		{
			name: "token generation error",
			request: handlers.SignInReq{
				Username: "testuser",
				Password: "password123",
			},
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo) {
				passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := database.NewUser("testuser", "", passwordHash, database.UserRoleName)

				mockUserRepo.EXPECT().
					GetUserByUsername(gomock.Any(), "testuser").
					Return(user, nil)

				mockAuth.EXPECT().
					CreateClaims(gomock.Eq(user)).
					Return(jwt.MapClaims{"sub": user.ID})

				mockAuth.EXPECT().
					GenerateToken(gomock.Any()).
					Return("", errors.New("token generation error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp: web.ErrResp{
				Error: internalErrorMsg,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth, mockUserRepo, _, _, authAPI, app, ctrl := setupTest(t, nil)
			defer ctrl.Finish()

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockUserRepo)
			}

			app.Post("/signin", authAPI.SignInHandler)

			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
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
