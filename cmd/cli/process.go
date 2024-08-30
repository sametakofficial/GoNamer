package cli

import (
	"context"
	"fmt"

	"github.com/nouuu/gonamer/internal/mediarenamer"
	"github.com/pterm/pterm"
)

func (c *Cli) ProcessMovieSuggestions(ctx context.Context, suggestion mediarenamer.MovieSuggestions) error {
	if len(suggestion.SuggestedMovies) == 1 {
		if mediarenamer.GenerateMovieFilename(c.Config.MoviePattern, suggestion.SuggestedMovies[0], suggestion.Movie) == suggestion.Movie.OriginalFilename {
			pterm.Success.Println("Original filename is already correct for ", pterm.Yellow(suggestion.Movie.OriginalFilename))
			return nil
		} else if c.Config.QuickMode {
			pterm.Success.Println("Quick - Renaming movie ", pterm.Yellow(suggestion.Movie.OriginalFilename), "to", pterm.Yellow(mediarenamer.GenerateMovieFilename(c.Config.MoviePattern, suggestion.SuggestedMovies[0], suggestion.Movie)))
			return c.RenameMovie(ctx, suggestion, suggestion.SuggestedMovies[0])
		}
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

	optionIndex := len(suggestion.SuggestedMovies) + 1

	options[fmt.Sprintf("%d. Skip", optionIndex)] = func() error {
		pterm.Info.Println("Skipping renaming of ", pterm.Yellow(suggestion.Movie.OriginalFilename))
		return nil
	}
	optionsArray = append(optionsArray, fmt.Sprintf("%d. Skip", optionIndex))
	optionIndex++

	options[fmt.Sprintf("%d. Search Manually", optionIndex)] = func() error {
		return c.SearchMovieSuggestionsManually(context.Background(), suggestion)
	}
	optionsArray = append(optionsArray, fmt.Sprintf("%d. Search Manually", optionIndex))
	optionIndex++

	options[fmt.Sprintf("%d. Rename Manually", optionIndex)] = func() error {
		return c.RenameMovieFileManually(ctx, suggestion)
	}
	optionsArray = append(optionsArray, fmt.Sprintf("%d. Rename Manually", optionIndex))
	optionIndex++

	options[fmt.Sprintf("%d. Exit", optionIndex)] = func() error {
		c.Exit()
		return nil
	}
	optionsArray = append(optionsArray, fmt.Sprintf("%d. Exit", optionIndex))
	optionIndex++

	selected, err := pterm.DefaultInteractiveSelect.WithMaxHeight(10).WithOptions(optionsArray).Show()
	if err != nil {
		pterm.Error.Println(pterm.Sprintf("Error selecting movie: %v", err))
		return err
	}

	return options[selected]()
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
