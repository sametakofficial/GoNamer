package main

import (
	"context"

	"github.com/nouuu/mediatracker/conf"
	"github.com/nouuu/mediatracker/internal/logger"
	"github.com/nouuu/mediatracker/internal/mediadata/tmdb"
	"github.com/nouuu/mediatracker/internal/mediarenamer"
	"github.com/nouuu/mediatracker/internal/mediascanner"
	"github.com/nouuu/mediatracker/internal/mediascanner/filescanner"
)

func main() {
	//logger.SetLoggerLevel(zapcore.InfoLevel)
	ctx := context.Background()
	log := logger.FromContext(ctx)

	config := conf.LoadConfig()

	scanner := filescanner.New()
	movieClient, err := tmdb.NewMovieClient(config.TMDBAPIKey, tmdb.WithLang("fr-FR"))
	if err != nil {
		log.Fatalf("Error creating movie client: %v", err)
	}
	mediaRenamer := mediarenamer.NewMediaRenamer(movieClient)

	movies, err := scanner.ScanMovies(ctx, "/mnt/nfs/Media/Films", mediascanner.ScanMoviesOptions{Recursively: true})
	if err != nil {
		log.Fatalf("Error scanning movies: %v", err)
	}
	_, err = mediaRenamer.RenameMovies(ctx, movies)
	if err != nil {
		log.Fatalf("Error renaming movies: %v", err)
	}
}
