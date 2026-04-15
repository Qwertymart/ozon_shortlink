package inmemory

import (
	"context"

	"github.com/Qwertymart/ozon_shortlink/internal/entity"
	inmemory "github.com/Qwertymart/ozon_shortlink/pkg/in_memory"
)

type MemoryRepository struct {
	db         *inmemory.Storage
	reverseMap *inmemory.Storage // вторая база для поиска по полному ключу
}

func NewRepository() *MemoryRepository {
	return &MemoryRepository{
		db:         inmemory.NewStorage(), 
		reverseMap: inmemory.NewStorage(),
	}
}

func (r *MemoryRepository) GetNextID(ctx context.Context) (uint64, error) {
	return r.db.IncrAndGet(), nil
}

func (r *MemoryRepository) Save(ctx context.Context, url entity.URL) error {
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

func (r *MemoryRepository) GetByFull(ctx context.Context, full string) (entity.URL, error) {
	short, exists := r.reverseMap.Get(full)
	if !exists {
		return entity.URL{}, entity.ErrNotFound
	}
	return entity.URL{Full: full, Short: short}, nil
}
