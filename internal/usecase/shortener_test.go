package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Qwertymart/ozon_shortlink/internal/entity"
)

type MockRepository struct {
	SaveFunc       func(ctx context.Context, url entity.URL) error
	GetByShortFunc func(ctx context.Context, short string) (entity.URL, error)
	GetNextIDFunc  func(ctx context.Context) (uint64, error)
	GetByFullFunc  func(ctx context.Context, full string) (entity.URL, error)
}

func (m *MockRepository) Save(ctx context.Context, url entity.URL) error { return m.SaveFunc(ctx, url) }
func (m *MockRepository) GetByShort(ctx context.Context, short string) (entity.URL, error) {
	return m.GetByShortFunc(ctx, short)
}
func (m *MockRepository) GetNextID(ctx context.Context) (uint64, error) { return m.GetNextIDFunc(ctx) }
func (m *MockRepository) GetByFull(ctx context.Context, full string) (entity.URL, error) {
	return m.GetByFullFunc(ctx, full)
}

func TestShortener_Create_Extended(t *testing.T) {
	ctx := context.Background()

	t.Run("success_new_url", func(t *testing.T) {
		repo := &MockRepository{
			GetByFullFunc: func(ctx context.Context, full string) (entity.URL, error) {
				return entity.URL{}, entity.ErrNotFound
			},
			GetNextIDFunc: func(ctx context.Context) (uint64, error) { return 100, nil },
			SaveFunc:      func(ctx context.Context, url entity.URL) error { return nil },
		}
		s := NewShortener(repo)
		_, err := s.Create(ctx, "https://ozon.ru")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("err_url_already_exists_in_db", func(t *testing.T) {
		expectedShort := "short12345"
		repo := &MockRepository{
			GetByFullFunc: func(ctx context.Context, full string) (entity.URL, error) {
				return entity.URL{Full: full, Short: expectedShort}, nil
			},
		}
		s := NewShortener(repo)
		short, err := s.Create(ctx, "https://ozon.ru")
		if err != nil || short != expectedShort {
			t.Errorf("expected existing short %s, got %s (err: %v)", expectedShort, short, err)
		}
	})

	t.Run("err_database_failure_on_id_generation", func(t *testing.T) {
		dbErr := errors.New("db connection lost")
		repo := &MockRepository{
			GetByFullFunc: func(ctx context.Context, full string) (entity.URL, error) {
				return entity.URL{}, entity.ErrNotFound
			},
			GetNextIDFunc: func(ctx context.Context) (uint64, error) {
				return 0, dbErr
			},
		}
		s := NewShortener(repo)
		_, err := s.Create(ctx, "https://ozon.ru")
		if !errors.Is(err, dbErr) {
			t.Errorf("expected db error, got %v", err)
		}
	})

	t.Run("handle_race_condition_on_save_conflict", func(t *testing.T) {
		existingShort := "conflict12"
		var calls int

		repo := &MockRepository{
			GetByFullFunc: func(ctx context.Context, full string) (entity.URL, error) {
				calls++
				if calls == 1 {
					// имитируем, что ссылки еще нет
					return entity.URL{}, entity.ErrNotFound
				}
				// имитируем, что ссылка появилась
				return entity.URL{Full: full, Short: existingShort}, nil
			},
			GetNextIDFunc: func(ctx context.Context) (uint64, error) {
				return 1, nil
			},
			SaveFunc: func(ctx context.Context, url entity.URL) error {
				// имитируем конфликт
				return entity.ErrConflict
			},
		}

		s := NewShortener(repo)
		short, err := s.Create(ctx, "https://ozon.ru")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if short != existingShort {
			t.Errorf("expected %s, got %s", existingShort, short)
		}

		if calls != 2 {
			t.Errorf("expected 2 calls to GetByFull, got %d", calls)
		}
	})
	t.Run("invalid_url_formats", func(t *testing.T) {
		testCases := []string{
			"",
			"just-text",
			"ftp://ozon.ru",
			"http://",
			"https://",
		}
		s := NewShortener(&MockRepository{})
		for _, tc := range testCases {
			_, err := s.Create(ctx, tc)
			if !errors.Is(err, entity.ErrInvalidFormat) {
				t.Errorf("input %s: expected ErrInvalidFormat, got %v", tc, err)
			}
		}
	})
}

func TestShortener_Get_Extended(t *testing.T) {
	ctx := context.Background()

	t.Run("err_invalid_id_length", func(t *testing.T) {
		s := NewShortener(&MockRepository{})
		_, err := s.Get(ctx, "too-short") // len = 9
		if !errors.Is(err, entity.ErrInvalidFormat) {
			t.Errorf("expected ErrInvalidFormat for wrong length, got %v", err)
		}
	})

	t.Run("err_not_found_in_db", func(t *testing.T) {
		repo := &MockRepository{
			GetByShortFunc: func(ctx context.Context, short string) (entity.URL, error) {
				return entity.URL{}, entity.ErrNotFound
			},
		}
		s := NewShortener(repo)
		_, err := s.Get(ctx, "1234567890")
		if !errors.Is(err, entity.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("err_db_internal_error", func(t *testing.T) {
		dbErr := errors.New("sql syntax error")
		repo := &MockRepository{
			GetByShortFunc: func(ctx context.Context, short string) (entity.URL, error) {
				return entity.URL{}, dbErr
			},
		}
		s := NewShortener(repo)
		_, err := s.Get(ctx, "1234567890")
		if !errors.Is(err, dbErr) {
			t.Errorf("expected internal db error, got %v", err)
		}
	})
}
