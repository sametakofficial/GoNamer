package mediarenamer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

func NewMediaRenamer(movieClient mediadata.MovieClient) *MediaRenamer {
	return &MediaRenamer{movieClient: movieClient}
}

func (mr *MediaRenamer) RenameMovies(ctx context.Context, movies []mediascanner.Movie) ([]MovieSuggestions, error) {
	log := logger.FromContext(ctx)
	var suggestions []MovieSuggestions
	start := time.Now()
	log.Infof("Renaming %d movies", len(movies))

	moviesSuggestions := mr.getMoviesSuggestions(ctx, movies, log)

	log.Infof("Renaming took %s for %d movies (found %d - %.2f%%)", time.Since(start), len(movies), len(moviesSuggestions), float64(len(moviesSuggestions))/float64(len(movies))*100)
	return suggestions, nil
}

func (mr *MediaRenamer) getMoviesSuggestions(ctx context.Context, movies []mediascanner.Movie, log *zap.SugaredLogger) (movieSuggestion []MovieSuggestions) {
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
			if err != nil {
				log.With("movie", movie).With("error", err).Error("Error suggesting movie")
				return
			}
			output := fmt.Sprintf("Suggested movie '%s - %d' -> '%s - %s'", suggestions.Movie.Name, suggestions.Movie.Year, suggestions.SuggestedMovies[0].Title, suggestions.SuggestedMovies[0].Year)
			if suggestions.Movie.Name == suggestions.SuggestedMovies[0].Title && strconv.Itoa(suggestions.Movie.Year) == suggestions.SuggestedMovies[0].Year {
				output += " (unchanged)"
			}
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

func (mr *MediaRenamer) RenameMovie(ctx context.Context, fileMovie mediascanner.Movie, mediadataMovie mediadata.Movie) error {
	log := logger.FromContext(ctx)
	filename := generateMovieFilename(mediadataMovie, fileMovie)
	fmt.Printf("Renaming file %s -> %s\n", fileMovie.OriginalFilename, filename)
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

func generateMovieFilename(movie mediadata.Movie, fileMovie mediascanner.Movie) string {
	return fmt.Sprintf("%s - %s%s", movie.Title, movie.Year, fileMovie.Extension)
}
