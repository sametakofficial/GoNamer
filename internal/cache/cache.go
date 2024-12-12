package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nouuu/gonamer/pkg/logger"
	gocache "github.com/patrickmn/go-cache"

	"github.com/eko/gocache/lib/v4/store"
	"github.com/nouuu/gonamer/internal/mediadata"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
)

const (
	movieSearchKey    = "search:movie:%s:year:%d:page:%d"
	tvShowSearchKey   = "search:tvshow:%s:year:%d:page:%d"
	movieKey          = "movie:%s"
	movieDetailsKey   = "movie:details:%s"
	tvShowKey         = "tvshow:%s"
	tvShowDetailsKey  = "tvshow:details:%s"
	seasonEpisodesKey = "tvshow:%s:season:%d"
	episodeKey        = "tvshow:%s:season:%d:episode:%d"
)

type Cache interface {
	// Recherches
	SetMovieSearch(ctx context.Context, query string, year int, page int, results mediadata.MovieResults) error
	GetMovieSearch(ctx context.Context, query string, year int, page int) (mediadata.MovieResults, error)
	SetTvShowSearch(ctx context.Context, query string, year int, page int, results mediadata.TvShowResults) error
	GetTvShowSearch(ctx context.Context, query string, year int, page int) (mediadata.TvShowResults, error)

	// Films
	SetMovie(ctx context.Context, id string, movie mediadata.Movie) error
	GetMovie(ctx context.Context, id string) (mediadata.Movie, error)
	SetMovieDetails(ctx context.Context, id string, details mediadata.MovieDetails) error
	GetMovieDetails(ctx context.Context, id string) (mediadata.MovieDetails, error)

	// Séries
	SetTvShow(ctx context.Context, id string, tvShow mediadata.TvShow) error
	GetTvShow(ctx context.Context, id string) (mediadata.TvShow, error)
	SetTvShowDetails(ctx context.Context, id string, details mediadata.TvShowDetails) error
	GetTvShowDetails(ctx context.Context, id string) (mediadata.TvShowDetails, error)

	// Episodes
	SetSeasonEpisodes(ctx context.Context, showID string, seasonNum int, episodes []mediadata.Episode) error
	GetSeasonEpisodes(ctx context.Context, showID string, seasonNum int) ([]mediadata.Episode, error)
	SetEpisode(ctx context.Context, showID string, seasonNum int, episodeNum int, episode mediadata.Episode) error
	GetEpisode(ctx context.Context, showID string, seasonNum int, episodeNum int) (mediadata.Episode, error)
}

func NewGoCache(ctx context.Context) (Cache, error) {

	goCacheClient := gocache.New(5*time.Minute, 10*time.Minute)

	if err := goCacheClient.LoadFile("gonamer-cache.gob"); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load cache file: %w", err)
		}
	}

	goCacheStore := gocache_store.NewGoCache(goCacheClient)

	go func(ctx context.Context) {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := goCacheClient.SaveFile("gonamer-cache.gob"); err != nil {
					logger.FromContext(ctx).With("error", err).Error("failed to save cache file")
				}

			}
		}
	}(ctx)

	return &goCache{
		marshaler: marshaler.New(cache.New[any](goCacheStore)),
	}, nil
}

type goCache struct {
	marshaler *marshaler.Marshaler
}

func (g *goCache) SetMovieSearch(ctx context.Context, query string, year int, page int, results mediadata.MovieResults) error {
	key := fmt.Sprintf(movieSearchKey, query, year, page)
	return g.marshaler.Set(ctx, key, results, store.WithExpiration(1*time.Hour))
}

func (g *goCache) GetMovieSearch(ctx context.Context, query string, year int, page int) (mediadata.MovieResults, error) {
	key := fmt.Sprintf(movieSearchKey, query, year, page)
	results, err := g.marshaler.Get(ctx, key, new(mediadata.MovieResults))
	if err != nil {
		return mediadata.MovieResults{}, err
	}
	return *results.(*mediadata.MovieResults), nil
}

func (g *goCache) SetTvShowSearch(ctx context.Context, query string, year int, page int, results mediadata.TvShowResults) error {
	key := fmt.Sprintf(tvShowSearchKey, query, year, page)
	return g.marshaler.Set(ctx, key, results, store.WithExpiration(1*time.Hour))
}

