package repository

import (
	"context"

	"github.com/Qwertymart/ozon_shortlink/internal/entity"
	inmemory "github.com/Qwertymart/ozon_shortlink/pkg/in_memory"
)

type MemoryRepository struct {
	db         *inmemory.Storage
	reverseMap *inmemory.Storage // вторая база для поиска по полному ключу 
}

func NewMemoryRepository(db *inmemory.Storage, reverse *inmemory.Storage) *MemoryRepository {
	return &MemoryRepository{
		db:         db,
		reverseMap: reverse,
	}
}

func (r *MemoryRepository) GetNextID(ctx context.Context) (uint64, error) {
	return r.db.IncrAndGet(), nil
}

func (r *MemoryRepository) Save(ctx context.Context, url entity.URL) error {
	// проверка на уникальность
	if _, exists := r.reverseMap.Get(url.Full); exists {
		return entity.ErrConflict
	}

	r.db.Set(url.Short, url.Full)
	r.reverseMap.Set(url.Full, url.Short)
	return nil
}

func (r *MemoryRepository) GetByShort(ctx context.Context, short string) (entity.URL, error) {
	full, exists := r.db.Get(short)
	if !exists {
		return entity.URL{}, entity.ErrNotFound
	}
	return entity.URL{Full: full, Short: short}, nil
}
