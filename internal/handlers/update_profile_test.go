package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	auth_ "github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
		setupMocks     func(*mocks.MockAuth, *mocks.MockUserFacade)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name:       "successful update name",
			authHeader: "Bearer valid-token",
			request: handlers.UpdateProfileReq{
				Name: &newName,
			},
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserFacade *mocks.MockUserFacade) {
				// Mock JWT validation
				claims := auth_.Claims{UserID: userID}
				mockAuth.EXPECT().
					ValidateToken("valid-token").
					Return(claims, nil).
					AnyTimes()

				updated := model.User{ID: userID, Username: "testuser", DisplayName: newName}
				mockUserFacade.EXPECT().
					UpdateUserProfile(gomock.Any(), userID, gomock.Any()).
					Return(updated, nil)

				mockAuth.EXPECT().
					CreateUserClaims(updated).
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
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserFacade *mocks.MockUserFacade) {
				// Mock JWT validation
				claims := auth_.Claims{UserID: userID}
				mockAuth.EXPECT().
					ValidateToken("valid-token").
					Return(claims, nil).
					AnyTimes()

				updated := model.User{ID: userID, Username: "testuser"}
				mockUserFacade.EXPECT().
					UpdateUserProfile(gomock.Any(), userID, gomock.Any()).
					Return(updated, nil)

				mockAuth.EXPECT().
					CreateUserClaims(updated).
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
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserFacade *mocks.MockUserFacade) {
				// Mock JWT validation
				claims := auth_.Claims{UserID: userID}
				mockAuth.EXPECT().
					ValidateToken("valid-token").
					Return(claims, nil).
					AnyTimes()

				mockUserFacade.EXPECT().
					UpdateUserProfile(gomock.Any(), userID, gomock.Any()).
					Return(model.User{}, facade.UpdateProfileUserNotFoundErr)
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
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserFacade *mocks.MockUserFacade) {
				// Mock JWT validation
				claims := auth_.Claims{UserID: userID}
				mockAuth.EXPECT().
					ValidateToken("valid-token").
					Return(claims, nil).
					AnyTimes()
				mockUserFacade.EXPECT().
					UpdateUserProfile(gomock.Any(), userID, gomock.Any()).
					Return(model.User{}, facade.UpdateProfileInvalidPasswordErr)
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
			setupMocks: func(mockAuth *mocks.MockAuth, mockUserFacade *mocks.MockUserFacade) {
				// Mock JWT validation
				claims := auth_.Claims{UserID: userID}
				mockAuth.EXPECT().
					ValidateToken("valid-token").
					Return(claims, nil).
					AnyTimes()
				mockUserFacade.EXPECT().
					UpdateUserProfile(gomock.Any(), userID, gomock.Any()).
					Return(model.User{}, errors.New("database error"))
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
			setupMocks:     func(_ *mocks.MockAuth, _ *mocks.MockUserFacade) {},
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
			setupMocks:     func(_ *mocks.MockAuth, _ *mocks.MockUserFacade) {},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: web.ErrResp{
				Error: "Invalid or missing authorization token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth, _, authAPI, mockUserFacade, app, ctrl := setupTest(t, nil)
			defer ctrl.Finish()

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockUserFacade)
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
