package mediascanner

type ScanMoviesOptions struct {
	Recursively bool
}

type ScanEpisodesOptions struct {
	Recursively bool
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
	ScanMovies(path string, options ...ScanMoviesOptions) ([]Movie, error)
	ScanEpisodes(path string, options ...ScanEpisodesOptions) ([]Episode, error)
}
