package main

import (
	"encoding/json"
	"fmt"
	"github.com/nouuu/mediatracker/conf"
	"github.com/nouuu/mediatracker/internal/mediadata"
	"github.com/nouuu/mediatracker/internal/mediadata/tmdb"
	"log/slog"
	"os"
)

func main() {
	config := conf.LoadConfig()
	movieClient := tmdb.NewMovieClient(
		config.TMDBAPIKey,
		tmdb.WithLang("fr-FR"),
		tmdb.WithAdult(false),
	)
	tvShowClient := tmdb.NewTvShowClient(
		config.TMDBAPIKey,
		tmdb.WithLang("fr-FR"),
		tmdb.WithAdult(false),
	)
	movies, err := movieClient.SearchMovie("Tarzan", 1)
	if err != nil {
		slog.Error("Failed to search movie", slog.Any("error", err))
		os.Exit(1)
	}
	showMovieResults(movies)
	movie, err := movieClient.GetMovie(movies.Movies[0].ID)
	if err != nil {
		slog.Error("Failed to get movie", slog.Any("error", err))
		os.Exit(1)
	}
	movieDetails, err := movieClient.GetMovieDetails(movies.Movies[0].ID)
	if err != nil {
		slog.Error("Failed to get movie details", slog.Any("error", err))
		os.Exit(1)
	}
	mJson, err := marshalMovieDetails(movieDetails)
	if err != nil {
		slog.Error("Failed to marshal movie details", slog.Any("error", err))
		os.Exit(1)
	}
	fmt.Println(mJson)
	showMovieResults(mediadata.MovieResults{Movies: []mediadata.Movie{movie}})
	tvShows, err := tvShowClient.SearchTvShow("Bleach", 1)
	if err != nil {
		slog.Error("Failed to search tv show", slog.Any("error", err))
		os.Exit(1)
	}
	showTvShowResults(tvShows)
}

func showMovieResults(movies mediadata.MovieResults) {
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

func marshalMovie(movie mediadata.Movie) (string, error) {
	mJson, err := json.MarshalIndent(movie, "", "  ")
	if err != nil {
		return "", err
	}
	return string(mJson), nil
}

func marshalMovieDetails(movieDetails mediadata.MovieDetails) (string, error) {
	mJson, err := json.MarshalIndent(movieDetails, "", "  ")
	if err != nil {
		return "", err
	}
	return string(mJson), nil
}

func showTvShowResults(tvShows mediadata.TvShowResults) {
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

func marshalTvShow(tvShow mediadata.TvShow) (string, error) {
	mJson, err := json.MarshalIndent(tvShow, "", "  ")
	if err != nil {
		return "", err
	}
	return string(mJson), nil
}

func marshalTvShowDetails(tvShowDetails mediadata.TvShowDetails) (string, error) {
	mJson, err := json.MarshalIndent(tvShowDetails, "", "  ")
	if err != nil {
		return "", err
	}
	return string(mJson), nil
}
