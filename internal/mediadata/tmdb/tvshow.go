package tmdb

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cyruzin/golang-tmdb"
	"github.com/nouuu/gonamer/internal/cache"
	"github.com/nouuu/gonamer/internal/mediadata"
	"github.com/nouuu/gonamer/pkg/logger"
)

func NewTvShowClient(APIKey string, cache cache.Cache, opts ...OptFunc) (mediadata.TvShowClient, error) {
	o := defaultOpts(APIKey)
	for _, optF := range opts {
		optF(&o.Opts)
	}

	client, err := tmdb.Init(o.APIKey)
	if err != nil {
		return nil, err
	}
	return &tmdbClient{client: client, cache: cache, opts: o}, nil
}

func (t *tmdbClient) SearchTvShow(ctx context.Context, query string, year int, page int) (mediadata.TvShowResults, error) {
	if result, err := t.cache.GetTvShowSearch(ctx, query, year, page); err == nil {
		return result, nil
	}
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
	results := mediadata.TvShowResults{
		TvShows:        buildTvShowFromResult(searchTvShows.SearchTVShowsResults),
		Totals:         searchTvShows.TotalResults,
		ResultsPerPage: 20,
	}
	if err := t.cache.SetTvShowSearch(ctx, query, year, page, results); err != nil {
		logger.FromContext(ctx).With("error", err).Error("failed to cache tv show search results")
	}
	return results, nil
}

func (t *tmdbClient) GetTvShow(ctx context.Context, id string) (mediadata.TvShow, error) {
	if show, err := t.cache.GetTvShow(ctx, id); err == nil {
		return show, nil
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return mediadata.TvShow{}, err
	}
	tvShowDetails, err := t.client.GetTVDetails(idInt, cfgMap(t.opts))
	if err != nil {
		return mediadata.TvShow{}, err
	}
	tvShow := buildTvShow(tvShowDetails)
	if err := t.cache.SetTvShow(ctx, id, tvShow); err != nil {
		logger.FromContext(ctx).With("error", err).Error("failed to cache tv show")
	}
	return tvShow, nil
}

func (t *tmdbClient) GetTvShowDetails(ctx context.Context, id string) (mediadata.TvShowDetails, error) {
	if details, err := t.cache.GetTvShowDetails(ctx, id); err == nil {
		return details, nil
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return mediadata.TvShowDetails{}, err
	}
	tvShowDetails, err := t.client.GetTVDetails(idInt, cfgMap(t.opts, map[string]string{
		"append_to_response": "credits",
	}))
	if err != nil {
		return mediadata.TvShowDetails{}, err
	}
	details := buildTvShowDetails(tvShowDetails)
	if err := t.cache.SetTvShowDetails(ctx, id, details); err != nil {
		logger.FromContext(ctx).With("error", err).Error("failed to cache tv show details")
	}
	return details, nil
}

func (t *tmdbClient) GetEpisode(ctx context.Context, id string, seasonNumber int, episodeNumber int) (mediadata.Episode, error) {
	if episode, err := t.cache.GetEpisode(ctx, id, seasonNumber, episodeNumber); err == nil {
		return episode, nil
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return mediadata.Episode{}, err
	}

	/*episodeDetails, err := t.client.GetTVEpisodeDetails(idInt, seasonNumber, episodeNumber, cfgMap(t.opts))
	if err != nil {
		return mediadata.Episode{}, err
	}*/

	season, err := t.client.GetTVSeasonDetails(idInt, seasonNumber, cfgMap(t.opts))
	if err != nil {
		return mediadata.Episode{}, err
	}

	episodes := make([]mediadata.Episode, 0, len(season.Episodes))

	for _, episode := range season.Episodes {
		episode := buildEpisode(struct {
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
		})
		episodes = append(episodes, episode)

		if err := t.cache.SetEpisode(ctx, id, seasonNumber, episodeNumber, episode); err != nil {
			logger.FromContext(ctx).With("error", err).Error("failed to cache episode")
		}

	}

	if err := t.cache.SetSeasonEpisodes(ctx, id, seasonNumber, episodes); err != nil {
		logger.FromContext(ctx).With("error", err).Error("failed to cache episode")
	}

	if episodeNumber <= 0 || episodeNumber > len(episodes) {
		return mediadata.Episode{}, fmt.Errorf("episode number %d out of range (season has %d episodes)", episodeNumber, len(episodes))
	}

	for _, episode := range episodes {
		if episode.EpisodeNumber == episodeNumber {
			return episode, nil
		}
	}

	return mediadata.Episode{}, fmt.Errorf("episode not found")
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
