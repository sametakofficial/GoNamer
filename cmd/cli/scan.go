package cli

import (
	"context"

	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/pterm/pterm"
)

func (c *Cli) ScanMovies(ctx context.Context) ([]mediascanner.Movie, error) {
	pterm.Info.Printfln("Scanning movies in '%s'...", c.Config.MoviePath)
	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start("Scanning movies...")
	defer spinner.Stop()
	movies, err := c.scanner.ScanMovies(ctx, c.Config.MoviePath, mediascanner.ScanMoviesOptions{Recursively: c.Config.Recursive})
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error scanning movies: %v", err))
		return nil, err
	}
	spinner.Success(pterm.Sprintf("Found %d movies", len(movies)))
	return movies, nil
}
