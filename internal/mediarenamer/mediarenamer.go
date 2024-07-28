package mediarenamer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nouuu/mediatracker/internal/logger"
	"github.com/nouuu/mediatracker/internal/mediadata"
	"github.com/nouuu/mediatracker/internal/mediascanner"
	"go.uber.org/zap"
)

type MediaRenamer struct {
	movieClient mediadata.MovieClient
}

type MovieSuggestions struct {
	Movie           mediascanner.Movie
	SuggestedMovies []mediadata.Movie
}

type FindSuggestionCallback func(suggestion MovieSuggestions, err error)

func NewMediaRenamer(movieClient mediadata.MovieClient) *MediaRenamer {
	return &MediaRenamer{movieClient: movieClient}
}

func (mr *MediaRenamer) FindSuggestions(ctx context.Context, movies []mediascanner.Movie, callback ...FindSuggestionCallback) ([]MovieSuggestions, error) {
	log := logger.FromContext(ctx)
	start := time.Now()
	log.Infof("Getting suggestions for %d movies", len(movies))

	suggestions := mr.getMoviesSuggestions(ctx, movies, log, callback...)

	log.Infof("Finished getting suggestions for %d movies in %s", len(movies), time.Since(start))
	return suggestions, nil
}

func (mr *MediaRenamer) getMoviesSuggestions(ctx context.Context, movies []mediascanner.Movie, log *zap.SugaredLogger, callback ...FindSuggestionCallback) (movieSuggestion []MovieSuggestions) {
	var wg sync.WaitGroup
	suggestionsCh := make(chan MovieSuggestions, len(movies))
	semaphore := make(chan struct{}, 5) // Limit to 5 concurrent threads

	for _, movie := range movies {
		wg.Add(1)
		go func(movie mediascanner.Movie) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			suggestions, err := mr.SuggestMovies(ctx, movie)
			if len(callback) > 0 {
				for _, cb := range callback {
					cb(suggestions, err)
				}
			}
			if err != nil {
				log.With("movie", movie).With("error", err).Error("Error suggesting movie")
				return
			}
			output := fmt.Sprintf("Suggested movie '%s (%d)' -> '%s (%s)'", suggestions.Movie.Name, suggestions.Movie.Year, suggestions.SuggestedMovies[0].Title, suggestions.SuggestedMovies[0].Year)
			log.With("suggestions", len(suggestions.SuggestedMovies)).Debug(output)
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

func (mr *MediaRenamer) RenameMovie(ctx context.Context, fileMovie mediascanner.Movie, mediadataMovie mediadata.Movie, dryrun bool) error {
	log := logger.FromContext(ctx)
	filename := generateMovieFilename("{name} - {year}{extension}", mediadataMovie, fileMovie)
	fmt.Printf("Renaming file %s -> %s\n", fileMovie.OriginalFilename, filename)
	if dryrun {
		return nil
	}
	err := os.Rename(fileMovie.FullPath, filepath.Join(filepath.Dir(fileMovie.FullPath), filename))
	if err != nil {
		log.With("error", err).Error("Error renaming file")
		return err
	}
	return nil
}

func (mr *MediaRenamer) SuggestMovies(ctx context.Context, movie mediascanner.Movie) (suggestions MovieSuggestions, err error) {
	log := logger.FromContext(ctx).With("movie", movie)
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
	suggestions.Movie = movie
	suggestions.SuggestedMovies = movies.Movies
	if len(suggestions.SuggestedMovies) > 5 {
		suggestions.SuggestedMovies = suggestions.SuggestedMovies[:5]
	}

	return
}