func (g *goCache) GetTvShowSearch(ctx context.Context, query string, year int, page int) (mediadata.TvShowResults, error) {
	key := fmt.Sprintf(tvShowSearchKey, query, year, page)
	results, err := g.marshaler.Get(ctx, key, new(mediadata.TvShowResults))
	if err != nil {
		return mediadata.TvShowResults{}, err
	}
	return *results.(*mediadata.TvShowResults), nil
}

// Films
func (g *goCache) SetMovie(ctx context.Context, id string, movie mediadata.Movie) error {
	key := fmt.Sprintf(movieKey, id)
	return g.marshaler.Set(ctx, key, movie, store.WithExpiration(24*time.Hour))
}

func (g *goCache) GetMovie(ctx context.Context, id string) (mediadata.Movie, error) {
	key := fmt.Sprintf(movieKey, id)
	result, err := g.marshaler.Get(ctx, key, new(mediadata.Movie))
	if err != nil {
		return mediadata.Movie{}, err
	}
	return *result.(*mediadata.Movie), nil
}

func (g *goCache) SetMovieDetails(ctx context.Context, id string, details mediadata.MovieDetails) error {
	key := fmt.Sprintf(movieDetailsKey, id)
	return g.marshaler.Set(ctx, key, details, store.WithExpiration(24*time.Hour))
}

func (g *goCache) GetMovieDetails(ctx context.Context, id string) (mediadata.MovieDetails, error) {
	key := fmt.Sprintf(movieDetailsKey, id)
	result, err := g.marshaler.Get(ctx, key, new(mediadata.MovieDetails))
	if err != nil {
		return mediadata.MovieDetails{}, err
	}
	return *result.(*mediadata.MovieDetails), nil
}

// Séries
func (g *goCache) SetTvShow(ctx context.Context, id string, tvShow mediadata.TvShow) error {
	key := fmt.Sprintf(tvShowKey, id)
	return g.marshaler.Set(ctx, key, tvShow, store.WithExpiration(24*time.Hour))
}

func (g *goCache) GetTvShow(ctx context.Context, id string) (mediadata.TvShow, error) {
	key := fmt.Sprintf(tvShowKey, id)
	result, err := g.marshaler.Get(ctx, key, new(mediadata.TvShow))
	if err != nil {
		return mediadata.TvShow{}, err
	}
	return *result.(*mediadata.TvShow), nil
}

func (g *goCache) SetTvShowDetails(ctx context.Context, id string, details mediadata.TvShowDetails) error {
	key := fmt.Sprintf(tvShowDetailsKey, id)
	return g.marshaler.Set(ctx, key, details, store.WithExpiration(24*time.Hour))
}

func (g *goCache) GetTvShowDetails(ctx context.Context, id string) (mediadata.TvShowDetails, error) {
	key := fmt.Sprintf(tvShowDetailsKey, id)
	result, err := g.marshaler.Get(ctx, key, new(mediadata.TvShowDetails))
	if err != nil {
		return mediadata.TvShowDetails{}, err
	}
	return *result.(*mediadata.TvShowDetails), nil
}

// Episodes
func (g *goCache) SetSeasonEpisodes(ctx context.Context, showID string, seasonNum int, episodes []mediadata.Episode) error {
	key := fmt.Sprintf(seasonEpisodesKey, showID, seasonNum)
	return g.marshaler.Set(ctx, key, episodes, store.WithExpiration(24*time.Hour))
}

func (g *goCache) GetSeasonEpisodes(ctx context.Context, showID string, seasonNum int) ([]mediadata.Episode, error) {
	key := fmt.Sprintf(seasonEpisodesKey, showID, seasonNum)
	result, err := g.marshaler.Get(ctx, key, new([]mediadata.Episode))
	if err != nil {
		return nil, err
	}
	return *result.(*[]mediadata.Episode), nil
}

func (g *goCache) SetEpisode(ctx context.Context, showID string, seasonNum int, episodeNum int, episode mediadata.Episode) error {
	key := fmt.Sprintf(episodeKey, showID, seasonNum, episodeNum)
	return g.marshaler.Set(ctx, key, episode, store.WithExpiration(24*time.Hour))
}

func (g *goCache) GetEpisode(ctx context.Context, showID string, seasonNum int, episodeNum int) (mediadata.Episode, error) {
	key := fmt.Sprintf(episodeKey, showID, seasonNum, episodeNum)
	result, err := g.marshaler.Get(ctx, key, new(mediadata.Episode))
	if err != nil {
		return mediadata.Episode{}, err
	}
	return *result.(*mediadata.Episode), nil
}
