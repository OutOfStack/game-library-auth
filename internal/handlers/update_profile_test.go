package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	auth_ "github.com/OutOfStack/game-library-auth/internal/auth"
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

func TestUpdateProfileHandler(t *testing.T) {
	userID := uuid.New().String()
	oldPassword := "oldpassword"
	newPassword := "newpassword"
	newName := "Updated DisplayName"

	tests := []struct {
		name           string
		authHeader     string
		request        handlers.UpdateProfileReq
		setupMocks     func(*mocks.MockAuth, *mocks.MockUserRepo)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name:       "successful update name",
			authHeader: "Bearer valid-token",
			request: handlers.UpdateProfileReq{
				Name: &newName,
			},
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo) {
				// Mock JWT validation
				claims := auth_.Claims{UserID: userID}
				mockAuth.EXPECT().
					ValidateToken("valid-token").
					Return(claims, nil).
					AnyTimes()

				user := database.User{
					ID:          userID,
					Username:    "testuser",
					DisplayName: "Old DisplayName",
				}

				mockUserRepo.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(user, nil)

				mockUserRepo.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, updatedUser database.User) error {
						assert.Equal(t, newName, updatedUser.DisplayName)
						return nil
					})

				mockAuth.EXPECT().
					CreateClaims(gomock.Any()).
					Return(jwt.MapClaims{"sub": userID})

				mockAuth.EXPECT().
					GenerateToken(gomock.Any()).
					Return("updated.jwt.token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: handlers.TokenResp{
				AccessToken: "updated.jwt.token",
			},
		},
		{
			name:       "successful update password",
			authHeader: "Bearer valid-token",
			request: handlers.UpdateProfileReq{
				Password:           &oldPassword,
				NewPassword:        &newPassword,
				ConfirmNewPassword: &newPassword,
			},
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo) {
				// Mock JWT validation
				claims := auth_.Claims{UserID: userID}
				mockAuth.EXPECT().
					ValidateToken("valid-token").
					Return(claims, nil).
					AnyTimes()

				passwordHash, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.MinCost)
				user := database.User{
					ID:           userID,
					Username:     "testuser",
					PasswordHash: passwordHash,
				}

				mockUserRepo.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(user, nil)

				mockUserRepo.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Return(nil)

				mockAuth.EXPECT().
					CreateClaims(gomock.Any()).
					Return(jwt.MapClaims{"sub": userID})

				mockAuth.EXPECT().
					GenerateToken(gomock.Any()).
					Return("updated.jwt.token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: handlers.TokenResp{
				AccessToken: "updated.jwt.token",
			},
		},
		{
			name:       "user not found",
			authHeader: "Bearer valid-token",
			request: handlers.UpdateProfileReq{
				Name: &newName,
			},
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo) {
				// Mock JWT validation
				claims := auth_.Claims{UserID: userID}
				mockAuth.EXPECT().
					ValidateToken("valid-token").
					Return(claims, nil).
					AnyTimes()

				mockUserRepo.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(database.User{}, database.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedResp: web.ErrResp{
				Error: "User does not exist",
			},
		},
		{
			name:       "invalid current password",
			authHeader: "Bearer valid-token",
			request: handlers.UpdateProfileReq{
				Password:           &oldPassword,
				NewPassword:        &newPassword,
				ConfirmNewPassword: &newPassword,
			},
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo) {
				// Mock JWT validation
				claims := auth_.Claims{UserID: userID}
				mockAuth.EXPECT().
					ValidateToken("valid-token").
					Return(claims, nil).
					AnyTimes()

				passwordHash, _ := bcrypt.GenerateFromPassword([]byte("differentpassword"), bcrypt.MinCost)
				user := database.User{
					ID:           userID,
					Username:     "testuser",
					PasswordHash: passwordHash,
				}

				mockUserRepo.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(user, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: web.ErrResp{
				Error: "Invalid current password",
			},
		},
		{
			name:       "user repo error on update",
			authHeader: "Bearer valid-token",
			request: handlers.UpdateProfileReq{
				Name: &newName,
			},
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserRepo *mocks.MockUserRepo) {
				// Mock JWT validation
				claims := auth_.Claims{UserID: userID}
				mockAuth.EXPECT().
					ValidateToken("valid-token").
					Return(claims, nil).
					AnyTimes()

				user := database.User{
					ID:       userID,
					Username: "testuser",
				}

				mockUserRepo.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(user, nil)

				mockUserRepo.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp: web.ErrResp{
				Error: internalErrorMsg,
			},
		},
		{
			name:       "missing authorization header",
			authHeader: "",
			request: handlers.UpdateProfileReq{
				Name: &newName,
			},
			setupMocks:     func(_ *mocks.MockAuth, _ *mocks.MockUserRepo) {},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: web.ErrResp{
				Error: "Invalid or missing authorization token",
			},
		},
		{
			name:       "invalid authorization header format",
			authHeader: "InvalidFormat",
			request: handlers.UpdateProfileReq{
				Name: &newName,
			},
			setupMocks:     func(_ *mocks.MockAuth, _ *mocks.MockUserRepo) {},
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

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockUserRepo)
			}

			app.Post("/update_profile", authAPI.UpdateProfileHandler)

			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/update_profile", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

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
