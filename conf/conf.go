package conf

import (
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
)

type Config struct {
	DBUser     string `env:"DB_USER" envDefault:"postgres"`
	DBPassword string `env:"DB_PASSWORD" envDefault:"postgres"`
	DBName     string `env:"DB_NAME" envDefault:"postgres"`
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     string `env:"DB_PORT" envDefault:"5432"`
	DBSync     bool   `env:"DB_SYNC" envDefault:"false"`
	DBType     string `env:"DB_TYPE" envDefault:"postgres"`
	TMDBAPIKey string `env:"TMDB_API_KEY" envDefault:"not-set"`
}

func LoadConfig() Config {
	_ = godotenv.Load()
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Failed to parse ENV: %v", err)
		os.Exit(1)
	}
	return cfg
}
