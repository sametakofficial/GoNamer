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

type SuggestedEpisode struct {
	TvShow  mediadata.TvShow
	Episode mediadata.Episode
}

type EpisodeSuggestions struct {
	Episode           mediascanner.Episode
	SuggestedEpisodes []SuggestedEpisode
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

	var destination string

	if filepath.IsAbs(filename) {
		destination = filename
	} else {
		destination = filepath.Join(filepath.Dir(fileMovie.FullPath), filename)
	}

	return mr.RenameFile(ctx, fileMovie.FullPath, destination, dryrun)
}

func (mr *MediaRenamer) RenameEpisode(ctx context.Context, fileEpisode mediascanner.Episode, tvShow mediadata.TvShow, episode mediadata.Episode, pattern string, dryrun bool) error {
	// "{name} - {season}x{episode} - {episode_title}{extension}" <3
	filename := GenerateEpisodeFilename(pattern, tvShow, episode, fileEpisode)

	var destination string

	if filepath.IsAbs(filename) {
		destination = filename
	} else {
		destination = filepath.Join(filepath.Dir(fileEpisode.FullPath), filename)
	}
	return mr.RenameFile(ctx, fileEpisode.FullPath, destination, dryrun)
}

func (mr *MediaRenamer) RenameFile(ctx context.Context, source, destination string, dryrun bool) error {
	log := logger.FromContext(ctx)
	log.Infof("Renaming file %s -> %s", source, destination)
	if dryrun {
		return nil
	}
	var err error

	destDir := filepath.Dir(destination)
	err = os.MkdirAll(destDir, 0755)
	if err != nil {
		log.With("error", err).Error("Error creating destination directory")
		return err
	}

	err = os.Rename(source, destination)
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
	movies, err := mr.movieClient.SearchMovie(ctx, movie.Name, movie.Year, 1)
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
	if len(suggestions.SuggestedMovies) > maxResults {
		suggestions.SuggestedMovies = suggestions.SuggestedMovies[:maxResults]
	}

	return
}

func (mr *MediaRenamer) SuggestEpisodes(ctx context.Context, episode mediascanner.Episode, maxResults int) (suggestions EpisodeSuggestions, err error) {
	log := logger.FromContext(ctx).With("episode", episode)
	suggestions.Episode = episode
	maxResults = int(math.Max(math.Min(float64(maxResults), 100), 1))
	tvShow, err := mr.tvShowClient.SearchTvShow(ctx, episode.Name, 0, 1)
	if err != nil {
		log.With("error", err).Error("Error searching tv show")
		return
	}
	if tvShow.Totals == 0 {
		log.Warnf("No tv show found for %s", episode.Name)
		err = errors.New("no tv show found")
		return
	}
	for _, tvShow := range tvShow.TvShows {
		episodes, err := mr.tvShowClient.GetEpisode(ctx, tvShow.ID, episode.Season, episode.Episode)
		if err != nil {
			log.With("error", err).Error("Error getting episode")
		}
		suggestions.SuggestedEpisodes = append(suggestions.SuggestedEpisodes, struct {
			TvShow  mediadata.TvShow
			Episode mediadata.Episode
		}{TvShow: tvShow, Episode: episodes})
	}

	if len(suggestions.SuggestedEpisodes) > maxResults {
		suggestions.SuggestedEpisodes = suggestions.SuggestedEpisodes[:maxResults]
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
