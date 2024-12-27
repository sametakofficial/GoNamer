package cli

import (
	"context"
	"errors"
	"os"

	"github.com/nouuu/gonamer/cmd/cli/handlers"
	"github.com/nouuu/gonamer/cmd/cli/ui"
	"github.com/nouuu/gonamer/conf"
	"github.com/nouuu/gonamer/internal/mediadata"
	"github.com/nouuu/gonamer/internal/mediarenamer"
	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/pterm/pterm"
)

type Cli struct {
	Config       conf.Config
	scanner      mediascanner.MediaScanner
	mediaRenamer *mediarenamer.MediaRenamer
	tvClient     mediadata.TvShowClient
	movieClient  mediadata.MovieClient
}

var ErrExit = errors.New("exit requested")

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
	movies, err := c.ScanMovies(ctx)
	if err != nil {
		ui.ShowError("Error scanning movies: %v", err)
		return
	}

	c.processMoviesList(ctx, movies)
}

func (c *Cli) processMoviesList(ctx context.Context, movies []mediascanner.Movie) {
	pb, _ := pterm.DefaultProgressbar.
		WithTotal(len(movies)).
		WithCurrent(1).
		WithTitle("Processing movies...").Start()
	defer pb.Stop()

	for i, movie := range movies {
		pb.Stop()

		// Affiche le titre du film en cours
		ui.ShowInfo("Processing movie %d/%d: %s", i+1, len(movies), pterm.Yellow(movie.OriginalFilename))

		// Recherche des suggestions
		suggestions, err := c.mediaRenamer.SuggestMovies(ctx, movie, c.Config.MaxResults)
		if err != nil {
			suggestions = mediarenamer.MovieSuggestions{Movie: movie}
			ui.ShowError("Error finding suggestions for %s: %v", movie.OriginalFilename, err)
		}

		// Création et exécution du handler
		handler := handlers.NewMovieHandler(
			handlers.NewBaseHandler(c.Config),
			suggestions,
			c.movieClient,
			c.mediaRenamer,
			func() error { return ErrExit },
		)

		if err := handler.Handle(ctx); err != nil {
			if errors.Is(err, ErrExit) {
				c.Exit()
			}
			ui.ShowError("Error handling movie: %v", err)
		}
		pb, _ = pterm.DefaultProgressbar.
			WithTotal(len(movies)).
			WithCurrent(i + 2).
			WithTitle("Processing movies...").
			Start()
	}

	ui.ShowSuccess("Finished processing movies")
}

func (c *Cli) processTvShow(ctx context.Context) {
	episodes, err := c.ScanTvEpisodes(ctx)
	if err != nil {
		ui.ShowError("Error scanning tv shows: %v", err)
		return
	}

	c.processEpisodesList(ctx, episodes)
}

func (c *Cli) processEpisodesList(ctx context.Context, episodes []mediascanner.Episode) {
	pb, _ := pterm.DefaultProgressbar.WithTotal(len(episodes)).WithCurrent(1).WithTitle("Processing episodes...").Start()
	defer pb.Stop()

	for i, episode := range episodes {
		// Stop la barre avant d'afficher le menu
		pb.Stop()

		// Affiche le titre de l'épisode en cours
		ui.ShowInfo("Processing episode %d/%d: %s", i+1, len(episodes), pterm.Yellow(episode.OriginalFilename))

		// Recherche des suggestions
		suggestions, err := c.mediaRenamer.SuggestEpisodes(ctx, episode, c.Config.MaxResults)
		if err != nil {
			suggestions = mediarenamer.EpisodeSuggestions{Episode: episode}
			ui.ShowError("Error finding suggestions for %s: %v", episode.OriginalFilename, err)
		}

		// Création et exécution du handler
		handler := handlers.NewTvShowHandler(
			handlers.NewBaseHandler(c.Config),
			suggestions,
			c.tvClient,
			c.mediaRenamer,
			func() error { return ErrExit },
		)

		if err := handler.Handle(ctx); err != nil {
			if errors.Is(err, ErrExit) {
				c.Exit()
			}
			ui.ShowError("Error handling episode: %v", err)
		}

		// Redémarre la barre après le menu
		pb, _ = pterm.DefaultProgressbar.
			WithTotal(len(episodes)).
			WithTitle("Processing episodes...").
			WithCurrent(i + 2).
			Start()
	}

	ui.ShowSuccess("Finished processing episodes")
}

func (c *Cli) Exit() {
	pterm.Info.Println("Exiting...")
	os.Exit(0)
}
