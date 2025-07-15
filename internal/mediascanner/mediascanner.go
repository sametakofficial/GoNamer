package mediascanner

import (
	"context"
	"github.com/nouuu/gonamer/pkg/config"
)
type ScanMoviesOptions struct {
	Recursively bool
}

type ScanEpisodesOptions struct {
	Recursively     bool
	ExcludeUnparsed bool
}

type Movie struct {
	OriginalFilename string
	FullPath         string
	Name             string
	Year             int
	Extension        string
	Quality          string
}

type Episode struct {
	OriginalFilename string
	FullPath         string
	Name             string
	Season           int
	Episode          int
	Extension        string
	Quality          string
}

type MediaScanner interface {
	ScanMovies(ctx context.Context, path string, cfg *config.Config, options ...ScanMoviesOptions) ([]Movie, error)
	ScanEpisodes(ctx context.Context, path string, cfg *config.Config, options ...ScanEpisodesOptions) ([]Episode, error)
}