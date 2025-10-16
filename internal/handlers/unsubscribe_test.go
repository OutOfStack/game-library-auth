package handlers_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/handlers"
	handlers_mocks "github.com/OutOfStack/game-library-auth/internal/handlers/mocks"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func setupUnsubscribeTest(t *testing.T) (*handlers.UnsubscribeAPI, *handlers_mocks.MockUnsubscribeFacade, *gomock.Controller, *auth.UnsubscribeTokenGenerator) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockFacade := handlers_mocks.NewMockUnsubscribeFacade(ctrl)
	log := zap.NewNop()
	tokenGen := auth.NewUnsubscribeTokenGenerator([]byte("test-secret-key"))
	api := handlers.NewUnsubscribeAPI(log, tokenGen, mockFacade, "test@example.com")
	return api, mockFacade, ctrl, tokenGen
}

func TestUnsubscribeHandler_Success(t *testing.T) {
	api, mockFacade, ctrl, tokenGen := setupUnsubscribeTest(t)
	defer ctrl.Finish()

	app := fiber.New(fiber.Config{
		Views: &mockTemplateEngine{},
	})
	app.Get("/unsubscribe", api.UnsubscribeHandler)

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	token := tokenGen.GenerateToken(email, expiresAt)

	mockFacade.EXPECT().
		IsEmailUnsubscribed(gomock.Any(), email).
		Return(false, nil)

	req := httptest.NewRequest(http.MethodGet, "/unsubscribe?token="+token, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestUnsubscribeHandler_MissingToken(t *testing.T) {
	api, _, ctrl, _ := setupUnsubscribeTest(t)
	defer ctrl.Finish()

	app := fiber.New()
	app.Get("/unsubscribe", api.UnsubscribeHandler)

	req := httptest.NewRequest(http.MethodGet, "/unsubscribe", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Missing token parameter") {
		t.Errorf("expected 'Missing token parameter' message, got %s", string(body))
	}
}

func TestUnsubscribeHandler_InvalidToken(t *testing.T) {
	api, _, ctrl, _ := setupUnsubscribeTest(t)
	defer ctrl.Finish()

	app := fiber.New()
	app.Get("/unsubscribe", api.UnsubscribeHandler)

	req := httptest.NewRequest(http.MethodGet, "/unsubscribe?token=invalid-token", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Invalid or expired unsubscribe link") {
		t.Errorf("expected 'Invalid or expired unsubscribe link' message, got %s", string(body))
	}
}

func TestUnsubscribeHandler_ExpiredToken(t *testing.T) {
	api, _, ctrl, tokenGen := setupUnsubscribeTest(t)
	defer ctrl.Finish()

	app := fiber.New()
	app.Get("/unsubscribe", api.UnsubscribeHandler)

	email := "test@example.com"
	expiresAt := time.Now().Add(-1 * time.Hour)
	token := tokenGen.GenerateToken(email, expiresAt)

	req := httptest.NewRequest(http.MethodGet, "/unsubscribe?token="+token, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Invalid or expired unsubscribe link") {
		t.Errorf("expected 'Invalid or expired unsubscribe link' message, got %s", string(body))
	}
}

func TestUnsubscribeHandler_AlreadyUnsubscribed(t *testing.T) {
	api, mockFacade, ctrl, tokenGen := setupUnsubscribeTest(t)
	defer ctrl.Finish()

	app := fiber.New(fiber.Config{
		Views: &mockTemplateEngine{},
	})
	app.Get("/unsubscribe", api.UnsubscribeHandler)

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	token := tokenGen.GenerateToken(email, expiresAt)

	mockFacade.EXPECT().
		IsEmailUnsubscribed(gomock.Any(), email).
		Return(true, nil)

	req := httptest.NewRequest(http.MethodGet, "/unsubscribe?token="+token, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestUnsubscribeConfirmHandler_Success(t *testing.T) {
	api, mockFacade, ctrl, tokenGen := setupUnsubscribeTest(t)
	defer ctrl.Finish()

	app := fiber.New(fiber.Config{
		Views: &mockTemplateEngine{},
	})
	app.Post("/unsubscribe", api.UnsubscribeConfirmHandler)

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	token := tokenGen.GenerateToken(email, expiresAt)

	mockFacade.EXPECT().
		UnsubscribeEmail(gomock.Any(), token).
		Return(email, nil)

	form := url.Values{}
	form.Add("token", token)

	req := httptest.NewRequest(http.MethodPost, "/unsubscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestUnsubscribeConfirmHandler_MissingToken(t *testing.T) {
	api, _, ctrl, _ := setupUnsubscribeTest(t)
	defer ctrl.Finish()

	app := fiber.New()
	app.Post("/unsubscribe", api.UnsubscribeConfirmHandler)

	req := httptest.NewRequest(http.MethodPost, "/unsubscribe", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Missing token") {
		t.Errorf("expected 'Missing token' message, got %s", string(body))
	}
}

func TestUnsubscribeConfirmHandler_NotFound(t *testing.T) {
	api, mockFacade, ctrl, tokenGen := setupUnsubscribeTest(t)
	defer ctrl.Finish()

	app := fiber.New()
	app.Post("/unsubscribe", api.UnsubscribeConfirmHandler)

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	token := tokenGen.GenerateToken(email, expiresAt)

	mockFacade.EXPECT().
		UnsubscribeEmail(gomock.Any(), token).
		Return("", database.ErrNotFound)

	form := url.Values{}
	form.Add("token", token)

	req := httptest.NewRequest(http.MethodPost, "/unsubscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Unsubscribe link not found or expired") {
		t.Errorf("expected 'Unsubscribe link not found or expired' message, got %s", string(body))
	}
}

func TestUnsubscribeConfirmHandler_DatabaseError(t *testing.T) {
	api, mockFacade, ctrl, tokenGen := setupUnsubscribeTest(t)
	defer ctrl.Finish()

	app := fiber.New()
	app.Post("/unsubscribe", api.UnsubscribeConfirmHandler)

	email := "test@example.com"
	expiresAt := time.Now().Add(24 * time.Hour)
	token := tokenGen.GenerateToken(email, expiresAt)

	dbError := errors.New("database error")
	mockFacade.EXPECT().
		UnsubscribeEmail(gomock.Any(), token).
		Return("", dbError)

	form := url.Values{}
	form.Add("token", token)

	req := httptest.NewRequest(http.MethodPost, "/unsubscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Failed to process unsubscribe request") {
		t.Errorf("expected 'Failed to process unsubscribe request' message, got %s", string(body))
	}
}

type mockTemplateEngine struct{}

func (m *mockTemplateEngine) Load() error {
	return nil
}

func (m *mockTemplateEngine) Render(_ io.Writer, _ string, _ interface{}, _ ...string) error {
	return nil
}
