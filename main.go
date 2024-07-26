package main

import (
	"context"
	"time"

	"github.com/nouuu/mediatracker/conf"
	"github.com/nouuu/mediatracker/internal/logger"
	"github.com/nouuu/mediatracker/internal/mediascanner"
	"github.com/nouuu/mediatracker/internal/mediascanner/filescanner"
)

func main() {
	//logger.SetLoggerLevel(zapcore.InfoLevel)
	ctx := context.Background()
	log := logger.FromContext(ctx)
	start := time.Now()

	_ = conf.LoadConfig()
	log.Infof("Config loaded in %s", time.Since(start))

	start = time.Now()
	scanner := filescanner.New()
	log.Infof("File scanner initialized in %s", time.Since(start))

	start = time.Now()
	movies, err := scanner.ScanMovies(ctx, "/mnt/nfs/Media/Films", mediascanner.ScanMoviesOptions{Recursively: true})
	if err != nil {
		log.Infof("Error scanning movies: %v", err)
		return
	}
	log.Infof("%d movies scanned in %s", len(movies), time.Since(start))

	start = time.Now()
	episodes, err := scanner.ScanEpisodes(ctx, "/mnt/nfs/Media/Anime", mediascanner.ScanEpisodesOptions{Recursively: true})
	if err != nil {
		log.Infof("Error scanning episodes: %v", err)
		return
	}
	log.Infof("%d episodes scanned in %s", len(episodes), time.Since(start))
}
