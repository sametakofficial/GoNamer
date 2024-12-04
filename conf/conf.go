package conf

import (
	"fmt"
	"os"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type MediaType string

const (
	Movie  MediaType = "movie"
	TvShow           = "tvshow"
)

type Config struct {
	TMDBAPIKey      string    `env:"TMDB_API_KEY" envDefault:"not-set"`
	MediaPath       string    `env:"MEDIA_PATH" envDefault:"./"`
	Recursive       bool      `env:"RECURSIVE" envDefault:"true"`
	IncludeNotFound bool      `env:"INCLUDE_NOT_FOUND" envDefault:"false"`
	DryRun          bool      `env:"DRY_RUN" envDefault:"true"`
	MoviePattern    string    `env:"MOVIE_PATTERN" envDefault:"{name} - {year}{extension}"`
	TvShowPattern   string    `env:"TVSHOW_PATTERN" envDefault:"{name} - {season}x{episode}{extension}"`
	Type            MediaType `env:"TYPE" envDefault:"movie"`
	MaxResults      int       `env:"MAX_RESULTS" envDefault:"5"`
	QuickMode       bool      `env:"QUICK_MODE" envDefault:"false"`
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
