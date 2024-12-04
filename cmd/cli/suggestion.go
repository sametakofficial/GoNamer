package cli

import (
	"context"
	"sync"

	"github.com/nouuu/gonamer/internal/mediarenamer"
	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/pterm/pterm"
)

func (c *Cli) FindMoviesSuggestions(ctx context.Context, movies []mediascanner.Movie) ([]mediarenamer.MovieSuggestions, error) {
	pb, _ := pterm.DefaultProgressbar.WithTotal(len(movies)).WithTitle("Finding Suggestion...").Start()
	defer pb.Stop()
	var failedSuggestions []mediarenamer.MovieSuggestions
	mutex := &sync.Mutex{}

	suggestions := c.mediaRenamer.FindMovieSuggestions(ctx, movies, c.Config.MaxResults, func(suggestion mediarenamer.MovieSuggestions, err error) {
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

func (c *Cli) FindTvEpisodesSuggestions(ctx context.Context, episodes []mediascanner.Episode) ([]mediarenamer.EpisodeSuggestions, error) {
	pb, _ := pterm.DefaultProgressbar.WithTotal(len(episodes)).WithTitle("Finding Suggestion...").Start()
	defer pb.Stop()
	var failedSuggestions []mediarenamer.EpisodeSuggestions
	mutex := &sync.Mutex{}

	suggestions := c.mediaRenamer.FindEpisodeSuggestions(ctx, episodes, c.Config.MaxResults, func(suggestion mediarenamer.EpisodeSuggestions, err error) {
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
			pterm.Println(pterm.Red(suggestion.Episode.OriginalFilename))
		}
	}
	if c.Config.IncludeNotFound {
		return append(suggestions, failedSuggestions...), nil
	}

	return suggestions, nil
}
