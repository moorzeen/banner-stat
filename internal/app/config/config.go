package config

import (
	"os"
)

type Config struct {
	Port string
	DB   string
}

func GetConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	db := os.Getenv("DATABASE_URL")
	if db == "" {
		db = "postgres://postgres:postgres@localhost:5432/banners?sslmode=disable"
	}

	return &Config{
		Port: port,
		DB:   db,
	}
}
