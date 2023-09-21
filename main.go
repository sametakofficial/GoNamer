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
		mJson, err := json.MarshalIndent(movie, "", "  ")
		if err != nil {
			slog.Error("Failed to marshal movie", slog.Any("error", err))
			os.Exit(1)
		}
		fmt.Println(string(mJson))
	}
}

func showTvShowResults(tvShows mediadata.TvShowResults) {
	slog.Info("TvShows")
	for _, tvShow := range tvShows.TvShows {
		mJson, err := json.MarshalIndent(tvShow, "", "  ")
		if err != nil {
			slog.Error("Failed to marshal tv show", slog.Any("error", err))
			os.Exit(1)
		}
		fmt.Println(string(mJson))
	}
}
