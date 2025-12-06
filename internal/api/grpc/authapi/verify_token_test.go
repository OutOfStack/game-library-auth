package authapi_test

import (
	"context"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/api/grpc/authapi"
	mocks "github.com/OutOfStack/game-library-auth/internal/api/grpc/authapi/mocks"
	pb "github.com/OutOfStack/game-library-auth/pkg/proto/authapi/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthService_VerifyToken(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		setupMocks    func(*mocks.MockAuthFacade)
		expectedValid bool
		expectedError error
	}{
		{
			name:  "valid token",
			token: "valid.jwt.token",
			setupMocks: func(mockAuthFacade *mocks.MockAuthFacade) {
				mockAuthFacade.EXPECT().
					ValidateAccessToken("valid.jwt.token").
					Return(true)
			},
			expectedValid: true,
			expectedError: nil,
		},
		{
			name:  "invalid token",
			token: "invalid.jwt.token",
			setupMocks: func(mockAuthFacade *mocks.MockAuthFacade) {
				mockAuthFacade.EXPECT().
					ValidateAccessToken("invalid.jwt.token").
					Return(false)
			},
			expectedValid: false,
			expectedError: nil,
		},
		{
			name:          "empty token",
			token:         "",
			setupMocks:    func(_ *mocks.MockAuthFacade) {},
			expectedValid: false,
			expectedError: status.Error(codes.InvalidArgument, "empty token"),
		},
		{
			name:          "token with only whitespace",
			token:         "   ",
			setupMocks:    func(_ *mocks.MockAuthFacade) {},
			expectedValid: false,
			expectedError: status.Error(codes.InvalidArgument, "empty token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthFacade := mocks.NewMockAuthFacade(ctrl)
			tt.setupMocks(mockAuthFacade)

			service := authapi.NewAuthService(zap.NewNop(), mockAuthFacade)

			req := &pb.VerifyTokenRequest{}
			req.SetToken(tt.token)

			resp, err := service.VerifyToken(t.Context(), req)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.expectedValid, resp.GetValid())
			}
		})
	}
}

func TestAuthService_VerifyToken_ContextCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthFacade := mocks.NewMockAuthFacade(ctrl)
	service := authapi.NewAuthService(zap.NewNop(), mockAuthFacade)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockAuthFacade.EXPECT().
		ValidateAccessToken("valid.jwt.token").
		Return(true)

	req := &pb.VerifyTokenRequest{}
	req.SetToken("valid.jwt.token")

	resp, err := service.VerifyToken(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.GetValid())
}
