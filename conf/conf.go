package conf

import (
	"fmt"
	"os"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	TMDBAPIKey string `env:"TMDB_API_KEY" envDefault:"not-set"`
}

func LoadConfig() Config {
	_ = godotenv.Load()
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return cfg
}
