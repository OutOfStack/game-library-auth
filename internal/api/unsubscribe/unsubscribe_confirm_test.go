package unsubscribe_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUnsubscribeConfirmHandler_Success(t *testing.T) {
	api, mockFacade, ctrl, tokenGen, mockViews := setupTest(t)
	defer ctrl.Finish()

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	token := tokenGen.GenerateToken(email, expiresAt)

	mockViews.EXPECT().Load().Return(nil).AnyTimes()
	mockViews.EXPECT().Render(gomock.Any(), "unsubscribe_success", gomock.Any()).Return(nil)

	mockFacade.EXPECT().UnsubscribeEmail(gomock.Any(), token).Return(email, nil)

	app := fiber.New(fiber.Config{Views: mockViews})
	app.Post("/unsubscribe", api.UnsubscribeConfirmHandler)

	form := url.Values{}
	form.Add("token", token)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/unsubscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUnsubscribeConfirmHandler_MissingToken(t *testing.T) {
	api, _, ctrl, _, _ := setupTest(t)
	defer ctrl.Finish()

	app := fiber.New()
	app.Post("/unsubscribe", api.UnsubscribeConfirmHandler)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/unsubscribe", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Missing token")
}

func TestUnsubscribeConfirmHandler_NotFound(t *testing.T) {
	api, mockFacade, ctrl, tokenGen, _ := setupTest(t)
	defer ctrl.Finish()

	app := fiber.New()
	app.Post("/unsubscribe", api.UnsubscribeConfirmHandler)

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	token := tokenGen.GenerateToken(email, expiresAt)

	mockFacade.EXPECT().UnsubscribeEmail(gomock.Any(), token).Return("", database.ErrNotFound)

	form := url.Values{}
	form.Add("token", token)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/unsubscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Unsubscribe link not found or expired")
}

func TestUnsubscribeConfirmHandler_DatabaseError(t *testing.T) {
	api, mockFacade, ctrl, tokenGen, _ := setupTest(t)
	defer ctrl.Finish()

	app := fiber.New()
	app.Post("/unsubscribe", api.UnsubscribeConfirmHandler)

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	token := tokenGen.GenerateToken(email, expiresAt)

	mockFacade.EXPECT().UnsubscribeEmail(gomock.Any(), token).Return("", errors.New("database error"))

	form := url.Values{}
	form.Add("token", token)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/unsubscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Failed to process unsubscribe request")
}
