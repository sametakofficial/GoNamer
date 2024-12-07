# Project Structure

- 📄 LICENSE.md
- 📄 Makefile
- 📄 README.md
- 📁 cmd
  - 📁 cli
    - 📄 cli.go
    - 📄 process.go
    - 📄 rename.go
    - 📄 scan.go
    - 📄 suggestion.go
  - 📄 main.go
- 📁 conf
  - 📄 conf.go
- 📄 go.mod
- 📁 internal
  - 📁 cache
    - 📄 cache.go
  - 📁 mediadata
    - 📄 mediadata.go
    - 📁 tmdb
      - 📄 config.go
      - 📄 movie.go
      - 📄 tvshow.go
  - 📁 mediarenamer
    - 📄 mediarenamer.go
    - 📄 renamer.go
  - 📁 mediascanner
    - 📁 filescanner
      - 📄 filescanner.go
      - 📄 sanitazer.go
    - 📄 mediascanner.go
- 📄 main.go
- 📁 pkg
  - 📁 logger
    - 📄 logger.go
- 📁 scripts
  - 📄 scan_project.go

# Files Content

# 📄 LICENSE.md
```md

The MIT License (MIT)

Copyright (c) 2023 Noé Larrieu-Lacoste

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

```


# 📄 Makefile
```

scan-project:
	@go run scripts/scan_project.go . -o project_knowledge.md
```


# 📄 README.md
```md
# GoNamer

```


# 📄 cmd/cli/cli.go
```go
package cli

import (
	"context"
	"os"

	"github.com/nouuu/gonamer/conf"
	"github.com/nouuu/gonamer/internal/mediadata"
	"github.com/nouuu/gonamer/internal/mediarenamer"
	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/nouuu/gonamer/pkg/logger"
	"github.com/pterm/pterm"
)

type Cli struct {
	Config       conf.Config
	scanner      mediascanner.MediaScanner
	mediaRenamer *mediarenamer.MediaRenamer
	tvClient     mediadata.TvShowClient
	movieClient  mediadata.MovieClient
}

func clear() {
	print("\033[H\033[2J")
}

func NewCli(scanner mediascanner.MediaScanner, mediaRenamer *mediarenamer.MediaRenamer, movieClient mediadata.MovieClient, tvClient mediadata.TvShowClient) *Cli {
	return &Cli{
		Config:       conf.LoadConfig(),
		scanner:      scanner,
		mediaRenamer: mediaRenamer,
		movieClient:  movieClient,
		tvClient:     tvClient,
	}
}

func (c *Cli) Run(ctx context.Context) {
	switch c.Config.Type {
	case conf.Movie:
		c.processMovie(ctx)
	case conf.TvShow:
		c.processTvShow(ctx)
	}
}

func (c *Cli) processMovie(ctx context.Context) {
	log := logger.FromContext(ctx)

	movies, err := c.ScanMovies(ctx)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error scanning movies: %v", err))
		log.With("error", err).Error("Error scanning movies")
		return
	}
	suggestions, err := c.FindMoviesSuggestions(ctx, movies)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error finding suggestions: %v", err))
		log.With("error", err).Error("Error finding suggestions")
		return
	}
	err = c.RenameMovies(ctx, suggestions)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error renaming movies: %v", err))
		log.With("error", err).Error("Error renaming movies")
		return
	}
}

func (c *Cli) processTvShow(ctx context.Context) {
	log := logger.FromContext(ctx)

	tvEpisodes, err := c.ScanTvEpisodes(ctx)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error scanning tv shows: %v", err))
		log.With("error", err).Error("Error scanning tv shows")
		return
	}
	suggestions, err := c.FindTvEpisodesSuggestions(ctx, tvEpisodes)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error finding suggestions: %v", err))
		log.With("error", err).Error("Error finding suggestions")
		return
	}
	err = c.RenameTvEpisodes(ctx, suggestions)
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error renaming tv shows: %v", err))
		log.With("error", err).Error("Error renaming tv shows")
		return
	}
}

func (c *Cli) Exit() {
	pterm.Info.Println("Exiting...")
	os.Exit(0)
}

```


# 📄 cmd/cli/process.go
```go
package cli

import (
	"context"
	"fmt"

	"github.com/nouuu/gonamer/internal/mediadata"
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

func (c *Cli) ProcessTvEpisodeSuggestions(ctx context.Context, suggestion mediarenamer.EpisodeSuggestions) error {
	if len(suggestion.SuggestedEpisodes) == 1 {
		if mediarenamer.GenerateEpisodeFilename(c.Config.TvShowPattern, suggestion.SuggestedEpisodes[0].TvShow, suggestion.SuggestedEpisodes[0].Episode, suggestion.Episode) == suggestion.Episode.OriginalFilename {
			pterm.Success.Println("Original filename is already correct for ", pterm.Yellow(suggestion.Episode.OriginalFilename))
			return nil
		} else if c.Config.QuickMode {
			pterm.Success.Println("Quick - Renaming episode ", pterm.Yellow(suggestion.Episode.OriginalFilename), "to", pterm.Yellow(mediarenamer.GenerateEpisodeFilename(c.Config.TvShowPattern, suggestion.SuggestedEpisodes[0].TvShow, suggestion.SuggestedEpisodes[0].Episode, suggestion.Episode)))
			return c.RenameTvEpisode(ctx, suggestion, suggestion.SuggestedEpisodes[0].TvShow, suggestion.SuggestedEpisodes[0].Episode)
		}
	}

	return c.ProcessTvEpisodeSuggestionsOptions(ctx, suggestion)
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

	tvShows, err := c.tvClient.SearchTvShow(query, 0, 1)
	if err != nil {
		pterm.Error.Println(pterm.Sprintf("Error searching tvShows: %v", err))
		return err
	}

	suggestions.SuggestedEpisodes = make([]struct {
		TvShow  mediadata.TvShow
		Episode mediadata.Episode
	}, 0, len(tvShows.TvShows))
	for _, tvShow := range tvShows.TvShows {
		episodes, err := c.tvClient.GetEpisode(tvShow.ID, suggestions.Episode.Season, suggestions.Episode.Episode)
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

```


