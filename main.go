package main

import (
	"encoding/json"
	"fmt"
	"github.com/nouuu/mediatracker/conf"
	"github.com/nouuu/mediatracker/internal/mediadata/tmdb"
	"log/slog"
	"os"
)

func main() {
	config := conf.LoadConfig()
	movieClient := tmdb.NewMovieClient(config.TMDBAPIKey, tmdb.WithLang("fr-FR"))
	movies, err := movieClient.SearchMovie("The matrix", 1)
	if err != nil {
		slog.Error("Failed to search movie", slog.Any("error", err))
		os.Exit(1)
	}
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
