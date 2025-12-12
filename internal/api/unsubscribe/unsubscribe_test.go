package unsubscribe_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUnsubscribeHandler_Success(t *testing.T) {
	api, mockFacade, ctrl, tokenGen, mockViews := setupTest(t)
	defer ctrl.Finish()

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	token := tokenGen.GenerateToken(email, expiresAt)

	mockViews.EXPECT().Load().Return(nil).AnyTimes()
	mockViews.EXPECT().Render(gomock.Any(), "unsubscribe", gomock.Any()).Return(nil)

	mockFacade.EXPECT().IsEmailUnsubscribed(gomock.Any(), email).Return(false, nil)

	app := fiber.New(fiber.Config{
		Views: mockViews,
	})
	app.Get("/unsubscribe", api.UnsubscribeHandler)

	req := httptest.NewRequest(http.MethodGet, "/unsubscribe?token="+token, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUnsubscribeHandler_MissingToken(t *testing.T) {
	api, _, ctrl, _, _ := setupTest(t)
	defer ctrl.Finish()

	app := fiber.New()
	app.Get("/unsubscribe", api.UnsubscribeHandler)

	req := httptest.NewRequest(http.MethodGet, "/unsubscribe", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Missing token parameter")
}

func TestUnsubscribeHandler_InvalidToken(t *testing.T) {
	api, _, ctrl, _, _ := setupTest(t)
	defer ctrl.Finish()

	app := fiber.New()
	app.Get("/unsubscribe", api.UnsubscribeHandler)

	req := httptest.NewRequest(http.MethodGet, "/unsubscribe?token=invalid-token", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Invalid or expired unsubscribe link")
}

func TestUnsubscribeHandler_ExpiredToken(t *testing.T) {
	api, _, ctrl, tokenGen, _ := setupTest(t)
	defer ctrl.Finish()

	app := fiber.New()
	app.Get("/unsubscribe", api.UnsubscribeHandler)

	email := "test@example.com"
	expiresAt := time.Now().Add(-1 * time.Hour)
	token := tokenGen.GenerateToken(email, expiresAt)

	req := httptest.NewRequest(http.MethodGet, "/unsubscribe?token="+token, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Invalid or expired unsubscribe link")
}

func TestUnsubscribeHandler_AlreadyUnsubscribed(t *testing.T) {
	api, mockFacade, ctrl, tokenGen, mockViews := setupTest(t)
	defer ctrl.Finish()

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	token := tokenGen.GenerateToken(email, expiresAt)

	mockViews.EXPECT().Load().Return(nil).AnyTimes()
	mockViews.EXPECT().Render(gomock.Any(), "unsubscribe_already", gomock.Any()).Return(nil)

	mockFacade.EXPECT().IsEmailUnsubscribed(gomock.Any(), email).Return(true, nil)

	app := fiber.New(fiber.Config{
		Views: mockViews,
	})
	app.Get("/unsubscribe", api.UnsubscribeHandler)

	req := httptest.NewRequest(http.MethodGet, "/unsubscribe?token="+token, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
