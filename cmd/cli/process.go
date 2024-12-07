package cli

import (
	"context"
	"fmt"

	"github.com/nouuu/gonamer/internal/mediadata"
	"github.com/nouuu/gonamer/internal/mediarenamer"
	"github.com/pterm/pterm"
)

type MediaSuggestion interface {
	GetOriginalFilename() string
	GenerateFilename(pattern string) string
	HasSingleSuggestion() bool
}

// MovieSuggestionWrapper implémente MediaSuggestion pour les films
type MovieSuggestionWrapper struct {
	suggestion mediarenamer.MovieSuggestions
	pattern    string
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

func (m MovieSuggestionWrapper) HasSingleSuggestion() bool {
	return len(m.suggestion.SuggestedMovies) == 1
}

type EpisodeSuggestionWrapper struct {
	suggestion mediarenamer.EpisodeSuggestions
	pattern    string
}

func (e EpisodeSuggestionWrapper) GetOriginalFilename() string {
	return e.suggestion.Episode.OriginalFilename
}

func (e EpisodeSuggestionWrapper) GenerateFilename(pattern string) string {
	if len(e.suggestion.SuggestedEpisodes) == 0 {
		return ""
	}
	return mediarenamer.GenerateEpisodeFilename(
		pattern,
		e.suggestion.SuggestedEpisodes[0].TvShow,
		e.suggestion.SuggestedEpisodes[0].Episode,
		e.suggestion.Episode,
	)
}

func (e EpisodeSuggestionWrapper) HasSingleSuggestion() bool {
	return len(e.suggestion.SuggestedEpisodes) == 1
}

func (c *Cli) processSuggestion(_ context.Context, suggestion MediaSuggestion, pattern string, processFunc func() error) error {
	if !suggestion.HasSingleSuggestion() {
		return processFunc()
	}

	newFilename := suggestion.GenerateFilename(pattern)
	if newFilename == suggestion.GetOriginalFilename() {
		pterm.Success.Println("Original filename is already correct for ", pterm.Yellow(suggestion.GetOriginalFilename()))
		return nil
	}

	if c.Config.QuickMode {
		pterm.Success.Println("Quick - Renaming file ", pterm.Yellow(suggestion.GetOriginalFilename()), "to", pterm.Yellow(newFilename))
		return processFunc()
	}

	return processFunc()
}

func (c *Cli) ProcessMovieSuggestions(ctx context.Context, suggestion mediarenamer.MovieSuggestions) error {
	wrapper := MovieSuggestionWrapper{suggestion: suggestion, pattern: c.Config.MoviePattern}
	return c.processSuggestion(ctx, wrapper, c.Config.MoviePattern, func() error {
		return c.ProcessMovieSuggestionsOptions(ctx, suggestion)
	})
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

	movies, err := c.movieClient.SearchMovie(ctx, query, 0, 1)
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

func (c *Cli) ProcessTvEpisodeSuggestions(ctx context.Context, suggestion mediarenamer.EpisodeSuggestions) error {
	wrapper := EpisodeSuggestionWrapper{suggestion: suggestion, pattern: c.Config.TvShowPattern}
	return c.processSuggestion(ctx, wrapper, c.Config.TvShowPattern, func() error {
		return c.ProcessTvEpisodeSuggestionsOptions(ctx, suggestion)
	})
}

func (c *Cli) ProcessTvEpisodeSuggestionsOptions(ctx context.Context, suggestion mediarenamer.EpisodeSuggestions) error {
	options := make(map[string]func() error)
	optionsArray := make([]string, 0)
	for i, episode := range suggestion.SuggestedEpisodes {
		key := fmt.Sprintf("%d. %s - %dx%02d - %s", i+1, episode.TvShow.Title, episode.Episode.SeasonNumber, episode.Episode.EpisodeNumber, episode.Episode.Name)
		options[key] = func() error {
			return c.RenameTvEpisode(ctx, suggestion, episode.TvShow, episode.Episode)
		}
		optionsArray = append(optionsArray, key)
	}

	optionIndex := len(suggestion.SuggestedEpisodes) + 1

	options[fmt.Sprintf("%d. Skip", optionIndex)] = func() error {
		pterm.Info.Println("Skipping renaming of ", pterm.Yellow(suggestion.Episode.OriginalFilename))
		return nil
	}
	optionsArray = append(optionsArray, fmt.Sprintf("%d. Skip", optionIndex))
	optionIndex++

	options[fmt.Sprintf("%d. Search Manually", optionIndex)] = func() error {
		return c.SearchTvEpisodeSuggestionsManually(context.Background(), suggestion)
	}
	optionsArray = append(optionsArray, fmt.Sprintf("%d. Search Manually", optionIndex))
	optionIndex++

	options[fmt.Sprintf("%d. Rename Manually", optionIndex)] = func() error {
		return c.RenameEpisodeFileManually(ctx, suggestion)
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
		pterm.Error.Println(pterm.Sprintf("Error selecting episode: %v", err))
		return err
	}

	return options[selected]()
}

func (c *Cli) SearchTvEpisodeSuggestionsManually(ctx context.Context, suggestions mediarenamer.EpisodeSuggestions) error {
	query, err := pterm.DefaultInteractiveTextInput.WithDefaultValue(fmt.Sprintf("%s", suggestions.Episode.Name)).Show(pterm.Sprintf("Search for '%s'", suggestions.Episode.OriginalFilename))
	if err != nil {
		pterm.Error.Println(pterm.Sprintf("Error getting search query: %v", err))
	}

	tvShows, err := c.tvClient.SearchTvShow(ctx, query, 0, 1)
	if err != nil {
		pterm.Error.Println(pterm.Sprintf("Error searching tvShows: %v", err))
		return err
	}

	suggestions.SuggestedEpisodes = make([]struct {
		TvShow  mediadata.TvShow
		Episode mediadata.Episode
	}, 0, len(tvShows.TvShows))
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
	return c.ProcessTvEpisodeSuggestionsOptions(ctx, suggestions)
}