# 📄 cmd/cli/rename.go
```go
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

func (c *Cli) RenameTvEpisodes(ctx context.Context, suggestions []mediarenamer.EpisodeSuggestions) error {
	slices.SortFunc(suggestions, func(i, j mediarenamer.EpisodeSuggestions) int {
		return strings.Compare(i.Episode.OriginalFilename, j.Episode.OriginalFilename)
	})

	pb, _ := pterm.DefaultProgressbar.WithTotal(len(suggestions)).WithTitle("Renaming TV episodes...").Start()
	for _, suggestion := range suggestions {
		pterm.Print("\n")
		pb.UpdateTitle("Renaming " + suggestion.Episode.OriginalFilename)
		pb.Increment()
		pb, _ = pb.Stop()

		err := c.ProcessTvEpisodeSuggestions(ctx, suggestion)
		if err != nil {
			return err
		}

		pb, _ = pb.Start()
	}
	pb.Stop()
	pterm.Success.Println("Finished renaming TV episodes")
	return nil
}

func (c *Cli) RenameTvEpisode(ctx context.Context, suggestion mediarenamer.EpisodeSuggestions, tvShow mediadata.TvShow, episode mediadata.Episode) error {
	pterm.Info.Println("Renaming episode ", pterm.Yellow(suggestion.Episode.OriginalFilename), "to", pterm.Yellow(mediarenamer.GenerateEpisodeFilename(c.Config.TvShowPattern, tvShow, episode, suggestion.Episode)))
	err := c.mediaRenamer.RenameEpisode(ctx, suggestion.Episode, tvShow, episode, c.Config.TvShowPattern, c.Config.DryRun)
	if err != nil {
		pterm.Error.Println(pterm.Sprintf("Error renaming episode: %v", err))
		return err
	}
	return nil
}

func (c *Cli) RenameEpisodeFileManually(ctx context.Context, suggestion mediarenamer.EpisodeSuggestions) error {
	pterm.Info.Println("Renaming manually for ", pterm.Yellow(suggestion.Episode.OriginalFilename))
	result, _ := pterm.DefaultInteractiveTextInput.
		WithDefaultText("Enter new filename (without extension) :").
		WithDefaultValue(fmt.Sprintf("%s - %dx%02d", suggestion.Episode.Name, suggestion.Episode.Season, suggestion.Episode.Episode)).
		Show()

	filename := fmt.Sprintf("%s.%s", result, suggestion.Episode.Extension)

	pterm.Info.Println("Renaming episode ", pterm.Yellow(suggestion.Episode.OriginalFilename), "to", pterm.Yellow(filename))

	return c.mediaRenamer.RenameFile(
		ctx,
		suggestion.Episode.FullPath,
		filepath.Join(filepath.Dir(suggestion.Episode.FullPath), filename),
		c.Config.DryRun,
	)
}

```


# 📄 cmd/cli/scan.go
```go
package cli

import (
	"context"

	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/pterm/pterm"
)

func (c *Cli) ScanMovies(ctx context.Context) ([]mediascanner.Movie, error) {
	pterm.Info.Printfln("Scanning movies in '%s'...", c.Config.MediaPath)
	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start("Scanning movies...")
	defer spinner.Stop()
	movies, err := c.scanner.ScanMovies(ctx, c.Config.MediaPath, mediascanner.ScanMoviesOptions{Recursively: c.Config.Recursive})
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error scanning movies: %v", err))
		return nil, err
	}
	spinner.Success(pterm.Sprintf("Found %d movies", len(movies)))
	return movies, nil
}

func (c *Cli) ScanTvEpisodes(ctx context.Context) ([]mediascanner.Episode, error) {
	pterm.Info.Printfln("Scanning TV episodes in '%s'...", c.Config.MediaPath)
	spinner, _ := pterm.DefaultSpinner.WithShowTimer(true).Start("Scanning TV shows...")
	defer spinner.Stop()
	tvShows, err := c.scanner.ScanEpisodes(ctx, c.Config.MediaPath, mediascanner.ScanEpisodesOptions{Recursively: c.Config.Recursive})
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error scanning TV shows: %v", err))
		return nil, err
	}
	spinner.Success(pterm.Sprintf("Found %d TV shows", len(tvShows)))
	return tvShows, nil
}

```


# 📄 cmd/cli/suggestion.go
```go
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

```


# 📄 cmd/main.go
```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nouuu/gonamer/cmd/cli"
	"github.com/nouuu/gonamer/conf"
	"github.com/nouuu/gonamer/internal/mediadata/tmdb"
	"github.com/nouuu/gonamer/internal/mediarenamer"
	"github.com/nouuu/gonamer/internal/mediascanner/filescanner"
	"github.com/nouuu/gonamer/pkg/logger"
	"github.com/pterm/pterm"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctx := context.Background()

	initLogger()

	startCli(ctx)
}

func initLogger() {
	logger.SetLoggerLevel(zapcore.InfoLevel)
	logfile, err := os.OpenFile("mediatracker.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening log file: %v\n", err)
		os.Exit(1)
	}

	logger.SetLoggerOutput(zapcore.WriteSyncer(logfile))
}

func startCli(ctx context.Context) {
	log := logger.FromContext(ctx)

	pterm.DefaultHeader.Println("Media Renamer")
	pterm.Print("\n\n")

	pterm.Info.Printfln("Loading configuration...\n")

	config := conf.LoadConfig()

	scanner := filescanner.New()
	movieClient, err := tmdb.NewMovieClient(config.TMDBAPIKey, tmdb.WithLang("fr-FR"))
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error creating movie client: %v", err))
		log.Fatalf("Error creating movie client: %v", err)
	}
	tvShowClient, err := tmdb.NewTvShowClient(config.TMDBAPIKey, tmdb.WithLang("fr-FR"))
	if err != nil {
		pterm.Error.Println(pterm.Error.Sprint("Error creating tv show client: %v", err))
		log.Fatalf("Error creating tv show client: %v", err)
	}

	mediaRenamer := mediarenamer.NewMediaRenamer(movieClient, tvShowClient)

	newCli := cli.NewCli(scanner, mediaRenamer, movieClient, tvShowClient)

	newCli.Run(ctx)
}

```


# 📄 conf/conf.go
```go
package conf

import (
	"fmt"
	"os"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type MediaType string

const (
	Movie  MediaType = "movie"
	TvShow           = "tvshow"
)

type Config struct {
	TMDBAPIKey      string    `env:"TMDB_API_KEY" envDefault:"not-set"`
	MediaPath       string    `env:"MEDIA_PATH" envDefault:"./"`
	Recursive       bool      `env:"RECURSIVE" envDefault:"true"`
	IncludeNotFound bool      `env:"INCLUDE_NOT_FOUND" envDefault:"false"`
	DryRun          bool      `env:"DRY_RUN" envDefault:"true"`
	MoviePattern    string    `env:"MOVIE_PATTERN" envDefault:"{name} - {year}{extension}"`
	TvShowPattern   string    `env:"TVSHOW_PATTERN" envDefault:"{name} - {season}x{episode}{extension}"`
	Type            MediaType `env:"TYPE" envDefault:"movie"`
	MaxResults      int       `env:"MAX_RESULTS" envDefault:"5"`
	QuickMode       bool      `env:"QUICK_MODE" envDefault:"false"`
}

func LoadConfig() Config {
	_ = godotenv.Load()
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return cfg
}

```


# 📄 go.mod
```mod
module github.com/nouuu/gonamer

go 1.22

require (
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/cyruzin/golang-tmdb v1.6.5
	github.com/dgraph-io/ristretto v0.2.0
	github.com/eko/gocache/lib/v4 v4.1.6
	github.com/eko/gocache/store/ristretto/v4 v4.2.2
	github.com/joho/godotenv v1.5.1
	github.com/pterm/pterm v0.12.79
	go.uber.org/zap v1.27.0
	golang.org/x/text v0.17.0
)

require (
	atomicgo.dev/cursor v0.2.0 // indirect
	atomicgo.dev/keyboard v0.2.9 // indirect
	atomicgo.dev/schedule v0.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/console v1.0.4 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gookit/color v1.5.4 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lithammer/fuzzysearch v1.1.8 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/term v0.23.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

```


# 📄 internal/cache/cache.go
```go
package cache

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/lib/v4/store"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	ristrettoStore "github.com/eko/gocache/store/ristretto/v4"
)

type Cache interface {
	Set(ctx context.Context, key string, value any) error
	Get(ctx context.Context, key string) (any, error)
}

func NewGoCache() (Cache, error) {
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1_000_000,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	ristrettoCacheStore := ristrettoStore.NewRistretto(ristrettoCache, store.WithExpiration(1*time.Hour))

	return &goCache{
		marshaler: marshaler.New(cache.New[any](ristrettoCacheStore)),
	}, nil
}

type goCache struct {
	marshaler *marshaler.Marshaler
}

func (g goCache) Set(ctx context.Context, key string, value any) error {
	//TODO implement me
	panic("implement me")
}

func (g goCache) Get(ctx context.Context, key string) (any, error) {
	//TODO implement me
	panic("implement me")
}

```


