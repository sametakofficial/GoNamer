package cli

import (
	"context"
	"os"

	"github.com/nouuu/gonamer/conf"
	"github.com/nouuu/gonamer/internal/mediadata"
	"github.com/nouuu/gonamer/internal/mediarenamer"
	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/nouuu/gonamer/pkg/logger"
	"github.com/pterm/pterm"
)

type Cli struct {
	Config       conf.Config
	scanner      mediascanner.MediaScanner
	mediaRenamer *mediarenamer.MediaRenamer
	tvClient     mediadata.TvShowClient
	movieClient  mediadata.MovieClient
}

func clear() {
	print("\033[H\033[2J")
}

func NewCli(scanner mediascanner.MediaScanner, mediaRenamer *mediarenamer.MediaRenamer, movieClient mediadata.MovieClient, tvClient mediadata.TvShowClient) *Cli {
	return &Cli{
		Config:       conf.LoadConfig(),
		scanner:      scanner,
		mediaRenamer: mediaRenamer,
		movieClient:  movieClient,
		tvClient:     tvClient,
	}
}

func (c *Cli) Run(ctx context.Context) {
	switch c.Config.Type {
	case conf.Movie:
		c.processMovie(ctx)
	case conf.TvShow:
		c.processTvShow(ctx)
	}
}

func (c *Cli) processMovie(ctx context.Context) {
	log := logger.FromContext(ctx)

	movies, err := c.ScanMovies(ctx)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error scanning movies: %v", err))
		log.With("error", err).Error("Error scanning movies")
		return
	}
	suggestions, err := c.FindMoviesSuggestions(ctx, movies)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error finding suggestions: %v", err))
		log.With("error", err).Error("Error finding suggestions")
		return
	}
	err = c.RenameMovies(ctx, suggestions)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error renaming movies: %v", err))
		log.With("error", err).Error("Error renaming movies")
		return
	}
}

func (c *Cli) processTvShow(ctx context.Context) {
	log := logger.FromContext(ctx)

	tvEpisodes, err := c.ScanTvEpisodes(ctx)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error scanning tv shows: %v", err))
		log.With("error", err).Error("Error scanning tv shows")
		return
	}
	suggestions, err := c.FindTvEpisodesSuggestions(ctx, tvEpisodes)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error finding suggestions: %v", err))
		log.With("error", err).Error("Error finding suggestions")
		return
	}
	err = c.RenameTvEpisodes(ctx, suggestions)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error renaming tv shows: %v", err))
		log.With("error", err).Error("Error renaming tv shows")
		return
	}
}

func (c *Cli) Exit() {
	pterm.Info.Println("Exiting...")
	os.Exit(0)
}
