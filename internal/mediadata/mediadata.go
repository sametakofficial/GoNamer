package mediadata

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

type Status string

const (
	StatusReturning Status = "Returning Series"
	StatusEnded     Status = "Ended"
)

type Genre struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Person struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Character  string `json:"character"`
	ProfileURL string `json:"profile_url"`
}

type Studio struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Movie struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	ReleaseDate string  `json:"release_date"`
	Year        string  `json:"year"`
	PosterURL   string  `json:"poster_url"`
	Rating      float32 `json:"rating"`
	RatingCount int64   `json:"rating_count"`
}

type MovieDetails struct {
	Movie
	Runtime int      `json:"runtime"`
	Genres  []Genre  `json:"genres"`
	Cast    []Person `json:"cast"`
	Studio  []Studio `json:"studio"`
}

type MovieResults struct {
	Movies         []Movie `json:"movies"`
	Totals         int64   `json:"totals"`
	ResultsPerPage int64   `json:"results_per_page"`
}

type Season struct {
	SeasonNumber int    `json:"season_number"`
	EpisodeCount int    `json:"episode_count"`
	AirDate      string `json:"air_date"`
	PosterURL    string `json:"poster_url"`
}

type Episode struct {
	ID            string  `json:"id"`
	AirDate       string  `json:"air_date"`
	EpisodeNumber int     `json:"episode_number"`
	SeasonNumber  int     `json:"season_number"`
	Name          string  `json:"name"`
	Overview      string  `json:"overview"`
	StillURL      string  `json:"still_url"`
	VoteAverage   float32 `json:"vote_average"`
	VoteCount     int64   `json:"vote_count"`
}

type TvShow struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	FistAirDate string  `json:"first_air_date"`
	Year        string  `json:"year"`
	PosterURL   string  `json:"poster_url"`
	Rating      float32 `json:"rating"`
	RatingCount int64   `json:"rating_count"`
}

type TvShowDetails struct {
	TvShow
	SeasonCount  int      `json:"season_count"`
	EpisodeCount int      `json:"episode_count"`
	LastEpisode  Episode  `json:"last_episode"`
	NextEpisode  Episode  `json:"next_episode"`
	Status       Status   `json:"status"`
	Seasons      []Season `json:"seasons"`
	Genres       []Genre  `json:"genres"`
	Cast         []Person `json:"cast"`
	Studio       []Studio `json:"studio"`
}

type TvShowResults struct {
	TvShows        []TvShow `json:"tv_shows"`
	Totals         int64    `json:"totals"`
	ResultsPerPage int64    `json:"results_per_page"`
}

type MovieClient interface {
	SearchMovie(ctx context.Context, query string, year int, page int) (MovieResults, error)
	GetMovie(ctx context.Context, id string) (Movie, error)
	GetMovieDetails(ctx context.Context, id string) (MovieDetails, error)
}

type TvShowClient interface {
	SearchTvShow(ctx context.Context, query string, year int, page int) (TvShowResults, error)
	GetTvShow(ctx context.Context, id string) (TvShow, error)
	GetTvShowDetails(ctx context.Context, id string) (TvShowDetails, error)
	GetEpisode(ctx context.Context, id string, seasonNumber int, episodeNumber int) (Episode, error)
}

func ShowMovieResults(movies MovieResults) {
	slog.Info("Movies")
	for _, movie := range movies.Movies {
		mJson, err := marshalMovie(movie)
		if err != nil {
			slog.Error("Failed to marshal movie", slog.Any("error", err))
			os.Exit(1)
		}
		fmt.Println(string(mJson))
	}
}

func ShowTvShowResults(tvShows TvShowResults) {
	slog.Info("TvShows")
	for _, tvShow := range tvShows.TvShows {
		mJson, err := marshalTvShow(tvShow)
		if err != nil {
			slog.Error("Failed to marshal tv show", slog.Any("error", err))
			os.Exit(1)
		}
		fmt.Println(string(mJson))
	}
}

func marshalMovie(movie Movie) (string, error) {
	mJson, err := json.MarshalIndent(movie, "", "  ")
	if err != nil {
		return "", err
	}
	return string(mJson), nil
}

func marshalMovieDetails(movieDetails MovieDetails) (string, error) {
	mJson, err := json.MarshalIndent(movieDetails, "", "  ")
	if err != nil {
		return "", err
	}
	return string(mJson), nil
}

func marshalTvShow(tvShow TvShow) (string, error) {
	mJson, err := json.MarshalIndent(tvShow, "", "  ")
	if err != nil {
		return "", err
	}
	return string(mJson), nil
}

func marshalTvShowDetails(tvShowDetails TvShowDetails) (string, error) {
	mJson, err := json.MarshalIndent(tvShowDetails, "", "  ")
	if err != nil {
		return "", err
	}
	return string(mJson), nil
}
