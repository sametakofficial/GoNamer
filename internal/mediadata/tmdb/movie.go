package tmdb

import (
	"context"
	"strconv"

	"github.com/cyruzin/golang-tmdb"
	"github.com/nouuu/gonamer/internal/cache"
	"github.com/nouuu/gonamer/internal/mediadata"
	"github.com/nouuu/gonamer/pkg/logger"
)

func NewMovieClient(APIKey string, cache cache.Cache, opts ...OptFunc) (mediadata.MovieClient, error) {
	o := defaultOpts(APIKey)
	for _, optF := range opts {
		optF(&o.Opts)
	}

	client, err := tmdb.Init(o.APIKey)
	if err != nil {
		return nil, err
	}
	return &tmdbClient{client: client, cache: cache, opts: o}, nil
}

func (t *tmdbClient) SearchMovie(ctx context.Context, query string, year int, page int) (mediadata.MovieResults, error) {
	if result, err := t.cache.GetMovieSearch(ctx, query, year, page); err == nil {
		return result, nil
	}
	opts := map[string]string{
		"page": strconv.Itoa(page),
	}
	if year != 0 {
		opts["year"] = strconv.Itoa(year)
	}
	searchMovies, err := t.client.GetSearchMovies(query, cfgMap(t.opts, opts))
	if err != nil {
		return mediadata.MovieResults{}, err
	}
	results := mediadata.MovieResults{
		Movies:         buildMovieFromResult(searchMovies.SearchMoviesResults),
		Totals:         searchMovies.TotalResults,
		ResultsPerPage: 20,
	}

	if err := t.cache.SetMovieSearch(ctx, query, year, page, results); err != nil {
		logger.FromContext(ctx).With("error", err).Error("failed to cache movie search results")
	}

	return results, nil
}

func (t *tmdbClient) GetMovie(ctx context.Context, id string) (mediadata.Movie, error) {
	if movie, err := t.cache.GetMovie(ctx, id); err == nil {
		return movie, nil
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return mediadata.Movie{}, err
	}
	movieDetails, err := t.client.GetMovieDetails(idInt, cfgMap(t.opts))
	if err != nil {
		return mediadata.Movie{}, err
	}
	movie := buildMovie(movieDetails)
	if err := t.cache.SetMovie(ctx, id, movie); err != nil {
		logger.FromContext(ctx).With("error", err).Error("failed to cache movie")
	}
	return movie, nil
}

func (t *tmdbClient) GetMovieDetails(ctx context.Context, id string) (mediadata.MovieDetails, error) {
	if details, err := t.cache.GetMovieDetails(ctx, id); err == nil {
		return details, nil
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return mediadata.MovieDetails{}, err
	}
	movieDetails, err := t.client.GetMovieDetails(idInt, cfgMap(t.opts, map[string]string{
		"append_to_response": "credits",
	}))
	if err != nil {
		return mediadata.MovieDetails{}, err
	}
	details := buildMovieDetails(movieDetails)
	if err := t.cache.SetMovieDetails(ctx, id, details); err != nil {
		logger.FromContext(ctx).With("error", err).Error("failed to cache movie details")
	}
	return details, nil
}

func buildMovie(movie *tmdb.MovieDetails) mediadata.Movie {
	releaseYear := ""
	if len(movie.ReleaseDate) >= 4 {
		releaseYear = movie.ReleaseDate[:4]
	}
	return mediadata.Movie{
		ID:          strconv.FormatInt(movie.ID, 10),
		Title:       movie.Title,
		Overview:    movie.Overview,
		ReleaseDate: movie.ReleaseDate,
		Year:        releaseYear,
		PosterURL:   tmdbImageBaseUrl + movie.PosterPath,
		Rating:      movie.VoteAverage,
		RatingCount: movie.VoteCount,
	}
}

func buildMovieDetails(details *tmdb.MovieDetails) mediadata.MovieDetails {
	releaseYear := ""
	if len(details.ReleaseDate) >= 4 {
		releaseYear = details.ReleaseDate[:4]

	}
	return mediadata.MovieDetails{
		Movie: mediadata.Movie{
			ID:          strconv.FormatInt(details.ID, 10),
			Title:       details.Title,
			Overview:    details.Overview,
			ReleaseDate: details.ReleaseDate,
			Year:        releaseYear,
			PosterURL:   tmdbImageBaseUrl + details.PosterPath,
			Rating:      details.VoteAverage,
			RatingCount: details.VoteCount,
		},
		Runtime: details.Runtime,
		Genres:  buildGenres(details.Genres),
		Cast:    buildMovieCast(details.Credits.Cast),
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

func buildMovieCast(cast []struct {
	Adult              bool    `json:"adult"`
	CastID             int64   `json:"cast_id"`
	Character          string  `json:"character"`
	CreditID           string  `json:"credit_id"`
	Gender             int     `json:"gender"`
	ID                 int64   `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	Order              int     `json:"order"`
	OriginalName       string  `json:"original_name"`
	Popularity         float32 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
}) []mediadata.Person {
	var c = make([]mediadata.Person, len(cast))
	for i, person := range cast {
		c[i] = mediadata.Person{
			ID:         strconv.FormatInt(person.ID, 10),
			Name:       person.Name,
			Character:  person.Character,
			ProfileURL: tmdbImageBaseUrl + person.ProfilePath,
		}
	}
	return c
}
