package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Qwertymart/ozon_shortlink/internal/entity"
	transport "github.com/Qwertymart/ozon_shortlink/internal/transport/http"
)

type MockShortenerService struct {
	CreateFunc func(ctx context.Context, url string) (string, error)
	GetFunc    func(ctx context.Context, shortURL string) (string, error)
}

func (m *MockShortenerService) Create(ctx context.Context, url string) (string, error) {
	return m.CreateFunc(ctx, url)
}

func (m *MockShortenerService) Get(ctx context.Context, shortURL string) (string, error) {
	return m.GetFunc(ctx, shortURL)
}

func TestHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := &MockShortenerService{
			CreateFunc: func(ctx context.Context, url string) (string, error) {
				return "baaaaaaaaa", nil
			},
		}
		handler := transport.NewHandler(mockService)
		router := handler.InitRoutes()

		body := `{"url": "https://google.com"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rr.Code)
		}

		var resp map[string]string
		json.Unmarshal(rr.Body.Bytes(), &resp)
		if resp["result"] != "baaaaaaaaa" {
			t.Errorf("expected baaaaaaaaa, got %s", resp["result"])
		}
	})

	t.Run("bad_request_json", func(t *testing.T) {
		handler := transport.NewHandler(&MockShortenerService{})
		router := handler.InitRoutes()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{invalid json`))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("invalid_url_format", func(t *testing.T) {
		mockService := &MockShortenerService{
			CreateFunc: func(ctx context.Context, url string) (string, error) {
				return "", entity.ErrInvalidFormat
			},
		}
		handler := transport.NewHandler(mockService)
		router := handler.InitRoutes()

		body := `{"url": "invalid"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rr.Code)
		}
	})
}

func TestHandler_Get(t *testing.T) {
	t.Run("success_redirect", func(t *testing.T) {
		targetURL := "https://google.com"
		mockService := &MockShortenerService{
			GetFunc: func(ctx context.Context, shortURL string) (string, error) {
				return targetURL, nil
			},
		}
		handler := transport.NewHandler(mockService)
		router := handler.InitRoutes()

		req := httptest.NewRequest(http.MethodGet, "/baaaaaaaaa", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusTemporaryRedirect {
			t.Errorf("expected 307, got %d", rr.Code)
		}

		if loc := rr.Header().Get("Location"); loc != targetURL {
			t.Errorf("expected location %s, got %s", targetURL, loc)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		mockService := &MockShortenerService{
			GetFunc: func(ctx context.Context, shortURL string) (string, error) {
				return "", entity.ErrNotFound
			},
		}
		handler := transport.NewHandler(mockService)
		router := handler.InitRoutes()

		req := httptest.NewRequest(http.MethodGet, "/baaaaaaaaa", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("invalid_id", func(t *testing.T) {
		mockService := &MockShortenerService{
			GetFunc: func(ctx context.Context, shortURL string) (string, error) {
				return "", entity.ErrInvalidFormat
			},
		}
		handler := transport.NewHandler(mockService)
		router := handler.InitRoutes()

		req := httptest.NewRequest(http.MethodGet, "/short", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("internal_error", func(t *testing.T) {
		mockService := &MockShortenerService{
			GetFunc: func(ctx context.Context, shortURL string) (string, error) {
				return "", errors.New("db error")
			},
		}
		handler := transport.NewHandler(mockService)
		router := handler.InitRoutes()

		req := httptest.NewRequest(http.MethodGet, "/baaaaaaaaa", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", rr.Code)
		}
	})
}