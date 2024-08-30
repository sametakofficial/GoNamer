package cli

import (
	"context"
	"sync"

	"github.com/nouuu/gonamer/internal/mediarenamer"
	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/pterm/pterm"
)

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