# 📄 internal/mediadata/mediadata.go
```go
package mediadata

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

type Status string

const (
	StatusReturning Status = "Returning Series"
	StatusEnded     Status = "Ended"
)

type Genre struct {
	ID   string
	Name string
}

type Person struct {
	ID         string
	Name       string
	Character  string
	ProfileURL string
}

type Studio struct {
	ID   string
	Name string
}

type Movie struct {
	ID          string
	Title       string
	Overview    string
	ReleaseDate string
	Year        string
	PosterURL   string
	Rating      float32
	RatingCount int64
}

type MovieDetails struct {
	Movie
	Runtime int
	Genres  []Genre
	Cast    []Person
	Studio  []Studio
}

type MovieResults struct {
	Movies         []Movie
	Totals         int64
	ResultsPerPage int64
}

type Season struct {
	SeasonNumber int
	EpisodeCount int
	AirDate      string
	PosterURL    string
}

type Episode struct {
	ID            string
	AirDate       string
	EpisodeNumber int
	SeasonNumber  int
	Name          string
	Overview      string
	StillURL      string
	VoteAverage   float32
	VoteCount     int64
}

type TvShow struct {
	ID          string
	Title       string
	Overview    string
	FistAirDate string
	Year        string
	PosterURL   string
	Rating      float32
	RatingCount int64
}

type TvShowDetails struct {
	TvShow
	SeasonCount  int
	EpisodeCount int
	LastEpisode  Episode
	NextEpisode  Episode
	Status       Status
	Seasons      []Season
	Genres       []Genre
	Cast         []Person
	Studio       []Studio
}

type TvShowResults struct {
	TvShows        []TvShow
	Totals         int64
	ResultsPerPage int64
}

type MovieClient interface {
	SearchMovie(query string, year int, page int) (MovieResults, error)
	GetMovie(id string) (Movie, error)
	GetMovieDetails(id string) (MovieDetails, error)
}

type TvShowClient interface {
	SearchTvShow(query string, year int, page int) (TvShowResults, error)
	GetTvShow(id string) (TvShow, error)
	GetTvShowDetails(id string) (TvShowDetails, error)
	GetEpisode(id string, seasonNumber int, episodeNumber int) (Episode, error)
}

func ShowMovieResults(movies MovieResults) {
	slog.Info("Movies")
	for _, movie := range movies.Movies {
		mJson, err := marshalMovie(movie)
		if err != nil {
			slog.Error("Failed to marshal movie", slog.Any("error", err))
			os.Exit(1)
		}
		fmt.Println(string(mJson))
	}
}

func ShowTvShowResults(tvShows TvShowResults) {
	slog.Info("TvShows")
	for _, tvShow := range tvShows.TvShows {
		mJson, err := marshalTvShow(tvShow)
		if err != nil {
			slog.Error("Failed to marshal tv show", slog.Any("error", err))
			os.Exit(1)
		}
		fmt.Println(string(mJson))
	}
}

func marshalMovie(movie Movie) (string, error) {
	mJson, err := json.MarshalIndent(movie, "", "  ")
	if err != nil {
		return "", err
	}
	return string(mJson), nil
}

func marshalMovieDetails(movieDetails MovieDetails) (string, error) {
	mJson, err := json.MarshalIndent(movieDetails, "", "  ")
	if err != nil {
		return "", err
	}
	return string(mJson), nil
}

func marshalTvShow(tvShow TvShow) (string, error) {
	mJson, err := json.MarshalIndent(tvShow, "", "  ")
	if err != nil {
		return "", err
	}
	return string(mJson), nil
}

func marshalTvShowDetails(tvShowDetails TvShowDetails) (string, error) {
	mJson, err := json.MarshalIndent(tvShowDetails, "", "  ")
	if err != nil {
		return "", err
	}
	return string(mJson), nil
}

```


# 📄 internal/mediadata/tmdb/config.go
```go
package tmdb

import (
	"strconv"

	"github.com/cyruzin/golang-tmdb"
	"github.com/nouuu/gonamer/internal/mediadata"
)

const tmdbImageBaseUrl = "https://image.tmdb.org/t/p/original"

type OptFunc func(opts *Opts)

type AllOpts struct {
	APIKey string
	Opts
}

type Opts struct {
	Lang  string
	Adult bool
}

func WithLang(lang string) OptFunc {
	return func(opts *Opts) {
		opts.Lang = lang
	}
}

func WithAdult(adult bool) OptFunc {
	return func(opts *Opts) {
		opts.Adult = adult
	}
}

func defaultOpts(apiKey string) AllOpts {
	return AllOpts{
		APIKey: apiKey,
		Opts: Opts{
			Lang:  "en-US",
			Adult: false,
		},
	}
}

type tmdbClient struct {
	client *tmdb.Client
	opts   AllOpts
}

func cfgMap(opts AllOpts, args ...map[string]string) map[string]string {
	cfg := map[string]string{
		"language":      opts.Lang,
		"include_adult": strconv.FormatBool(opts.Adult),
	}

	for _, arg := range args {
		for k, v := range arg {
			cfg[k] = v
		}
	}
	return cfg
}

func buildGenres(genres []struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}) []mediadata.Genre {
	var g = make([]mediadata.Genre, len(genres))
	for i, genre := range genres {
		g[i] = mediadata.Genre{
			ID:   strconv.FormatInt(genre.ID, 10),
			Name: genre.Name,
		}
	}
	return g
}

func buildStudio(studios []struct {
	Name          string `json:"name"`
	ID            int64  `json:"id"`
	LogoPath      string `json:"logo_path"`
	OriginCountry string `json:"origin_country"`
}) []mediadata.Studio {
	var s = make([]mediadata.Studio, len(studios))
	for i, studio := range studios {
		s[i] = mediadata.Studio{
			ID:   strconv.FormatInt(studio.ID, 10),
			Name: studio.Name,
		}
	}
	return s
}

```


