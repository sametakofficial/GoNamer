package mediadata

import (
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
	ID   string
	Name string
}

type Person struct {
	ID         string
	Name       string
	Character  string
	ProfileURL string
}

type Studio struct {
	ID   string
	Name string
}

type Movie struct {
	ID          string
	Title       string
	Overview    string
	ReleaseDate string
	PosterURL   string
	Rating      float32
	RatingCount int64
}

type MovieDetails struct {
	Movie
	Runtime int
	Genres  []Genre
	Cast    []Person
	Studio  []Studio
}

type MovieResults struct {
	Movies         []Movie
	Totals         int64
	ResultsPerPage int64
}

type Season struct {
	SeasonNumber int
	EpisodeCount int
	AirDate      string
	PosterURL    string
}

type Episode struct {
	ID            string
	AirDate       string
	EpisodeNumber int
	SeasonNumber  int
	Name          string
	Overview      string
	StillURL      string
	VoteAverage   float32
	VoteCount     int64
}

type TvShow struct {
	ID          string
	Title       string
	Overview    string
	FistAirDate string
	PosterURL   string
	Rating      float32
	RatingCount int64
}

type TvShowDetails struct {
	TvShow
	SeasonCount  int
	EpisodeCount int
	LastEpisode  Episode
	NextEpisode  Episode
	Status       Status
	Seasons      []Season
	Genres       []Genre
	Cast         []Person
	Studio       []Studio
}

type TvShowResults struct {
	TvShows        []TvShow
	Totals         int64
	ResultsPerPage int64
}

type MovieClient interface {
	SearchMovie(query string, page int) (MovieResults, error)
	GetMovie(id string) (Movie, error)
	GetMovieDetails(id string) (MovieDetails, error)
}

type TvShowClient interface {
	SearchTvShow(query string, page int) (TvShowResults, error)
	GetTvShow(id string) (TvShow, error)
	GetTvShowDetails(id string) (TvShowDetails, error)
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
