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
	newName := "Updated Name"
	newAvatarURL := "https://example.com/avatar.jpg"

	tests := []struct {
		name           string
		request        handlers.UpdateProfileReq
		setupMocks     func(*mocks.MockAuth, *mocks.MockStorage)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name: "successful update name and avatar",
			request: handlers.UpdateProfileReq{
				UserID:    userID,
				Name:      &newName,
				AvatarURL: &newAvatarURL,
			},
			setupMocks: func(mockAuth *mocks.MockAuth, mockStorage *mocks.MockStorage) {
				user := database.User{
					ID:       uuid.MustParse(userID),
					Username: "testuser",
					Name:     "Old Name",
				}

				mockStorage.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(user, nil)

				mockStorage.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, updatedUser database.User) error {
						assert.Equal(t, newName, updatedUser.Name)
						assert.Equal(t, newAvatarURL, updatedUser.AvatarURL.String)
						assert.True(t, updatedUser.AvatarURL.Valid)
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
			name: "successful update password",
			request: handlers.UpdateProfileReq{
				UserID:             userID,
				Password:           &oldPassword,
				NewPassword:        &newPassword,
				ConfirmNewPassword: &newPassword,
			},
			setupMocks: func(mockAuth *mocks.MockAuth, mockStorage *mocks.MockStorage) {
				passwordHash, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.MinCost)
				user := database.User{
					ID:           uuid.MustParse(userID),
					Username:     "testuser",
					PasswordHash: passwordHash,
				}

				mockStorage.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(user, nil)

				mockStorage.EXPECT().
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
			name: "user not found",
			request: handlers.UpdateProfileReq{
				UserID: userID,
				Name:   &newName,
			},
			setupMocks: func(_ *mocks.MockAuth, mockStorage *mocks.MockStorage) {
				mockStorage.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(database.User{}, database.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedResp: web.ErrResp{
				Error: "User does not exist",
			},
		},
		{
			name: "invalid current password",
			request: handlers.UpdateProfileReq{
				UserID:             userID,
				Password:           &oldPassword,
				NewPassword:        &newPassword,
				ConfirmNewPassword: &newPassword,
			},
			setupMocks: func(_ *mocks.MockAuth, mockStorage *mocks.MockStorage) {
				passwordHash, _ := bcrypt.GenerateFromPassword([]byte("differentpassword"), bcrypt.MinCost)
				user := database.User{
					ID:           uuid.MustParse(userID),
					Username:     "testuser",
					PasswordHash: passwordHash,
				}

				mockStorage.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(user, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: web.ErrResp{
				Error: "Invalid current password",
			},
		},
		{
			name: "storage error on update",
			request: handlers.UpdateProfileReq{
				UserID: userID,
				Name:   &newName,
			},
			setupMocks: func(_ *mocks.MockAuth, mockStorage *mocks.MockStorage) {
				user := database.User{
					ID:       uuid.MustParse(userID),
					Username: "testuser",
				}

				mockStorage.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(user, nil)

				mockStorage.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
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
			mockAuth, mockStorage, authAPI, app, ctrl := setupTest(t)
			defer ctrl.Finish()

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockStorage)
			}

			app.Post("/update_profile", authAPI.UpdateProfileHandler)

			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/update_profile", bytes.NewReader(reqBody))
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
