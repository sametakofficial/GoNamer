package mediarenamer

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp" // HATA DÜZELTME: Bu satır eklendi.
	"strings"
	"sync"
	"time"

	"github.com/nouuu/gonamer/internal/mediadata"
	"github.com/nouuu/gonamer/internal/mediascanner"
	"github.com/nouuu/gonamer/pkg/config"
	"github.com/nouuu/gonamer/pkg/logger"
	"go.uber.org/zap"
)

// DÜZELTME: Gereksiz olan filescanner import'u kaldırıldı.

func findUniqueFilename(destination string) string {
	if _, err := os.Stat(destination); os.IsNotExist(err) {
		return destination
	}
	ext := filepath.Ext(destination)
	baseName := destination[0 : len(destination)-len(ext)]
	for i := 1; ; i++ {
		newDestination := fmt.Sprintf("%s (%d)%s", baseName, i, ext)
		if _, err := os.Stat(newDestination); os.IsNotExist(err) {
			return newDestination
		}
	}
}

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

// DÜZELTME: Bu fonksiyon zincirinin imzaları, config parametresini taşıyacak şekilde düzeltildi.
func (mr *MediaRenamer) FindMovieSuggestions(ctx context.Context, movies []mediascanner.Movie, maxResults int, cfg *config.Config, callback ...FindMovieSuggestionCallback) []MovieSuggestions {
	log := logger.FromContext(ctx)
	start := time.Now()
	log.Infof("Getting suggestions for %d movies", len(movies))
	suggestions := mr.getMoviesSuggestions(ctx, movies, maxResults, cfg, log, callback...)
	log.Infof("Finished getting suggestions for %d movies in %s", len(movies), time.Since(start))
	return suggestions
}
func (mr *MediaRenamer) FindEpisodeSuggestions(ctx context.Context, episodes []mediascanner.Episode, maxResults int, cfg *config.Config, callback ...FindEpisodeSuggestionCallback) []EpisodeSuggestions {
	log := logger.FromContext(ctx)
	start := time.Now()
	log.Infof("Getting suggestions for %d episodes", len(episodes))

	suggestions := mr.getEpisodesSuggestions(ctx, episodes, maxResults, cfg, log, callback...)

	log.Infof("Finished getting suggestions for %d episodes in %s", len(episodes), time.Since(start))
	return suggestions
}
func (mr *MediaRenamer) RenameMovie(ctx context.Context, fileMovie mediascanner.Movie, mediadataMovie mediadata.Movie, pattern string, dryrun bool) (string, error) {
	filename := GenerateMovieFilename(pattern, mediadataMovie, fileMovie)
	var destination string
	if filepath.IsAbs(filename) {
		destination = filename
	} else {
		destination = filepath.Join(filepath.Dir(fileMovie.FullPath), filename)
	}
	return mr.RenameFile(ctx, fileMovie.FullPath, destination, dryrun)
}

func (mr *MediaRenamer) RenameEpisode(ctx context.Context, fileEpisode mediascanner.Episode, tvShow mediadata.TvShow, episode mediadata.Episode, pattern string, dryrun bool) (string, error) {
	filename := GenerateEpisodeFilename(pattern, tvShow, episode, fileEpisode)
	var destination string
	if filepath.IsAbs(filename) {
		destination = filename
	} else {
		destination = filepath.Join(filepath.Dir(fileEpisode.FullPath), filename)
	}
	return mr.RenameFile(ctx, fileEpisode.FullPath, destination, dryrun)
}

