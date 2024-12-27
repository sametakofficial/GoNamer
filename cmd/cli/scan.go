package cli

import (
	"context"

	"github.com/nouuu/gonamer/cmd/cli/ui"
	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/pterm/pterm"
)

func (c *Cli) ScanMovies(ctx context.Context) ([]mediascanner.Movie, error) {
	ui.ShowInfo("Scanning movies in '%s'...", c.Config.MediaPath)
	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start("Scanning movies...")
	defer ui.HandleSpinnerStop(spinner)
	movies, err := c.scanner.ScanMovies(ctx, c.Config.MediaPath, mediascanner.ScanMoviesOptions{Recursively: c.Config.Recursive})
	if err != nil {
		ui.ShowError("Error scanning movies: %v", err)
		return nil, err
	}
	spinner.Success(pterm.Sprintf("Found %d movies", len(movies)))
	return movies, nil
}

func (c *Cli) ScanTvEpisodes(ctx context.Context) ([]mediascanner.Episode, error) {
	ui.ShowInfo("Scanning TV shows in '%s'...", c.Config.MediaPath)
	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start("Scanning TV shows...")
	defer ui.HandleSpinnerStop(spinner)
	tvShows, err := c.scanner.ScanEpisodes(ctx, c.Config.MediaPath, mediascanner.ScanEpisodesOptions{Recursively: c.Config.Recursive})
	if err != nil {
		ui.ShowError("Error scanning TV shows: %v", err)
		return nil, err
	}
	spinner.Success(pterm.Sprintf("Found %d TV shows", len(tvShows)))
	return tvShows, nil
}
