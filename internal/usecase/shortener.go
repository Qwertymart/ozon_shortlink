package usecase

import (
	"context"
	"errors"
	"net/url"

	"github.com/Qwertymart/ozon_shortlink/internal/entity"
	"github.com/Qwertymart/ozon_shortlink/pkg/base63"
)

type ShortenerUseCase struct {
	repo URLRepository
}

func NewShortener(repo URLRepository) *ShortenerUseCase {
	return &ShortenerUseCase{
		repo: repo,
	}
}

func (u *ShortenerUseCase) Create(ctx context.Context, originalURL string) (string, error) {
	if err := validateURL(originalURL); err != nil {
		return "", err
	}

	// идемпотентность
	existingURL, err := u.repo.GetByFull(ctx, originalURL)
	if err == nil {
		return existingURL.Short, nil
	}

	if !errors.Is(err, entity.ErrNotFound) {
		return "", err
	}

	id, err := u.repo.GetNextID(ctx)
	if err != nil {
		return "", err
	}

	shortURL, err := base63.Encode(id)
	if err != nil {
		return "", err
	}

	urlEntity := entity.URL{
		Short: shortURL,
		Full:  originalURL,
	}

	err = u.repo.Save(ctx, urlEntity)
	if err != nil {
		// обработка Race Condition
		if errors.Is(err, entity.ErrConflict) {
			existingURL, getErr := u.repo.GetByFull(ctx, originalURL)
			if getErr != nil {
				return "", getErr
			}
			return existingURL.Short, nil
		}
		return "", err
	}

	return shortURL, nil
}

func (u *ShortenerUseCase) Get(ctx context.Context, shortURL string) (string, error) {
	if len(shortURL) != 10 {
		return "", entity.ErrInvalidFormat
	}

	urlEntity, err := u.repo.GetByShort(ctx, shortURL)
	if err != nil {
		return "", err
	}

	return urlEntity.Full, nil
}

func validateURL(rawURL string) error {
	if rawURL == "" {
		return entity.ErrInvalidFormat
	}
	
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return entity.ErrInvalidFormat
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return entity.ErrInvalidFormat
	}

	return nil
}