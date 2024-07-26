package mediascanner

import "context"

type ScanMoviesOptions struct {
	Recursively bool
}

type ScanEpisodesOptions struct {
	Recursively     bool
	ExcludeUnparsed bool
}

type Movie struct {
	OriginalFilename string
	Name             string
	Year             int
	Extension        string
}

type Episode struct {
	OriginalFilename string
	Name             string
	Season           int
	Episode          int
	Extension        string
}

type MediaScanner interface {
	ScanMovies(ctx context.Context, path string, options ...ScanMoviesOptions) ([]Movie, error)
	ScanEpisodes(ctx context.Context, path string, options ...ScanEpisodesOptions) ([]Episode, error)
}
