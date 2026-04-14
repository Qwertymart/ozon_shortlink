package usecase

import (
	"context"

	"github.com/Qwertymart/ozon_shortlink/internal/entity"
)


type URLRepository interface{
	Save(ctx context.Context, url entity.URL) error

	GetByShort(ctx context.Context, short string) (entity.URL, error)

	// GetByFull(ctx context.Context, full string) (entity.URL, error)
}