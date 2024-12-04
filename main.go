package main

import (
	"context"
	"fmt"

	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/nouuu/gonamer/internal/mediascanner/filescanner"
	"github.com/nouuu/gonamer/pkg/logger"
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

	episodes, err := scanner.ScanEpisodes(ctx, "/mnt/nfs/Media/TV", mediascanner.ScanEpisodesOptions{Recursively: true})

	if err != nil {
		logger.FromContext(ctx).Errorf("Error scanning episodes: %v", err)
	}
	for _, episode := range episodes {
		fmt.Printf("Episode: %s %dx%02d\n", episode.Name, episode.Season, episode.Episode)
	}

}
