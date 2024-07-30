package main

import (
	"context"
	"fmt"
	"os"
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

	movies, err := c.ScanMovies(ctx)
	if err != nil {
		return
	}
	suggestions, err := c.FindSuggestions(ctx, movies)
	if err != nil {
		return
	}
	err = c.RenameMovies(ctx, suggestions)

}

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

func (c *Cli) FindSuggestions(ctx context.Context, movies []mediascanner.Movie) ([]mediarenamer.MovieSuggestions, error) {
	pb, _ := pterm.DefaultProgressbar.WithTotal(len(movies)).WithTitle("Finding Suggestion...").Start()
	defer pb.Stop()
	var failedSuggestions []mediarenamer.MovieSuggestions
	mutex := &sync.Mutex{}

	suggestions := c.mediaRenamer.FindSuggestions(ctx, movies, c.Config.MaxResults, func(suggestion mediarenamer.MovieSuggestions, err error) {
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
	if c.Config.IncludeNotFound {
		return append(suggestions, failedSuggestions...), nil
	}

	return suggestions, nil
}

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

func (c *Cli) ProcessMovieSuggestions(ctx context.Context, suggestion mediarenamer.MovieSuggestions) error {
	if len(suggestion.SuggestedMovies) == 1 {
		if mediarenamer.GenerateMovieFilename(c.Config.MoviePattern, suggestion.SuggestedMovies[0], suggestion.Movie) == suggestion.Movie.OriginalFilename {
			pterm.Success.Println("Original filename is already correct for ", pterm.Yellow(suggestion.Movie.OriginalFilename))
		} else if c.Config.QuickMode {
			pterm.Success.Println("Quick - Renaming movie ", pterm.Yellow(suggestion.Movie.OriginalFilename), "to", pterm.Yellow(mediarenamer.GenerateMovieFilename(c.Config.MoviePattern, suggestion.SuggestedMovies[0], suggestion.Movie)))
		}
		return nil
	}

	return c.ProcessMovieSuggestionsOptions(ctx, suggestion)
}

func (c *Cli) ProcessMovieSuggestionsOptions(ctx context.Context, suggestion mediarenamer.MovieSuggestions) error {
	options := make(map[string]func() error)
	optionsArray := make([]string, 0)
	for i, movie := range suggestion.SuggestedMovies {
		key := fmt.Sprintf("%d. %s (%s)", i+1, movie.Title, movie.Year)
		options[key] = func() error {
			return c.RenameMovie(ctx, suggestion, movie)
		}
		optionsArray = append(optionsArray, key)
	}
	options[fmt.Sprintf("%d. Skip", len(suggestion.SuggestedMovies)+1)] = func() error {
		pterm.Info.Println("Skipping renaming of ", pterm.Yellow(suggestion.Movie.OriginalFilename))
		return nil
	}
	optionsArray = append(optionsArray, fmt.Sprintf("%d. Skip", len(suggestion.SuggestedMovies)+1))

	options[fmt.Sprintf("%d. Search Manually", len(suggestion.SuggestedMovies)+2)] = func() error {
		return c.SearchMovieSuggestionsManually(context.Background(), suggestion)
	}
	optionsArray = append(optionsArray, fmt.Sprintf("%d. Search Manually", len(suggestion.SuggestedMovies)+2))

	options[fmt.Sprintf("%d. Rename Manually", len(suggestion.SuggestedMovies)+3)] = func() error {
		pterm.Info.Println("Renaming manually for ", pterm.Yellow(suggestion.Movie.OriginalFilename))
		return nil
	}
	optionsArray = append(optionsArray, fmt.Sprintf("%d. Rename Manually", len(suggestion.SuggestedMovies)+3))

	options[fmt.Sprintf("%d. Exit", len(suggestion.SuggestedMovies)+4)] = func() error {
		c.Exit()
		return nil
	}
	optionsArray = append(optionsArray, fmt.Sprintf("%d. Exit", len(suggestion.SuggestedMovies)+4))

	selected, err := pterm.DefaultInteractiveSelect.WithMaxHeight(10).WithOptions(optionsArray).Show()
	if err != nil {
		pterm.Error.Println(pterm.Sprintf("Error selecting movie: %v", err))
		return err
	}

	return options[selected]()
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

func (c *Cli) SearchMovieSuggestionsManually(ctx context.Context, suggestions mediarenamer.MovieSuggestions) error {
	query, err := pterm.DefaultInteractiveTextInput.WithDefaultValue(suggestions.Movie.Name).Show(pterm.Sprintf("Search for '%s'", suggestions.Movie.OriginalFilename))
	if err != nil {
		pterm.Error.Println(pterm.Sprintf("Error getting search query: %v", err))
	}

	movies, err := c.movieClient.SearchMovie(query, 0, 1)
	if err != nil {
		pterm.Error.Println(pterm.Sprintf("Error searching movie: %v", err))
		return err
	}
	suggestions.SuggestedMovies = movies.Movies
	if len(suggestions.SuggestedMovies) > c.Config.MaxResults {
		suggestions.SuggestedMovies = suggestions.SuggestedMovies[:c.Config.MaxResults]
	}
	return c.ProcessMovieSuggestionsOptions(ctx, suggestions)
}

func (c *Cli) Exit() {
	pterm.Info.Println("Exiting...")
	os.Exit(0)
}
