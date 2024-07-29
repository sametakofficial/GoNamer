package main

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/nouuu/mediatracker/conf"
	"github.com/nouuu/mediatracker/internal/mediadata"
	"github.com/nouuu/mediatracker/internal/mediarenamer"
	"github.com/nouuu/mediatracker/internal/mediascanner"
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

	movies, err := c.ScanMovies(ctx, "/mnt/nfs/Media/Films", true)
	if err != nil {
		return
	}
	suggestions, err := c.FindSuggestions(ctx, movies)
	if err != nil {
		return
	}
	err = c.RenameMovies(ctx, suggestions, "{name} - {year}{extension}", false)

}

func (c *Cli) ScanMovies(ctx context.Context, path string, recursive bool) ([]mediascanner.Movie, error) {
	pterm.Info.Printfln("Scanning movies...")
	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start("Scanning movies...")
	defer spinner.Stop()
	movies, err := c.scanner.ScanMovies(ctx, path, mediascanner.ScanMoviesOptions{Recursively: true})
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error scanning movies: %v", err))
		return nil, err
	}
	spinner.Success(pterm.Sprintf("Found %d movies", len(movies)))
	return movies, nil
}

func (c *Cli) FindSuggestions(ctx context.Context, movies []mediascanner.Movie) ([]mediarenamer.MovieSuggestions, error) {
	pb, _ := pterm.DefaultProgressbar.WithTotal(len(movies)).WithTitle("Finding Suggestion...").Start()
	defer pb.Stop()
	var failedSuggestions []mediarenamer.MovieSuggestions
	mutex := &sync.Mutex{}

	suggestions := c.mediaRenamer.FindSuggestions(ctx, movies, func(suggestion mediarenamer.MovieSuggestions, err error) {
		pb.Increment()
		if err != nil {
			mutex.Lock()
			failedSuggestions = append(failedSuggestions, suggestion)
			mutex.Unlock()
		}
	})
	pterm.Success.Println("Finished finding suggestions")
	if len(failedSuggestions) > 0 {
		pterm.Error.Println("Some suggestions failed :")
		for _, suggestion := range failedSuggestions {
			pterm.Println(pterm.Red(suggestion.Movie.OriginalFilename))
		}
	}

	return suggestions, nil
}

func (c *Cli) RenameMovies(ctx context.Context, suggestions []mediarenamer.MovieSuggestions, pattern string, dryrun bool) error {
	slices.SortFunc(suggestions, func(i, j mediarenamer.MovieSuggestions) int {
		return strings.Compare(i.Movie.OriginalFilename, j.Movie.OriginalFilename)
	})

	pb, _ := pterm.DefaultProgressbar.WithTotal(len(suggestions)).WithTitle("Renaming movies...").Start()
	for _, suggestion := range suggestions {
		pterm.Print("\n")
		pb.UpdateTitle("Renaming " + suggestion.Movie.OriginalFilename)
		pb.Increment()
		pb, _ = pb.Stop()

		if len(suggestion.SuggestedMovies) == 1 &&
			mediarenamer.GenerateMovieFilename(pattern, suggestion.SuggestedMovies[0], suggestion.Movie) == suggestion.Movie.OriginalFilename {
			pterm.Success.Println("Original filename is already correct for ", pterm.Yellow(suggestion.Movie.OriginalFilename))
			pb, _ = pb.Start()
			continue
		}

		options := make(map[string]func())
		optionsArray := make([]string, 0)
		for _, movie := range suggestion.SuggestedMovies {
			key := fmt.Sprintf("%s (%s)", movie.Title, movie.Year)
			options[key] = func() {
				pterm.Success.Println("Renaming movie ", pterm.Yellow(suggestion.Movie.OriginalFilename), "to", pterm.Yellow(mediarenamer.GenerateMovieFilename(pattern, movie, suggestion.Movie)))
			}
			optionsArray = append(optionsArray, key)
		}

		selected, err := pterm.DefaultInteractiveSelect.WithOptions(optionsArray).Show("Test box")
		if err != nil {
			pterm.Error.Println(pterm.Sprintf("Error selecting movie: %v", err))
			pb, _ = pb.Start()
			continue
		}
		options[selected]()
		pb, _ = pb.Start()
	}
	pb.Stop()
	pterm.Success.Println("Finished renaming movies")
	return nil
}
