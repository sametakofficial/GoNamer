package tmdb

import (
	"github.com/cyruzin/golang-tmdb"
	"github.com/nouuu/mediatracker/internal/mediadata"
	"log/slog"
	"strconv"
)

func NewTvShowClient(APIKey string, opts ...OptFunc) mediadata.TvShowClient {
	o := defaultOpts(APIKey)
	for _, optF := range opts {
		optF(&o.Opts)
	}

	client, err := tmdb.Init(o.APIKey)
	if err != nil {
		slog.Error("Failed to initialize TMDB client", slog.Any("error", err))
	}
	return &tmdbClient{client: client, opts: o}
}

func (t *tmdbClient) SearchTvShow(query string, page int) (mediadata.TvShowResults, error) {
	searchTvShows, err := t.client.GetSearchTVShow(query, cfgMap(t.opts, map[string]string{
		"page": strconv.Itoa(page),
	}))
	if err != nil {
		return mediadata.TvShowResults{}, err
	}
	var tvShows []mediadata.TvShow = buildTvShowFromResult(searchTvShows.SearchTVShowsResults)
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

func buildTvShow(tvShow *tmdb.TVDetails) mediadata.TvShow {
	return mediadata.TvShow{
		ID:          strconv.FormatInt(tvShow.ID, 10),
		Title:       tvShow.Name,
		Overview:    tvShow.Overview,
		FistAirDate: tvShow.FirstAirDate,
		PosterURL:   tmdbImageBaseUrl + tvShow.PosterPath,
		Rating:      tvShow.VoteAverage,
		RatingCount: tvShow.VoteCount,
	}
}

func buildTvShowDetails(details *tmdb.TVDetails) mediadata.TvShowDetails {
	return mediadata.TvShowDetails{
		TvShow: mediadata.TvShow{
			ID:          strconv.FormatInt(details.ID, 10),
			Title:       details.Name,
			Overview:    details.Overview,
			FistAirDate: details.FirstAirDate,
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
	AirDate      string `json:"air_date"`
	EpisodeCount int    `json:"episode_count"`
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
	SeasonNumber int    `json:"season_number"`
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
