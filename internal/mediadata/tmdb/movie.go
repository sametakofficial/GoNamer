package tmdb

import (
	"github.com/cyruzin/golang-tmdb"
	"github.com/nouuu/mediatracker/internal/mediadata"
	"log/slog"
	"strconv"
)

func NewMovieClient(APIKey string, opts ...OptFunc) mediadata.MovieClient {
	o := defaultOpts(APIKey)
	for _, optF := range opts {
		optF(&o.Opts)
	}

	client, err := tmdb.Init(o.APIKey)
	if err != nil {
		slog.Error("Failed to initialize TMDB client", slog.Any("error", err))
	}
	return &tmdbClient{client: client, opts: o}
}

func (t *tmdbClient) SearchMovie(query string, page int) (mediadata.MovieResults, error) {
	searchMovies, err := t.client.GetSearchMovies(query, cfgMap(t.opts, map[string]string{
		"page": strconv.Itoa(page),
	}))
	if err != nil {
		return mediadata.MovieResults{}, err
	}
	var movies []mediadata.Movie = buildMovieFromResult(searchMovies.SearchMoviesResults)
	return mediadata.MovieResults{
		Movies:         movies,
		Totals:         searchMovies.TotalResults,
		ResultsPerPage: 20,
	}, nil
}

func (t *tmdbClient) GetMovie(id string) (mediadata.Movie, error) {
	var idInt int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return mediadata.Movie{}, err
	}
	movieDetails, err := t.client.GetMovieDetails(
		idInt,
		cfgMap(t.opts),
	)
	if err != nil {
		return mediadata.Movie{}, err
	}
	return buildMovie(movieDetails), nil
}

func (t *tmdbClient) GetMovieDetails(id string) (mediadata.MovieDetails, error) {
	var idInt int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return mediadata.MovieDetails{}, err
	}
	movieDetails, err := t.client.GetMovieDetails(
		idInt,
		cfgMap(t.opts, map[string]string{
			"append_to_response": "credits",
		}),
	)
	if err != nil {
		return mediadata.MovieDetails{}, err
	}
	return buildMovieDetails(movieDetails), nil
}

func buildMovie(movie *tmdb.MovieDetails) mediadata.Movie {
	return mediadata.Movie{
		ID:          strconv.FormatInt(movie.ID, 10),
		Title:       movie.Title,
		Overview:    movie.Overview,
		ReleaseDate: movie.ReleaseDate,
		PosterURL:   tmdbImageBaseUrl + movie.PosterPath,
		Rating:      movie.VoteAverage,
		RatingCount: movie.VoteCount,
	}
}

func buildMovieDetails(details *tmdb.MovieDetails) mediadata.MovieDetails {
	return mediadata.MovieDetails{
		Movie: mediadata.Movie{
			ID:          strconv.FormatInt(details.ID, 10),
			Title:       details.Title,
			Overview:    details.Overview,
			ReleaseDate: details.ReleaseDate,
			PosterURL:   tmdbImageBaseUrl + details.PosterPath,
			Rating:      details.VoteAverage,
			RatingCount: details.VoteCount,
		},
		Runtime: details.Runtime,
		Genres:  buildGenres(details.Genres),
		Cast:    buildCast(details.Credits.Cast),
		Studio:  buildStudio(details.ProductionCompanies),
	}
}

func buildMovieFromResult(result *tmdb.SearchMoviesResults) []mediadata.Movie {
	var movies = make([]mediadata.Movie, len(result.Results))
	for i, movie := range result.Results {
		movies[i] = buildMovie(&tmdb.MovieDetails{
			ID:          movie.ID,
			Title:       movie.Title,
			Overview:    movie.Overview,
			ReleaseDate: movie.ReleaseDate,
			PosterPath:  movie.PosterPath,
			VoteAverage: movie.VoteAverage,
			VoteCount:   movie.VoteCount,
		})
	}
	return movies
}
