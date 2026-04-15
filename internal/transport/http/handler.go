package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Qwertymart/ozon_shortlink/internal/entity"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type ShortenerService interface {
	Create(ctx context.Context, originalURL string) (string, error)
	Get(ctx context.Context, shortURL string) (string, error)
}

type Handler struct {
	service ShortenerService
}

func NewHandler(service ShortenerService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) InitRoutes() *chi.Mux {
	r := chi.NewRouter()

	// подключаем стандартные middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/", h.createShortLink)
	r.Get("/{id}", h.redirectToOriginal)

	return r
}

type createRequest struct {
	URL string `json:"url"`
}

type createResponse struct {
	Result string `json:"result"`
}

func (h *Handler) createShortLink(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	shortURL, err := h.service.Create(r.Context(), req.URL)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidFormat) {
			http.Error(w, "invalid url format", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := createResponse{
		Result: shortURL,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) redirectToOriginal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	originalURL, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, entity.ErrInvalidFormat) {
			http.Error(w, "invalid id format", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// выполняем редирект
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}