# 📄 internal/mediadata/tmdb/movie.go
```go
package tmdb

import (
	"strconv"

	"github.com/cyruzin/golang-tmdb"
	"github.com/nouuu/gonamer/internal/mediadata"
)

func NewMovieClient(APIKey string, opts ...OptFunc) (mediadata.MovieClient, error) {
	o := defaultOpts(APIKey)
	for _, optF := range opts {
		optF(&o.Opts)
	}

	client, err := tmdb.Init(o.APIKey)
	if err != nil {
		return nil, err
	}
	return &tmdbClient{client: client, opts: o}, nil
}

func (t *tmdbClient) SearchMovie(query string, year int, page int) (mediadata.MovieResults, error) {
	opts := map[string]string{
		"page": strconv.Itoa(page),
	}
	if year != 0 {
		opts["year"] = strconv.Itoa(year)
	}
	searchMovies, err := t.client.GetSearchMovies(query, cfgMap(t.opts, opts))
	if err != nil {
		return mediadata.MovieResults{}, err
	}
	movies := buildMovieFromResult(searchMovies.SearchMoviesResults)
	return mediadata.MovieResults{
		Movies:         movies,
		Totals:         searchMovies.TotalResults,
		ResultsPerPage: 20,
	}, nil
}

func (t *tmdbClient) GetMovie(id string) (mediadata.Movie, error) {
	var idInt int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return mediadata.Movie{}, err
	}
	movieDetails, err := t.client.GetMovieDetails(
		idInt,
		cfgMap(t.opts),
	)
	if err != nil {
		return mediadata.Movie{}, err
	}
	return buildMovie(movieDetails), nil
}

func (t *tmdbClient) GetMovieDetails(id string) (mediadata.MovieDetails, error) {
	var idInt int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return mediadata.MovieDetails{}, err
	}
	movieDetails, err := t.client.GetMovieDetails(
		idInt,
		cfgMap(t.opts, map[string]string{
			"append_to_response": "credits",
		}),
	)
	if err != nil {
		return mediadata.MovieDetails{}, err
	}
	return buildMovieDetails(movieDetails), nil
}

func buildMovie(movie *tmdb.MovieDetails) mediadata.Movie {
	releaseYear := ""
	if len(movie.ReleaseDate) >= 4 {
		releaseYear = movie.ReleaseDate[:4]
	}
	return mediadata.Movie{
		ID:          strconv.FormatInt(movie.ID, 10),
		Title:       movie.Title,
		Overview:    movie.Overview,
		ReleaseDate: movie.ReleaseDate,
		Year:        releaseYear,
		PosterURL:   tmdbImageBaseUrl + movie.PosterPath,
		Rating:      movie.VoteAverage,
		RatingCount: movie.VoteCount,
	}
}

func buildMovieDetails(details *tmdb.MovieDetails) mediadata.MovieDetails {
	releaseYear := ""
	if len(details.ReleaseDate) >= 4 {
		releaseYear = details.ReleaseDate[:4]

	}
	return mediadata.MovieDetails{
		Movie: mediadata.Movie{
			ID:          strconv.FormatInt(details.ID, 10),
			Title:       details.Title,
			Overview:    details.Overview,
			ReleaseDate: details.ReleaseDate,
			Year:        releaseYear,
			PosterURL:   tmdbImageBaseUrl + details.PosterPath,
			Rating:      details.VoteAverage,
			RatingCount: details.VoteCount,
		},
		Runtime: details.Runtime,
		Genres:  buildGenres(details.Genres),
		Cast:    buildMovieCast(details.Credits.Cast),
		Studio:  buildStudio(details.ProductionCompanies),
	}
}

func buildMovieFromResult(result *tmdb.SearchMoviesResults) []mediadata.Movie {
	var movies = make([]mediadata.Movie, len(result.Results))
	for i, movie := range result.Results {
		movies[i] = buildMovie(&tmdb.MovieDetails{
			ID:          movie.ID,
			Title:       movie.Title,
			Overview:    movie.Overview,
			ReleaseDate: movie.ReleaseDate,
			PosterPath:  movie.PosterPath,
			VoteAverage: movie.VoteAverage,
			VoteCount:   movie.VoteCount,
		})
	}
	return movies
}

func buildMovieCast(cast []struct {
	Adult              bool    `json:"adult"`
	CastID             int64   `json:"cast_id"`
	Character          string  `json:"character"`
	CreditID           string  `json:"credit_id"`
	Gender             int     `json:"gender"`
	ID                 int64   `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	Order              int     `json:"order"`
	OriginalName       string  `json:"original_name"`
	Popularity         float32 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
}) []mediadata.Person {
	var c = make([]mediadata.Person, len(cast))
	for i, person := range cast {
		c[i] = mediadata.Person{
			ID:         strconv.FormatInt(person.ID, 10),
			Name:       person.Name,
			Character:  person.Character,
			ProfileURL: tmdbImageBaseUrl + person.ProfilePath,
		}
	}
	return c
}

```


# 📄 internal/mediadata/tmdb/tvshow.go
```go
package tmdb

import (
	"strconv"

	"github.com/cyruzin/golang-tmdb"
	"github.com/nouuu/gonamer/internal/mediadata"
)

func NewTvShowClient(APIKey string, opts ...OptFunc) (mediadata.TvShowClient, error) {
	o := defaultOpts(APIKey)
	for _, optF := range opts {
		optF(&o.Opts)
	}

	client, err := tmdb.Init(o.APIKey)
	if err != nil {
		return nil, err
	}
	return &tmdbClient{client: client, opts: o}, nil
}

func (t *tmdbClient) SearchTvShow(query string, year int, page int) (mediadata.TvShowResults, error) {
	opts := map[string]string{
		"page": strconv.Itoa(page),
	}
	if year != 0 {
		opts["year"] = strconv.Itoa(year)
	}

	searchTvShows, err := t.client.GetSearchTVShow(query, cfgMap(t.opts, opts))
	if err != nil {
		return mediadata.TvShowResults{}, err
	}
	tvShows := buildTvShowFromResult(searchTvShows.SearchTVShowsResults)
	return mediadata.TvShowResults{
		TvShows:        tvShows,
		Totals:         searchTvShows.TotalResults,
		ResultsPerPage: 20,
	}, nil
}

func (t *tmdbClient) GetTvShow(id string) (mediadata.TvShow, error) {
	var idInt int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return mediadata.TvShow{}, err
	}
	tvShowDetails, err := t.client.GetTVDetails(
		idInt,
		cfgMap(t.opts),
	)
	if err != nil {
		return mediadata.TvShow{}, err
	}
	return buildTvShow(tvShowDetails), nil
}

func (t *tmdbClient) GetTvShowDetails(id string) (mediadata.TvShowDetails, error) {
	var idInt int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return mediadata.TvShowDetails{}, err
	}
	tvShowDetails, err := t.client.GetTVDetails(
		idInt,
		cfgMap(t.opts, map[string]string{
			"append_to_response": "credits",
		}),
	)
	if err != nil {
		return mediadata.TvShowDetails{}, err
	}
	return buildTvShowDetails(tvShowDetails), nil
}

func (t *tmdbClient) GetEpisode(id string, seasonNumber int, episodeNumber int) (mediadata.Episode, error) {
	var idInt int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return mediadata.Episode{}, err
	}
	episode, err := t.client.GetTVEpisodeDetails(
		idInt,
		seasonNumber,
		episodeNumber,
		cfgMap(t.opts),
	)
	if err != nil {
		return mediadata.Episode{}, err
	}
	return buildEpisode(struct {
		AirDate        string  `json:"air_date"`
		EpisodeNumber  int     `json:"episode_number"`
		ID             int64   `json:"id"`
		Name           string  `json:"name"`
		Overview       string  `json:"overview"`
		ProductionCode string  `json:"production_code"`
		SeasonNumber   int     `json:"season_number"`
		ShowID         int64   `json:"show_id"`
		StillPath      string  `json:"still_path"`
		VoteAverage    float32 `json:"vote_average"`
		VoteCount      int64   `json:"vote_count"`
	}{
		AirDate:        episode.AirDate,
		EpisodeNumber:  episode.EpisodeNumber,
		ID:             episode.ID,
		Name:           episode.Name,
		Overview:       episode.Overview,
		ProductionCode: episode.ProductionCode,
		SeasonNumber:   episode.SeasonNumber,
		ShowID:         int64(idInt),
		StillPath:      episode.StillPath,
		VoteAverage:    episode.VoteAverage,
		VoteCount:      episode.VoteCount,
	}), nil
}

