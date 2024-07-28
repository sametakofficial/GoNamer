package tmdb

import (
	"strconv"

	"github.com/cyruzin/golang-tmdb"
	"github.com/nouuu/mediatracker/internal/mediadata"
)

func NewMovieClient(APIKey string, opts ...OptFunc) (mediadata.MovieClient, error) {
	o := defaultOpts(APIKey)
	for _, optF := range opts {
		optF(&o.Opts)
	}

	client, err := tmdb.Init(o.APIKey)
	if err != nil {
		return nil, err
	}
	return &tmdbClient{client: client, opts: o}, nil
}

func (t *tmdbClient) SearchMovie(query string, year int, page int) (mediadata.MovieResults, error) {
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
	movies := buildMovieFromResult(searchMovies.SearchMoviesResults)
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
