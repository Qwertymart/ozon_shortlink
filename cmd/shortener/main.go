package main

import (
	"context"
	"embed"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Qwertymart/ozon_shortlink/internal/config"
	"github.com/Qwertymart/ozon_shortlink/internal/repository/postgres"
	transport "github.com/Qwertymart/ozon_shortlink/internal/transport/http"
	"github.com/Qwertymart/ozon_shortlink/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	runMigrations(pool)

	repo := postgres.NewRepository(pool)
	service := usecase.NewShortener(repo)
	handler := transport.NewHandler(service)

	router := handler.InitRoutes()

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second, // максимально время чтения
		WriteTimeout: 10 * time.Second, // максимальное время записи
		IdleTimeout:  30 * time.Second, // keep-alive
	}

	go func() {
		log.Printf("Starting HTTP server on %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Listen and serve error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func runMigrations(pool *pgxpool.Pool) {
	db := stdlib.OpenDBFromPool(pool)
	// закрываем только обертку stdlib
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set goose dialect: %v", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations are up to date")
}