func buildTvShow(tvShow *tmdb.TVDetails) mediadata.TvShow {
	releaseYear := ""
	if len(tvShow.FirstAirDate) >= 4 {
		releaseYear = tvShow.FirstAirDate[:4]
	}
	return mediadata.TvShow{
		ID:          strconv.FormatInt(tvShow.ID, 10),
		Title:       tvShow.Name,
		Overview:    tvShow.Overview,
		FistAirDate: tvShow.FirstAirDate,
		Year:        releaseYear,
		PosterURL:   tmdbImageBaseUrl + tvShow.PosterPath,
		Rating:      tvShow.VoteAverage,
		RatingCount: tvShow.VoteCount,
	}
}

func buildTvShowDetails(details *tmdb.TVDetails) mediadata.TvShowDetails {
	releaseYear := ""
	if len(details.FirstAirDate) >= 4 {
		releaseYear = details.FirstAirDate[:4]
	}
	return mediadata.TvShowDetails{
		TvShow: mediadata.TvShow{
			ID:          strconv.FormatInt(details.ID, 10),
			Title:       details.Name,
			Overview:    details.Overview,
			FistAirDate: details.FirstAirDate,
			Year:        releaseYear,
			PosterURL:   tmdbImageBaseUrl + details.PosterPath,
			Rating:      details.VoteAverage,
			RatingCount: details.VoteCount,
		},
		Status:       mediadata.Status(details.Status),
		EpisodeCount: details.NumberOfEpisodes,
		SeasonCount:  details.NumberOfSeasons,
		Seasons:      buildSeasons(details.Seasons),
		LastEpisode:  buildEpisode(details.LastEpisodeToAir),
		NextEpisode:  buildEpisode(details.NextEpisodeToAir),
		Cast:         buildTvShowCast(details.Credits.Cast),
		Genres:       buildGenres(details.Genres),
		Studio:       buildStudio(details.Networks),
	}
}

func buildSeasons(seasons []struct {
	AirDate      string  `json:"air_date"`
	EpisodeCount int     `json:"episode_count"`
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	SeasonNumber int     `json:"season_number"`
	VoteAverage  float32 `json:"vote_average"`
}) []mediadata.Season {
	var s = make([]mediadata.Season, len(seasons))
	for i, season := range seasons {
		s[i] = mediadata.Season{
			SeasonNumber: season.SeasonNumber,
			EpisodeCount: season.EpisodeCount,
			AirDate:      season.AirDate,
			PosterURL:    tmdbImageBaseUrl + season.PosterPath,
		}
	}
	return s
}

func buildEpisode(episode struct {
	AirDate        string  `json:"air_date"`
	EpisodeNumber  int     `json:"episode_number"`
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Overview       string  `json:"overview"`
	ProductionCode string  `json:"production_code"`
	SeasonNumber   int     `json:"season_number"`
	ShowID         int64   `json:"show_id"`
	StillPath      string  `json:"still_path"`
	VoteAverage    float32 `json:"vote_average"`
	VoteCount      int64   `json:"vote_count"`
}) mediadata.Episode {
	return mediadata.Episode{
		ID:            strconv.FormatInt(episode.ID, 10),
		AirDate:       episode.AirDate,
		EpisodeNumber: episode.EpisodeNumber,
		SeasonNumber:  episode.SeasonNumber,
		Name:          episode.Name,
		Overview:      episode.Overview,
		StillURL:      tmdbImageBaseUrl + episode.StillPath,
		VoteAverage:   episode.VoteAverage,
		VoteCount:     episode.VoteCount,
	}
}

func buildTvShowFromResult(result *tmdb.SearchTVShowsResults) []mediadata.TvShow {
	var tvShows = make([]mediadata.TvShow, len(result.Results))
	for i, tvShow := range result.Results {
		tvShows[i] = buildTvShow(&tmdb.TVDetails{
			ID:           tvShow.ID,
			Name:         tvShow.Name,
			Overview:     tvShow.Overview,
			FirstAirDate: tvShow.FirstAirDate,
			PosterPath:   tvShow.PosterPath,
			VoteAverage:  tvShow.VoteAverage,
			VoteCount:    tvShow.VoteCount,
		})
	}
	return tvShows
}

