package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/nouuu/gonamer/internal/mediadata"
	"github.com/nouuu/gonamer/internal/mediarenamer"
	"github.com/pterm/pterm"
)

func (c *Cli) RenameMovies(ctx context.Context, suggestions []mediarenamer.MovieSuggestions) error {
	slices.SortFunc(suggestions, func(i, j mediarenamer.MovieSuggestions) int {
		return strings.Compare(i.Movie.OriginalFilename, j.Movie.OriginalFilename)
	})

	pb, _ := pterm.DefaultProgressbar.WithTotal(len(suggestions)).WithTitle("Renaming movies...").Start()
	for _, suggestion := range suggestions {
		pterm.Print("\n")
		pb.UpdateTitle("Renaming " + suggestion.Movie.OriginalFilename)
		pb.Increment()
		pb, _ = pb.Stop()

		err := c.ProcessMovieSuggestions(ctx, suggestion)
		if err != nil {
			return err
		}

		pb, _ = pb.Start()
	}
	pb.Stop()
	pterm.Success.Println("Finished renaming movies")
	return nil
}

func (c *Cli) RenameMovie(ctx context.Context, suggestion mediarenamer.MovieSuggestions, movie mediadata.Movie) error {
	pterm.Info.Println("Renaming movie ", pterm.Yellow(suggestion.Movie.OriginalFilename), "to", pterm.Yellow(mediarenamer.GenerateMovieFilename(c.Config.MoviePattern, movie, suggestion.Movie)))
	err := c.mediaRenamer.RenameMovie(ctx, suggestion.Movie, movie, c.Config.MoviePattern, c.Config.DryRun)
	if err != nil {
		pterm.Error.Println(pterm.Sprintf("Error renaming movie: %v", err))
		return err
	}
	return nil
}

func (c *Cli) RenameMovieFileManually(ctx context.Context, suggestion mediarenamer.MovieSuggestions) error {
	pterm.Info.Println("Renaming manually for ", pterm.Yellow(suggestion.Movie.OriginalFilename))
	result, _ := pterm.DefaultInteractiveTextInput.
		WithDefaultText("Enter new filename (without extension) :").
		WithDefaultValue(fmt.Sprintf("%s (%d)", suggestion.Movie.Name, suggestion.Movie.Year)).
		Show()

	filename := fmt.Sprintf("%s.%s", result, suggestion.Movie.Extension)

	pterm.Info.Println("Renaming movie ", pterm.Yellow(suggestion.Movie.OriginalFilename), "to", pterm.Yellow(filename))

	return c.mediaRenamer.RenameFile(
		ctx,
		suggestion.Movie.FullPath,
		filepath.Join(filepath.Dir(suggestion.Movie.FullPath), filename),
		c.Config.DryRun,
	)
}
