package facade_test

import (
	"encoding/base64"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/facade"
	mocks "github.com/OutOfStack/game-library-auth/internal/facade/mocks"
	"github.com/OutOfStack/game-library-auth/internal/notification"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*facade.Provider, *mocks.MockUserRepo, *mocks.MockEmailPreparer, *mocks.MockAuth, *mocks.MockInfoAPIClient, *gomock.Controller) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepo(ctrl)
	mockEmailSender := mocks.NewMockEmailPreparer(ctrl)
	mockAuth := mocks.NewMockAuth(ctrl)
	mockInfoAPIClient := mocks.NewMockInfoAPIClient(ctrl)
	unsubscribeTokenGenerator := auth.NewUnsubscribeTokenGenerator([]byte("test-secret-key"))

	cipher, err := notification.NewCipher(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	require.NoError(t, err)
	provider := facade.New(zap.NewNop(), mockUserRepo, mockEmailSender, mockAuth, unsubscribeTokenGenerator, mockInfoAPIClient, cipher)

	return provider, mockUserRepo, mockEmailSender, mockAuth, mockInfoAPIClient, ctrl
}