func buildTvShowCast(cast []struct {
	Character          string  `json:"character"`
	CreditID           string  `json:"credit_id"`
	Gender             int     `json:"gender"`
	ID                 int64   `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	Order              int     `json:"order"`
	OriginalName       string  `json:"original_name"`
	Popularity         float32 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
}) []mediadata.Person {
	var c = make([]mediadata.Person, len(cast))
	for i, person := range cast {
		c[i] = mediadata.Person{
			ID:         strconv.FormatInt(person.ID, 10),
			Name:       person.Name,
			Character:  person.Character,
			ProfileURL: tmdbImageBaseUrl + person.ProfilePath,
		}
	}
	return c
}

```


# 📄 internal/mediarenamer/mediarenamer.go
```go
package mediarenamer

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nouuu/gonamer/internal/mediadata"
	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/nouuu/gonamer/pkg/logger"
	"go.uber.org/zap"
)

type MediaRenamer struct {
	movieClient  mediadata.MovieClient
	tvShowClient mediadata.TvShowClient
}

type MovieSuggestions struct {
	Movie           mediascanner.Movie
	SuggestedMovies []mediadata.Movie
}

type EpisodeSuggestions struct {
	Episode           mediascanner.Episode
	SuggestedEpisodes []struct {
		TvShow  mediadata.TvShow
		Episode mediadata.Episode
	}
}

type FindMovieSuggestionCallback func(suggestion MovieSuggestions, err error)

type FindEpisodeSuggestionCallback func(suggestion EpisodeSuggestions, err error)

func NewMediaRenamer(movieClient mediadata.MovieClient, tvShowClient mediadata.TvShowClient) *MediaRenamer {
	return &MediaRenamer{movieClient: movieClient, tvShowClient: tvShowClient}
}

func (mr *MediaRenamer) FindMovieSuggestions(ctx context.Context, movies []mediascanner.Movie, maxResults int, callback ...FindMovieSuggestionCallback) []MovieSuggestions {
	log := logger.FromContext(ctx)
	start := time.Now()
	log.Infof("Getting suggestions for %d movies", len(movies))

	suggestions := mr.getMoviesSuggestions(ctx, movies, maxResults, log, callback...)

	log.Infof("Finished getting suggestions for %d movies in %s", len(movies), time.Since(start))
	return suggestions
}

func (mr *MediaRenamer) FindEpisodeSuggestions(ctx context.Context, episodes []mediascanner.Episode, maxResults int, callback ...FindEpisodeSuggestionCallback) []EpisodeSuggestions {
	log := logger.FromContext(ctx)
	start := time.Now()
	log.Infof("Getting suggestions for %d episodes", len(episodes))

	suggestions := mr.getEpisodesSuggestions(ctx, episodes, maxResults, log, callback...)

	log.Infof("Finished getting suggestions for %d episodes in %s", len(episodes), time.Since(start))
	return suggestions
}

func (mr *MediaRenamer) RenameMovie(ctx context.Context, fileMovie mediascanner.Movie, mediadataMovie mediadata.Movie, pattern string, dryrun bool) error {
	// "{name} - {year}{extension}" <3
	filename := GenerateMovieFilename(pattern, mediadataMovie, fileMovie)

	return mr.RenameFile(ctx, fileMovie.FullPath, filepath.Join(filepath.Dir(fileMovie.FullPath), filename), dryrun)
}

func (mr *MediaRenamer) RenameEpisode(ctx context.Context, fileEpisode mediascanner.Episode, tvShow mediadata.TvShow, episode mediadata.Episode, pattern string, dryrun bool) error {
	// "{name} - {season}x{episode} - {episode_title}{extension}" <3
	filename := GenerateEpisodeFilename(pattern, tvShow, episode, fileEpisode)

	return mr.RenameFile(ctx, fileEpisode.FullPath, filepath.Join(filepath.Dir(fileEpisode.FullPath), filename), dryrun)
}

func (mr *MediaRenamer) RenameFile(ctx context.Context, source, destination string, dryrun bool) error {
	log := logger.FromContext(ctx)
	log.Infof("Renaming file %s -> %s", source, destination)
	if dryrun {
		return nil
	}
	err := os.Rename(source, destination)
	if err != nil {
		log.With("error", err).Error("Error renaming file")
		return err
	}
	return nil
}

func (mr *MediaRenamer) SuggestMovies(ctx context.Context, movie mediascanner.Movie, maxResults int) (suggestions MovieSuggestions, err error) {
	log := logger.FromContext(ctx).With("movie", movie)
	suggestions.Movie = movie
	maxResults = int(math.Max(math.Min(float64(maxResults), 100), 1))
	movies, err := mr.movieClient.SearchMovie(movie.Name, movie.Year, 1)
	if err != nil {
		log.With("error", err).Error("Error searching movie")
		return
	}
	if movies.Totals == 0 {
		log.Warnf("No movie found for %s", movie.Name)
		err = errors.New("no movie found")
		return
	}
	suggestions.SuggestedMovies = movies.Movies
	if len(suggestions.SuggestedMovies) > 5 {
		suggestions.SuggestedMovies = suggestions.SuggestedMovies[:5]
	}

	return
}

func (mr *MediaRenamer) SuggestEpisodes(ctx context.Context, episode mediascanner.Episode, maxResults int) (suggestions EpisodeSuggestions, err error) {
	log := logger.FromContext(ctx).With("episode", episode)
	suggestions.Episode = episode
	maxResults = int(math.Max(math.Min(float64(maxResults), 100), 1))
	tvShow, err := mr.tvShowClient.SearchTvShow(episode.Name, 0, 1)
	if err != nil {
		log.With("error", err).Error("Error searching tv show")
		return
	}
	if tvShow.Totals == 0 {
		log.Warnf("No tv show found for %s", episode.Name)
		err = errors.New("no tv show found")
		return
	}
	for _, tvShow := range tvShow.TvShows { // TODO implement GetSeasonEpisodes and put answer in cache to avoid multiple requests
		episodes, err := mr.tvShowClient.GetEpisode(tvShow.ID, episode.Season, episode.Episode)
		if err != nil {
			log.With("error", err).Error("Error getting episode")
		}
		suggestions.SuggestedEpisodes = append(suggestions.SuggestedEpisodes, struct {
			TvShow  mediadata.TvShow
			Episode mediadata.Episode
		}{TvShow: tvShow, Episode: episodes})
	}
	return
}

func (mr *MediaRenamer) getMoviesSuggestions(ctx context.Context, movies []mediascanner.Movie, maxResults int, log *zap.SugaredLogger, callback ...FindMovieSuggestionCallback) (movieSuggestion []MovieSuggestions) {
	var wg sync.WaitGroup
	suggestionsCh := make(chan MovieSuggestions, len(movies))
	semaphore := make(chan struct{}, 5) // Limit to 5 concurrent threads

	for _, movie := range movies {
		wg.Add(1)
		go func(movie mediascanner.Movie) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			suggestions, err := mr.getMovieSuggestions(ctx, movie, maxResults)
			for _, cb := range callback {
				cb(suggestions, err)
			}
			if err != nil {
				log.With("error", err).Error("Error getting movie suggestions")
				return
			}
			suggestionsCh <- suggestions
		}(movie)
	}

	wg.Wait()
	close(suggestionsCh)
	for suggestion := range suggestionsCh {
		movieSuggestion = append(movieSuggestion, suggestion)
	}
	return
}

func (mr *MediaRenamer) getEpisodesSuggestions(ctx context.Context, episodes []mediascanner.Episode, maxResults int, log *zap.SugaredLogger, callback ...FindEpisodeSuggestionCallback) (episodeSuggestions []EpisodeSuggestions) {
	var wg sync.WaitGroup
	suggestionsCh := make(chan EpisodeSuggestions, len(episodes))
	semaphore := make(chan struct{}, 5) // Limit to 5 concurrent threads

	for _, episode := range episodes {
		wg.Add(1)
		go func(episode mediascanner.Episode) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			suggestions, err := mr.getEpisodeSuggestions(ctx, episode, maxResults)
			for _, cb := range callback {
				cb(suggestions, err)
			}
			if err != nil {
				log.With("error", err).Error("Error getting episode suggestions")
				return
			}
			suggestionsCh <- suggestions
		}(episode)
	}

	wg.Wait()
	close(suggestionsCh)
	for suggestion := range suggestionsCh {
		episodeSuggestions = append(episodeSuggestions, suggestion)
	}
	return
}

func (mr *MediaRenamer) getMovieSuggestions(ctx context.Context, movie mediascanner.Movie, maxResults int) (suggestions MovieSuggestions, err error) {
	log := logger.FromContext(ctx).With("movie", movie)

	suggestions, err = mr.SuggestMovies(ctx, movie, maxResults)
	if err != nil {
		log.With("movie", movie).With("error", err).Error("Error suggesting movie")
		return
	}
	output := fmt.Sprintf("Suggested movie '%s (%d)' -> '%s (%s)'", suggestions.Movie.Name, suggestions.Movie.Year, suggestions.SuggestedMovies[0].Title, suggestions.SuggestedMovies[0].Year)
	log.With("suggestions", len(suggestions.SuggestedMovies)).Debug(output)
	return
}

func (mr *MediaRenamer) getEpisodeSuggestions(ctx context.Context, episode mediascanner.Episode, maxResults int) (suggestions EpisodeSuggestions, err error) {
	log := logger.FromContext(ctx).With("episode", episode)

	suggestions, err = mr.SuggestEpisodes(ctx, episode, maxResults)
	if err != nil {
		log.With("episode", episode).With("error", err).Error("Error suggesting episode")
		return
	}
	output := fmt.Sprintf("Suggested episode '%s (%d)' -> '%s %dx%02d - %s'", suggestions.Episode.Name, suggestions.Episode.Season, suggestions.SuggestedEpisodes[0].TvShow.Title, suggestions.SuggestedEpisodes[0].Episode.SeasonNumber, suggestions.SuggestedEpisodes[0].Episode.EpisodeNumber, suggestions.SuggestedEpisodes[0].Episode.Name)
	log.With("suggestions", len(suggestions.SuggestedEpisodes)).Debug(output)
	return
}

```


# 📄 internal/mediarenamer/renamer.go
```go
package mediarenamer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nouuu/gonamer/internal/mediadata"
	"github.com/nouuu/gonamer/internal/mediascanner"
)

type Field string

const (
	FieldName         Field = "{name}"
	FieldYear         Field = "{year}"
	FieldDate         Field = "{date}"
	FieldExt          Field = "{extension}"
	FieldSeason       Field = "{season}"
	FieldEpisode      Field = "{episode}"
	FieldEpisodeTitle Field = "{episode_title}"
)

func GenerateMovieFilename(pattern string, movie mediadata.Movie, fileMovie mediascanner.Movie) string {
	//return fmt.Sprintf("%s - %s%s", movie.Title, movie.Year, fileMovie.Extension)
	filename := pattern
	filename = replaceField(filename, FieldName, movie.Title)
	filename = replaceField(filename, FieldYear, movie.Year)
	filename = replaceField(filename, FieldDate, movie.ReleaseDate)
	filename = replaceField(filename, FieldExt, fileMovie.Extension)
	return filename
}

func GenerateEpisodeFilename(pattern string, show mediadata.TvShow, episode mediadata.Episode, fileEpisode mediascanner.Episode) string {
	filename := pattern
	filename = replaceField(filename, FieldName, show.Title)
	filename = replaceField(filename, FieldYear, show.Year)
	filename = replaceFieldInt(filename, FieldSeason, episode.SeasonNumber)
	filename = replaceFieldInt(filename, FieldEpisode, episode.EpisodeNumber)
	filename = replaceField(filename, FieldEpisodeTitle, episode.Name)
	filename = replaceField(filename, FieldExt, fileEpisode.Extension)
	return filename
}

func generateDefaultMovieFilename(fileMovie mediascanner.Movie) string {
	filename := fileMovie.Name
	if fileMovie.Year != 0 {
		filename += " - " + strconv.Itoa(fileMovie.Year)
	}
	filename += fileMovie.Extension
	return filename
}

func replaceField(pattern string, field Field, value string) string {
	return strings.ReplaceAll(pattern, string(field), value)
}

func replaceFieldInt(pattern string, field Field, value int) string {
	return strings.ReplaceAll(pattern, string(field), fmt.Sprintf("%02d", value))
}

```


# 📄 internal/mediascanner/filescanner/filescanner.go
```go
package filescanner

import (
	"context"
	"io/fs"
	"path/filepath"
	"slices"

	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/nouuu/gonamer/pkg/logger"
)

var (
	allowedExt = []string{".mkv", ".mp4", ".avi", ".mov", ".flv", ".wmv", ".webm", ".m4v", ".mpg", ".mpeg", ".3gp", ".3g2", ".ogv", ".ogg", ".drc", ".gif", ".gifv", ".mng", ".avi", ".mov", ".qt", ".wmv", ".yuv", ".rm", ".rmvb", ".asf", ".amv", ".mp4", ".m4p", ".m4v", ".mpg", ".mp2", ".mpeg", ".mpe", ".mpv", ".mpg", ".mpeg", ".m2v", ".m4v", ".svi", ".3gp", ".3g2", ".mxf", ".roq", ".nsv", ".flv", ".f4v", ".f4p", ".f4a", ".f4b"}
)

type FileScanner struct {
}

func New() mediascanner.MediaScanner {
	return &FileScanner{}
}

func (f *FileScanner) ScanMovies(ctx context.Context, path string, options ...mediascanner.ScanMoviesOptions) (movies []mediascanner.Movie, err error) {
	log := logger.FromContext(ctx)
	var opts mediascanner.ScanMoviesOptions
	if len(options) > 0 {
		opts = options[0]
	}
	files, err := scanDirectory(ctx, path, opts.Recursively)
	if err != nil {
		log.With("error", err).Error("Error scanning directory")
		return
	}

	for _, file := range files {
		if isFileAllowedExt(file) {
			movies = append(movies, parseMovieFileName(ctx, file))
		}
	}
	return
}

func (f *FileScanner) ScanEpisodes(ctx context.Context, path string, options ...mediascanner.ScanEpisodesOptions) (episodes []mediascanner.Episode, err error) {
	log := logger.FromContext(ctx)
	var opts mediascanner.ScanEpisodesOptions
	if len(options) > 0 {
		opts = options[0]
	}
	files, err := scanDirectory(ctx, path, opts.Recursively)
	if err != nil {
		log.With("error", err).Error("Error scanning directory")
		return
	}

	for _, file := range files {
		if isFileAllowedExt(file) {
			ctx = logger.InjectLogger(ctx, log.With("file", file))
			parsed := parseEpisodeFileName(ctx, file, opts.ExcludeUnparsed)
			if parsed.Name == "" {
				continue
			}
			episodes = append(episodes, parsed)
		}
	}
	return
}

func scanDirectory(ctx context.Context, path string, recursive bool) (files []string, err error) {
	log := logger.FromContext(ctx)
	err = filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			log.With("error", err).Error("Error accessing path")
			return err
		}

		if !d.IsDir() {
			// Append absolute path
			files = append(files, filePath)
		} else if !recursive && path != filePath {
			return filepath.SkipDir
		}

		return nil
	})
	return
}

func isFileAllowedExt(filename string) bool {
	return slices.Contains(allowedExt, filepath.Ext(filepath.Base(filename)))
}

```


# 📄 internal/mediascanner/filescanner/sanitazer.go
```go
package filescanner

import (
	"context"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/nouuu/gonamer/pkg/logger"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	capitaliser   = cases.Title(language.French)
	deleteRegexes = []*regexp.Regexp{
		regexp.MustCompile(`[\[\(].*?[\]\)]|-\s*\d+p.*`),
		regexp.MustCompile(`\s+$`),
		regexp.MustCompile(`\s(FR EN|FR-EN|MULTI|TRUEFRENCH|FRENCH|VFF)\s?`),
	}
	spaceRegexes = []*regexp.Regexp{
		regexp.MustCompile(`[^\pL\s_\d]+`),
	}
	extractDateRegex    = regexp.MustCompile(`^(.+)(19\d{2}|20\d{2}).*$`)
	tvShowRegex         = regexp.MustCompile(`^(.+?)[sS]?(\d+)[eExX](\d+)(?:.*|$)`)
	fallbackTvShowRegex = regexp.MustCompile(`^(.+?)(\d+)[a-z]{2}\s*Season\s*(\d+)(?:.*|$)`)
	episodeOnlyRegex    = regexp.MustCompile(`^(.+?)(\d{2,})(?:.*|$)`)
)

func parseMovieFileName(ctx context.Context, fileName string) (movie mediascanner.Movie) {
	filename := filepath.Base(fileName)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	movie.OriginalFilename = filename
	movie.FullPath = fileName
	movie.Extension = ext

	movie.Name, movie.Year = sanitizeMovieName(ctx, nameWithoutExt)

	return
}

func parseEpisodeFileName(ctx context.Context, fileName string, excludeUnparsed bool) (episode mediascanner.Episode) {
	filename := filepath.Base(fileName)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	episode.OriginalFilename = filename
	episode.FullPath = fileName
	episode.Extension = ext

	var ignore bool
	episode.Name, episode.Season, episode.Episode, ignore = sanitizeEpisodeName(ctx, nameWithoutExt, excludeUnparsed)

	if ignore {
		episode = mediascanner.Episode{
			OriginalFilename: filename,
		}
	}
	return
}

func sanitizeMovieName(ctx context.Context, nameWithoutExt string) (name string, year int) {
	log := logger.FromContext(ctx)
	nameWithoutExt = sanitizeString(nameWithoutExt)

	matches := extractDateRegex.FindStringSubmatch(nameWithoutExt)
	if len(matches) == 3 {
		name = strings.TrimSpace(matches[1])
		year, _ = strconv.Atoi(matches[2])
	} else {
		log.With("name", nameWithoutExt).Debug("Could not extract year from movie name")
		name = nameWithoutExt
	}
	return
}

func sanitizeEpisodeName(ctx context.Context, nameWithoutExt string, excludeUnparsed bool) (name string, season int, episode int, ignore bool) {
	nameWithoutExt = sanitizeString(nameWithoutExt)

	name, season, episode, ignore = parseEpisodeName(ctx, nameWithoutExt, excludeUnparsed)
	return
}

func parseEpisodeName(ctx context.Context, name string, excludeUnparsed bool) (string, int, int, bool) {
	log := logger.FromContext(ctx)

	if matches := tvShowRegex.FindStringSubmatch(name); len(matches) == 4 {
		return extractNameAndEpisode(matches, 1, 2, 3)
	}
	log.With("name", name).Debug("Could not extract season and episode from episode name, trying fallback regex")

	if matches := fallbackTvShowRegex.FindStringSubmatch(name); len(matches) == 4 {
		return extractNameAndEpisode(matches, 1, 2, 3)
	}
	log.With("name", name).Debug("Could not extract season and episode from episode name")

	if matches := episodeOnlyRegex.FindStringSubmatch(name); len(matches) == 3 {
		return extractNameAndEpisode(matches, 1, 0, 2) // Season set to 1
	}
	log.With("name", name).Debug("Could not extract episode from episode name")

	if excludeUnparsed {
		return name, 0, 0, true
	}
	return name, 1, 1, false
}

func extractNameAndEpisode(matches []string, nameIndex, seasonIndex, episodeIndex int) (string, int, int, bool) {
	name := strings.TrimSpace(matches[nameIndex])
	season := 1
	if seasonIndex != 0 {
		season, _ = strconv.Atoi(matches[seasonIndex])
	}
	episode, _ := strconv.Atoi(matches[episodeIndex])
	return name, season, episode, false
}

func sanitizeString(str string) string {
	for _, regex := range spaceRegexes {
		str = regex.ReplaceAllString(str, " ")
	}

	for _, regex := range deleteRegexes {
		str = regex.ReplaceAllString(str, "")
	}

	return capitaliser.String(strings.TrimSpace(str))
}

```


# 📄 internal/mediascanner/mediascanner.go
```go
package mediascanner

import "context"

type ScanMoviesOptions struct {
	Recursively bool
}

type ScanEpisodesOptions struct {
	Recursively     bool
	ExcludeUnparsed bool
}

type Movie struct {
	OriginalFilename string
	FullPath         string
	Name             string
	Year             int
	Extension        string
}

type Episode struct {
	OriginalFilename string
	FullPath         string
	Name             string
	Season           int
	Episode          int
	Extension        string
}

type MediaScanner interface {
	ScanMovies(ctx context.Context, path string, options ...ScanMoviesOptions) ([]Movie, error)
	ScanEpisodes(ctx context.Context, path string, options ...ScanEpisodesOptions) ([]Episode, error)
}

```


# 📄 main.go
```go
package main

import (
	"context"
	"fmt"

	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/nouuu/gonamer/internal/mediascanner/filescanner"
	"github.com/nouuu/gonamer/pkg/logger"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger.SetLoggerLevel(zapcore.ErrorLevel)
	ctx := context.Background()

	//config := conf.LoadConfig()

	scanner := filescanner.New()
	//movieClient, err := tmdb.NewMovieClient(config.TMDBAPIKey, tmdb.WithLang("fr-FR"))
	//if err != nil {
	//	log.Fatalf("Error creating movie client: %v", err)
	//}
	//mediaRenamer := mediarenamer.NewMediaRenamer(movieClient)

	episodes, err := scanner.ScanEpisodes(ctx, "/mnt/nfs/Media/TV", mediascanner.ScanEpisodesOptions{Recursively: true})

	if err != nil {
		logger.FromContext(ctx).Errorf("Error scanning episodes: %v", err)
	}
	for _, episode := range episodes {
		fmt.Printf("Episode: %s %dx%02d\n", episode.Name, episode.Season, episode.Episode)
	}

}

```


# 📄 pkg/logger/logger.go
```go
package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger
var config zap.Config

func init() {
	config = zap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	log = logger.Sugar()
}

func InjectLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, "logger", logger)
}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	if contextLogger, ok := ctx.Value("logger").(*zap.SugaredLogger); ok {
		return contextLogger
	}
	InjectLogger(ctx, log)
	return log
}

func SetLoggerLevel(level zapcore.Level) {
	config.Level = zap.NewAtomicLevelAt(level)
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	log = logger.Sugar()
}

func SetLoggerOutput(output zapcore.WriteSyncer) {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config.EncoderConfig),
		output,
		config.Level,
	)
	logger := zap.New(core)
	log = logger.Sugar()
}

```


# 📄 scripts/scan_project.go
```go
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type ProjectScanner struct {
	rootDir        string
	ignorePatterns []string
	structure      []string
	contents       []string
}

func NewProjectScanner(rootDir string) *ProjectScanner {
	return &ProjectScanner{
		rootDir: rootDir,
		ignorePatterns: []string{
			".git",
			"node_modules",
			".env",
			".idea",
			"vendor",
			"dist",
			"build",
			"mediatracker.log",
			"project_knowledge.md",
			"go.sum",
		},
		structure: make([]string, 0),
		contents:  make([]string, 0),
	}
}

func (ps *ProjectScanner) shouldIgnore(path string) bool {
	for _, pattern := range ps.ignorePatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func (ps *ProjectScanner) addToStructure(path string, info fs.FileInfo, depth int) {
	indent := strings.Repeat("  ", depth)
	if info.IsDir() {
		ps.structure = append(ps.structure, fmt.Sprintf("%s- 📁 %s", indent, info.Name()))
	} else {
		ps.structure = append(ps.structure, fmt.Sprintf("%s- 📄 %s", indent, info.Name()))
	}
}

func (ps *ProjectScanner) addToContents(path string, relPath string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", path, err)
	}

	ext := filepath.Ext(path)
	if ext != "" {
		ext = ext[1:] // Remove the dot
	}

	fileContent := fmt.Sprintf("\n# 📄 %s\n```%s\n%s\n```\n",
		relPath, ext, string(content))
	ps.contents = append(ps.contents, fileContent)
	return nil
}

func (ps *ProjectScanner) scan() error {
	return filepath.Walk(ps.rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		// Skip ignored patterns
		if ps.shouldIgnore(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(ps.rootDir, path)
		if err != nil {
			return fmt.Errorf("error getting relative path for %s: %w", path, err)
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Calculate depth for indentation
		depth := len(strings.Split(relPath, string(os.PathSeparator))) - 1

		// Add to structure
		ps.addToStructure(path, info, depth)

		// If it's a file, add its contents
		if !info.IsDir() {
			if err := ps.addToContents(path, relPath); err != nil {
				fmt.Printf("Warning: %v\n", err)
			}
		}

		return nil
	})
}

func (ps *ProjectScanner) generateOutput() string {
	return fmt.Sprintf("# Project Structure\n\n%s\n\n# Files Content\n%s",
		strings.Join(ps.structure, "\n"),
		strings.Join(ps.contents, "\n"))
}

func main() {
	outputFile := flag.String("o", "project_knowledge.md", "Output file path")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Error: Please provide the root directory path")
		fmt.Println("Usage: scan-project <root-dir> [-o output-file]")
		os.Exit(1)
	}

	rootDir := args[0]
	scanner := NewProjectScanner(rootDir)

	if err := scanner.scan(); err != nil {
		fmt.Printf("Error scanning project: %v\n", err)
		os.Exit(1)
	}

	output := scanner.generateOutput()
	if err := os.WriteFile(*outputFile, []byte(output), 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated project structure in %s\n", *outputFile)
}

```