func (mr *MediaRenamer) RenameFile(ctx context.Context, source, destination string, dryrun bool) (string, error) {
	finalDestination := findUniqueFilename(destination)
	log := logger.FromContext(ctx)
	if dryrun {
		return finalDestination, nil
	}
	var err error
	destDir := filepath.Dir(finalDestination)
	err = os.MkdirAll(destDir, 0755)
	if err != nil {
		log.With("error", err).Error("Error creating destination directory")
		return "", err
	}
	err = os.Rename(source, finalDestination)
	if err != nil {
		log.With("error", err).Error("Error renaming file")
		return "", err
	}
	return finalDestination, nil
}
func (mr *MediaRenamer) SuggestMovies(ctx context.Context, movie mediascanner.Movie, maxResults int, cfg *config.Config) (suggestions MovieSuggestions, err error) {
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

func (mr *MediaRenamer) SuggestEpisodes(ctx context.Context, episode mediascanner.Episode, maxResults int, cfg *config.Config) (suggestions EpisodeSuggestions, err error) {
	log := logger.FromContext(ctx).With("episode", episode)
	suggestions.Episode = episode
	log.Debugf("Plan A: Searching for show with name '%s'", episode.Name)
	suggestions, err = mr.searchEpisodesByName(ctx, episode, maxResults)
	if err == nil && len(suggestions.SuggestedEpisodes) > 0 {
		log.Infof("Plan A successful: Found %d suggestions for '%s'", len(suggestions.SuggestedEpisodes), episode.Name)
		return suggestions, nil
	}
	log.Warnf("Plan A failed. Trying Plan B: Parsing filename '%s'", episode.OriginalFilename)
	filenameOnly := strings.TrimSuffix(episode.OriginalFilename, episode.Extension)
	re := regexp.MustCompile(`(?i)S\d{1,2}E\d{1,3}`)
	potentialShowName := strings.TrimSpace(re.Split(filenameOnly, -1)[0])
	if strings.EqualFold(potentialShowName, episode.Name) {
		log.Debugf("Plan B skipped: Filename-based name is the same as folder-based name.")
		return suggestions, err
	}
	log.Debugf("Plan B: Searching for show with name '%s'", potentialShowName)
	fallbackEpisode := episode
	fallbackEpisode.Name = potentialShowName
	fallbackSuggestions, fallbackErr := mr.searchEpisodesByName(ctx, fallbackEpisode, maxResults)
	if fallbackErr == nil && len(fallbackSuggestions.SuggestedEpisodes) > 0 {
		log.Infof("Plan B successful: Found %d suggestions for '%s'", len(fallbackSuggestions.SuggestedEpisodes), potentialShowName)
		return fallbackSuggestions, nil
	}
	log.Errorf("Plan B also failed. Could not find any match.")
	return suggestions, err
}

func (mr *MediaRenamer) searchEpisodesByName(ctx context.Context, episode mediascanner.Episode, maxResults int) (suggestions EpisodeSuggestions, err error) {
	log := logger.FromContext(ctx)
	suggestions.Episode = episode
	tvShows, err := mr.tvShowClient.SearchTvShow(ctx, episode.Name, 0, 1)
	if err != nil {
		return suggestions, err
	}
	if tvShows.Totals == 0 {
		return suggestions, errors.New("no tv show found")
	}
	for _, tvShow := range tvShows.TvShows {
		foundEpisode, err := mr.tvShowClient.GetEpisode(ctx, tvShow.ID, episode.Season, episode.Episode)
		if err != nil {
			log.Debugf("Could not find S%02dE%02d in show '%s'. Error: %v", episode.Season, episode.Episode, tvShow.Title, err)
			continue
		}
		suggestions.SuggestedEpisodes = append(suggestions.SuggestedEpisodes, SuggestedEpisode{
			TvShow:  tvShow,
			Episode: foundEpisode,
		})
	}
	if len(suggestions.SuggestedEpisodes) == 0 {
		return suggestions, errors.New("show found, but specific episode not found")
	}
	if len(suggestions.SuggestedEpisodes) > maxResults {
		suggestions.SuggestedEpisodes = suggestions.SuggestedEpisodes[:maxResults]
	}
	return suggestions, nil
}

func (mr *MediaRenamer) getMoviesSuggestions(ctx context.Context, movies []mediascanner.Movie, maxResults int, cfg *config.Config, log *zap.SugaredLogger, callback ...FindMovieSuggestionCallback) (movieSuggestion []MovieSuggestions) {
	var wg sync.WaitGroup
	suggestionsCh := make(chan MovieSuggestions, len(movies))
	semaphore := make(chan struct{}, 5)
	for _, movie := range movies {
		wg.Add(1)
		go func(movie mediascanner.Movie) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			suggestions, err := mr.getMovieSuggestions(ctx, movie, maxResults, cfg)
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

func (mr *MediaRenamer) getEpisodesSuggestions(ctx context.Context, episodes []mediascanner.Episode, maxResults int, cfg *config.Config, log *zap.SugaredLogger, callback ...FindEpisodeSuggestionCallback) (episodeSuggestions []EpisodeSuggestions) {
	var wg sync.WaitGroup
	suggestionsCh := make(chan EpisodeSuggestions, len(episodes))
	semaphore := make(chan struct{}, 5)

	for _, episode := range episodes {
		wg.Add(1)
		go func(episode mediascanner.Episode) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			suggestions, err := mr.getEpisodeSuggestions(ctx, episode, maxResults, cfg)
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
func (mr *MediaRenamer) getMovieSuggestions(ctx context.Context, movie mediascanner.Movie, maxResults int, cfg *config.Config) (suggestions MovieSuggestions, err error) {
	log := logger.FromContext(ctx).With("movie", movie)
	suggestions, err = mr.SuggestMovies(ctx, movie, maxResults, cfg)
	if err != nil {
		log.With("movie", movie).With("error", err).Error("Error suggesting movie")
		return
	}
	if len(suggestions.SuggestedMovies) > 0 {
		output := fmt.Sprintf("Suggested movie '%s (%d)' -> '%s (%s)'", suggestions.Movie.Name, suggestions.Movie.Year, suggestions.SuggestedMovies[0].Title, suggestions.SuggestedMovies[0].Year)
		log.With("suggestions", len(suggestions.SuggestedMovies)).Debug(output)
	}
	return
}
func (mr *MediaRenamer) getEpisodeSuggestions(ctx context.Context, episode mediascanner.Episode, maxResults int, cfg *config.Config) (suggestions EpisodeSuggestions, err error) {
	log := logger.FromContext(ctx).With("episode", episode)

	suggestions, err = mr.SuggestEpisodes(ctx, episode, maxResults, cfg)
	if err != nil {
		log.With("episode", episode).With("error", err).Error("Error suggesting episode")
		return
	}
	if len(suggestions.SuggestedEpisodes) > 0 {
		output := fmt.Sprintf("Suggested episode '%s' S%02dE%02d -> '%s' S%02dE%02d - '%s'", suggestions.Episode.Name, suggestions.Episode.Season, suggestions.Episode.Episode, suggestions.SuggestedEpisodes[0].TvShow.Title, suggestions.SuggestedEpisodes[0].Episode.SeasonNumber, suggestions.SuggestedEpisodes[0].Episode.EpisodeNumber, suggestions.SuggestedEpisodes[0].Episode.Name)
		log.With("suggestions", len(suggestions.SuggestedEpisodes)).Debug(output)
	}
	return
}