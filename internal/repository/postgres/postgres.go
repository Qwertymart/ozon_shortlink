package postgres

import (
	"context"
	"errors"

	"github.com/Qwertymart/ozon_shortlink/internal/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) GetNextID(ctx context.Context) (uint64, error) {
	var id uint64

	query := `SELECT nextval('urls_id_seq')`

	err := r.pool.QueryRow(ctx, query).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) Save(ctx context.Context, url entity.URL) error {
	query := `
		INSERT INTO urls (short_url, original_url) 
		VALUES ($1, $2)
	`

	_, err := r.pool.Exec(ctx, query, url.Short, url.Full)
	if err != nil {
		// 23505 - проблемы с уникальностью
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.ErrConflict
		}
		return err
	}

	return nil
}

func (r *Repository) GetByShort(ctx context.Context, short string) (entity.URL, error) {
	query := `
		SELECT original_url 
		FROM urls 
		WHERE short_url = $1
	`

	var fullURL string
	err := r.pool.QueryRow(ctx, query, short).Scan(&fullURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.URL{}, entity.ErrNotFound
		}
		return entity.URL{}, err
	}

	return entity.URL{
		Short: short,
		Full:  fullURL,
	}, nil
}

func (r *Repository) GetByFull(ctx context.Context, full string) (entity.URL, error) {
	query := `
		SELECT short_url 
		FROM urls 
		WHERE original_url = $1
	`

	var shortURL string
	err := r.pool.QueryRow(ctx, query, full).Scan(&shortURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.URL{}, entity.ErrNotFound
		}
		return entity.URL{}, err
	}

	return entity.URL{
		Short: shortURL,
		Full:  full,
	}, nil
}
