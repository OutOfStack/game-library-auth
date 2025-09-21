package database_test

import (
	"context"
	"testing"

	"github.com/OutOfStack/game-library-auth/pkg/database"
	mocks "github.com/OutOfStack/game-library-auth/pkg/database/mocks"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mockResult struct{}

func (m mockResult) LastInsertId() (int64, error) { return 1, nil }
func (m mockResult) RowsAffected() (int64, error) { return 1, nil }

func TestWithTx(t *testing.T) {
	tx := &sqlx.Tx{}

	ctxWithTx := database.WithTx(t.Context(), tx)

	retrievedTx, exists := database.TxFromContext(ctxWithTx)
	assert.True(t, exists)
	assert.Equal(t, tx, retrievedTx)
}

func TestTxFromContext_WithTransaction(t *testing.T) {
	tx := &sqlx.Tx{}
	ctxWithTx := database.WithTx(t.Context(), tx)

	retrievedTx, exists := database.TxFromContext(ctxWithTx)

	assert.True(t, exists)
	assert.Equal(t, tx, retrievedTx)
}

func TestTxFromContext_NoTransaction(t *testing.T) {
	retrievedTx, exists := database.TxFromContext(t.Context())

	assert.False(t, exists)
	assert.Nil(t, retrievedTx)
}

func TestTxFromContext_WrongType(t *testing.T) {
	type ctxKeyType string
	ctx := context.WithValue(t.Context(), ctxKeyType("tx"), "not-a-transaction")

	retrievedTx, exists := database.TxFromContext(ctx)

	assert.False(t, exists)
	assert.Nil(t, retrievedTx)
}

func TestNewQuerier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mocks.NewMockExecutor(ctrl)

	querier := database.NewQuerier(mockDB)

	assert.NotNil(t, querier)
}

func TestEx_Exec(t *testing.T) {
	ctx := t.Context()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mocks.NewMockExecutor(ctrl)

	query := "INSERT INTO users (name) VALUES (?)"
	args := []interface{}{"test"}
	expectedResult := mockResult{}

	mockDB.EXPECT().ExecContext(ctx, query, args[0]).Return(expectedResult, nil)

	querier := database.NewQuerier(mockDB)
	result, err := querier.Exec(ctx, query, args...)

	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestEx_Get(t *testing.T) {
	ctx := t.Context()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mocks.NewMockExecutor(ctrl)

	query := "SELECT name FROM users WHERE id = ?"
	args := []interface{}{1}
	dest := &struct{ Name string }{}

	mockDB.EXPECT().GetContext(ctx, dest, query, args[0]).Return(nil)

	querier := database.NewQuerier(mockDB)
	err := querier.Get(ctx, dest, query, args...)

	assert.NoError(t, err)
}
