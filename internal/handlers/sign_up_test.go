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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSignUpHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        handlers.SignUpReq
		setupMocks     func(*mocks.MockAuth, *mocks.MockStorage)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name: "successful user signup",
			request: handlers.SignUpReq{
				Username:        "newuser",
				Name:            "New User",
				Password:        "password123",
				ConfirmPassword: "password123",
				IsPublisher:     false,
			},
			setupMocks: func(_ *mocks.MockAuth, mockStorage *mocks.MockStorage) {
				mockStorage.EXPECT().
					GetUserByUsername(gomock.Any(), "newuser").
					Return(database.User{}, database.ErrNotFound)

				mockStorage.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, user database.User) error {
						assert.Equal(t, "newuser", user.Username)
						assert.Equal(t, "New User", user.Name)
						assert.Equal(t, database.UserRoleName, user.Role)
						return nil
					})
			},
			expectedStatus: http.StatusOK,
			expectedResp:   handlers.SignUpResp{},
		},
		{
			name: "successful publisher signup",
			request: handlers.SignUpReq{
				Username:        "newpublisher",
				Name:            "Publisher Co",
				Password:        "password123",
				ConfirmPassword: "password123",
				IsPublisher:     true,
			},
			setupMocks: func(_ *mocks.MockAuth, mockStorage *mocks.MockStorage) {
				mockStorage.EXPECT().
					GetUserByUsername(gomock.Any(), "newpublisher").
					Return(database.User{}, database.ErrNotFound)

				mockStorage.EXPECT().
					CheckUserExists(gomock.Any(), "Publisher Co", database.PublisherRoleName).
					Return(false, nil)

				mockStorage.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, user database.User) error {
						assert.Equal(t, "newpublisher", user.Username)
						assert.Equal(t, "Publisher Co", user.Name)
						assert.Equal(t, database.PublisherRoleName, user.Role)
						return nil
					})
			},
			expectedStatus: http.StatusOK,
			expectedResp:   handlers.SignUpResp{},
		},
		{
			name: "username already exists",
			request: handlers.SignUpReq{
				Username:        "existinguser",
				Name:            "Existing User",
				Password:        "password123",
				ConfirmPassword: "password123",
				IsPublisher:     false,
			},
			setupMocks: func(_ *mocks.MockAuth, mockStorage *mocks.MockStorage) {
				mockStorage.EXPECT().
					GetUserByUsername(gomock.Any(), "existinguser").
					Return(database.User{Username: "existinguser"}, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedResp: web.ErrResp{
				Error: "This username is already taken",
			},
		},
		{
			name: "publisher name already exists",
			request: handlers.SignUpReq{
				Username:        "newpublisher",
				Name:            "Existing Publisher",
				Password:        "password123",
				ConfirmPassword: "password123",
				IsPublisher:     true,
			},
			setupMocks: func(_ *mocks.MockAuth, mockStorage *mocks.MockStorage) {
				mockStorage.EXPECT().
					GetUserByUsername(gomock.Any(), "newpublisher").
					Return(database.User{}, database.ErrNotFound)

				mockStorage.EXPECT().
					CheckUserExists(gomock.Any(), "Existing Publisher", database.PublisherRoleName).
					Return(true, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedResp: web.ErrResp{
				Error: "Publisher with this name already exists",
			},
		},
		{
			name: "database error on create",
			request: handlers.SignUpReq{
				Username:        "newuser",
				Name:            "New User",
				Password:        "password123",
				ConfirmPassword: "password123",
				IsPublisher:     false,
			},
			setupMocks: func(_ *mocks.MockAuth, mockStorage *mocks.MockStorage) {
				mockStorage.EXPECT().
					GetUserByUsername(gomock.Any(), "newuser").
					Return(database.User{}, database.ErrNotFound)

				mockStorage.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
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
			mockAuth, mockStorage, _, authAPI, app, ctrl := setupTest(t, nil)
			defer ctrl.Finish()

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockStorage)
			}

			app.Post("/signup", authAPI.SignUpHandler)

			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			switch v := tt.expectedResp.(type) {
			case handlers.SignUpResp:
				var actual handlers.SignUpResp
				err = json.Unmarshal(body, &actual)
				require.NoError(t, err)
				assert.NotEmpty(t, actual.ID) // We don't know the exact ID but it should not be empty
			case web.ErrResp:
				var actual web.ErrResp
				err = json.Unmarshal(body, &actual)
				require.NoError(t, err)
				assert.Equal(t, v.Error, actual.Error)
			}
		})
	}
}
