package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/nouuu/gonamer/cmd/cli/handlers"
	"github.com/nouuu/gonamer/cmd/cli/ui"
	"github.com/nouuu/gonamer/internal/mediadata"
	"github.com/nouuu/gonamer/internal/mediarenamer"
	"github.com/pterm/pterm"
)

type MediaSuggestion interface {
	GetOriginalFilename() string
	GenerateFilename(pattern string) string
	RenameFileManually(ctx context.Context) error
	HasSingleSuggestion() bool
}

// MovieSuggestionWrapper implémente MediaSuggestion pour les films
type MovieSuggestionWrapper struct {
	suggestion mediarenamer.MovieSuggestions
	pattern    string
	renameFunc func(context.Context, mediarenamer.MovieSuggestions, mediadata.Movie) error
}

func (m MovieSuggestionWrapper) GetOriginalFilename() string {
	return m.suggestion.Movie.OriginalFilename
}

func (m MovieSuggestionWrapper) GenerateFilename(pattern string) string {
	if len(m.suggestion.SuggestedMovies) == 0 {
		return ""
	}
	return mediarenamer.GenerateMovieFilename(pattern, m.suggestion.SuggestedMovies[0], m.suggestion.Movie)
}

func (m MovieSuggestionWrapper) RenameFileManually(ctx context.Context) error {
	return m.renameFunc(ctx, m.suggestion, m.suggestion.SuggestedMovies[0])
}

func (m MovieSuggestionWrapper) HasSingleSuggestion() bool {
	return len(m.suggestion.SuggestedMovies) == 1
}

type EpisodeSuggestionWrapper struct {
	suggestions mediarenamer.EpisodeSuggestions
	pattern     string
	renameFunc  func(context.Context, mediarenamer.EpisodeSuggestions, mediadata.TvShow, mediadata.Episode) error
}

func (e EpisodeSuggestionWrapper) GetOriginalFilename() string {
	return e.suggestions.Episode.OriginalFilename
}

func (e EpisodeSuggestionWrapper) GenerateFilename(pattern string) string {
	if len(e.suggestions.SuggestedEpisodes) == 0 {
		return ""
	}
	return mediarenamer.GenerateEpisodeFilename(
		pattern,
		e.suggestions.SuggestedEpisodes[0].TvShow,
		e.suggestions.SuggestedEpisodes[0].Episode,
		e.suggestions.Episode,
	)
}

func (e EpisodeSuggestionWrapper) HasSingleSuggestion() bool {
	return len(e.suggestions.SuggestedEpisodes) == 1
}

func (e EpisodeSuggestionWrapper) RenameFileManually(ctx context.Context) error {
	return e.renameFunc(ctx, e.suggestions, e.suggestions.SuggestedEpisodes[0].TvShow, e.suggestions.SuggestedEpisodes[0].Episode)
}

func (c *Cli) processSuggestion(ctx context.Context, suggestion MediaSuggestion, pattern string, processFunc func() error) error {
	if !suggestion.HasSingleSuggestion() {
		return processFunc()
	}

	if c.Config.QuickMode {
		newFilename := suggestion.GenerateFilename(pattern)
		pterm.Success.Println("Quick - Renaming file ", pterm.Yellow(suggestion.GetOriginalFilename()), "to", pterm.Yellow(newFilename))
		return suggestion.RenameFileManually(ctx)
	}

	return processFunc()
}

func (c *Cli) ProcessMovieSuggestions(ctx context.Context, suggestions mediarenamer.MovieSuggestions) error {
	handler := handlers.NewMovieHandler(
		handlers.NewBaseHandler(c.Config),
		suggestions,
		c.movieClient,
		c.mediaRenamer,
	)
	err := handler.Handle(ctx)
	if errors.Is(err, handlers.ErrExit) {
		c.Exit()
	}
	return err
}

