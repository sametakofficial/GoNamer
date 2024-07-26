package main

import (
	"github.com/nouuu/mediatracker/conf"
	"github.com/nouuu/mediatracker/internal/logger"
	"github.com/nouuu/mediatracker/internal/mediascanner"
	"github.com/nouuu/mediatracker/internal/mediascanner/filescanner"
)

var log = logger.GetLogger()

func main() {
	_ = conf.LoadConfig()
	scanner := filescanner.New()
	movies, err := scanner.ScanMovies("/mnt/nfs/Media/Films", mediascanner.ScanMoviesOptions{Recursively: true})
	if err != nil {
		log.Error("Error scanning movies", "error", err)
		return
	}
	for _, movie := range movies {
		log.Infof("Movie: %s (%d)", movie.Name, movie.Year)
	}
	return
	episodes, err := scanner.ScanEpisodes("/mnt/nfs/Download/direct_download/tv", mediascanner.ScanEpisodesOptions{Recursively: true})
	if err != nil {
		log.Error("Error scanning episodes", "error", err)
		return
	}
	for _, episode := range episodes {
		log.Infof("Episode: %s S%02dE%02d (original: %s)", episode.Name, episode.Season, episode.Episode, episode.OriginalFilename)
	}
}
