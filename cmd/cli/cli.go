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
	movieClient  mediadata.MovieClient
}

func clear() {
	print("\033[H\033[2J")
}

func NewCli(scanner mediascanner.MediaScanner, mediaRenamer *mediarenamer.MediaRenamer, movieClient mediadata.MovieClient) *Cli {
	return &Cli{
		Config:       conf.LoadConfig(),
		scanner:      scanner,
		mediaRenamer: mediaRenamer,
		movieClient:  movieClient,
	}
}

func (c *Cli) Run(ctx context.Context) {
	log := logger.FromContext(ctx)

	movies, err := c.ScanMovies(ctx)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error scanning movies: %v", err))
		log.With("error", err).Error("Error scanning movies")
		return
	}
	suggestions, err := c.FindSuggestions(ctx, movies)
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

func (c *Cli) Exit() {
	pterm.Info.Println("Exiting...")
	os.Exit(0)
}
