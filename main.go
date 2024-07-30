package main

import (
	"context"

	"github.com/nouuu/mediatracker/internal/mediascanner"
	"github.com/nouuu/mediatracker/internal/mediascanner/filescanner"
	"github.com/nouuu/mediatracker/pkg/logger"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger.SetLoggerLevel(zapcore.ErrorLevel)
	ctx := context.Background()

	//config := conf.LoadConfig()

	scanner := filescanner.New()
	//movieClient, err := tmdb.NewMovieClient(config.TMDBAPIKey, tmdb.WithLang("fr-FR"))
	//if err != nil {
	//	log.Fatalf("Error creating movie client: %v", err)
	//}
	//mediaRenamer := mediarenamer.NewMediaRenamer(movieClient)

	scanner.ScanEpisodes(ctx, "/mnt/nfs/Download/direct_download/tv", mediascanner.ScanEpisodesOptions{Recursively: true})

}
