package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseDSN string
	ServerPort  string
	StorageType string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading configuration from environment")
	}

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN environment variable is required")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	storageType := os.Getenv("STORAGE_TYPE")
    if storageType == "" {
        storageType = "postgres" 
    }

	return &Config{
		DatabaseDSN: dsn,
		ServerPort:  port,
		StorageType: storageType,
	}
}