func (c *Cli) ProcessTvEpisodeSuggestions(ctx context.Context, suggestion mediarenamer.EpisodeSuggestions) (*mediadata.TvShow, error) {
	wrapper := EpisodeSuggestionWrapper{suggestions: suggestion, pattern: c.Config.TvShowPattern, renameFunc: c.RenameTvEpisode}
	var selectedTvShow *mediadata.TvShow
	err := c.processSuggestion(ctx, wrapper, c.Config.TvShowPattern, func() error {
		return c.ProcessTvEpisodeSuggestionsOptions(ctx, suggestion, selectedTvShow)
	})
	return selectedTvShow, err
}

func (c *Cli) ProcessTvEpisodeSuggestionsOptions(ctx context.Context, suggestion mediarenamer.EpisodeSuggestions, newTvShowSuggestion *mediadata.TvShow) error {
	menuBuilder := ui.NewMenuBuilder()

	for _, episode := range suggestion.SuggestedEpisodes {
		episode := episode
		label := fmt.Sprintf("%s - %dx%02d - %s", episode.TvShow.Title, episode.Episode.SeasonNumber, episode.Episode.EpisodeNumber, episode.Episode.Name)
		menuBuilder.AddOption(label, func() error {
			return c.processSuggestion(ctx, EpisodeSuggestionWrapper{suggestions: suggestion, pattern: c.Config.TvShowPattern}, c.Config.TvShowPattern, func() error {
				newTvShowSuggestion = &episode.TvShow
				return c.RenameTvEpisode(ctx, suggestion, episode.TvShow, episode.Episode)
			})
		})
	}

	menuBuilder.AddOption("Search Manually", func() error {
		return c.SearchTvEpisodeSuggestionsManually(context.Background(), suggestion, newTvShowSuggestion)
	})

	menuBuilder.AddOption("Rename Manually", func() error {
		return c.RenameEpisodeFileManually(ctx, suggestion)
	})

	menuBuilder.AddStandardOptions(
		func() error {
			ui.ShowInfo("Skipping renaming of %s", pterm.Yellow(suggestion.Episode.OriginalFilename))
			return nil
		},
		func() error {
			c.Exit()
			return nil
		},
	)

	return menuBuilder.Build()
}

func (c *Cli) SearchTvEpisodeSuggestionsManually(ctx context.Context, suggestions mediarenamer.EpisodeSuggestions, newTvShowSuggestion *mediadata.TvShow) error {
	query, err := pterm.DefaultInteractiveTextInput.WithDefaultValue(fmt.Sprintf("%s", suggestions.Episode.Name)).Show(pterm.Sprintf("Search for '%s'", suggestions.Episode.OriginalFilename))
	if err != nil {
		pterm.Error.Println(pterm.Sprintf("Error getting search query: %v", err))
	}

	tvShows, err := c.tvClient.SearchTvShow(ctx, query, 0, 1)
	if err != nil {
		pterm.Error.Println(pterm.Sprintf("Error searching tvShows: %v", err))
		return err
	}

	suggestions.SuggestedEpisodes = make([]mediarenamer.SuggestedEpisode, 0, len(tvShows.TvShows))
	for _, tvShow := range tvShows.TvShows {
		episodes, err := c.tvClient.GetEpisode(ctx, tvShow.ID, suggestions.Episode.Season, suggestions.Episode.Episode)
		if err != nil {
			pterm.Error.Println(pterm.Sprintf("Error getting episode: %v", err))
		}
		suggestions.SuggestedEpisodes = append(suggestions.SuggestedEpisodes, struct {
			TvShow  mediadata.TvShow
			Episode mediadata.Episode
		}{TvShow: tvShow, Episode: episodes})
	}
	wrapper := EpisodeSuggestionWrapper{suggestions: suggestions, pattern: c.Config.TvShowPattern, renameFunc: c.RenameTvEpisode}
	return c.processSuggestion(ctx, wrapper, c.Config.TvShowPattern, func() error {
		return c.ProcessTvEpisodeSuggestionsOptions(ctx, suggestions, newTvShowSuggestion)
	})
}
