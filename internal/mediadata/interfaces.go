package mediadata

type Movie struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Overview    string `json:"overview"`
	ReleaseDate string `json:"release_date"`
	PosterURL   string `json:"poster_url"`
}

type MovieResults struct {
	Movies         []Movie `json:"movies"`
	Totals         int64   `json:"totals"`
	ResultsPerPage int64   `json:"results_per_page"`
}

type TvShow struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Overview    string `json:"overview"`
	FistAirDate string `json:"first_air_date"`
	PosterURL   string `json:"poster_url"`
}

type TvShowResults struct {
	TvShows        []TvShow `json:"tv_shows"`
	Totals         int64    `json:"totals"`
	ResultsPerPage int64    `json:"results_per_page"`
}

type MovieClient interface {
	SearchMovie(query string, page int) (MovieResults, error)
	GetMovie(id string) (Movie, error)
}

type TvShowClient interface {
	SearchTvShow(query string, page int) (TvShowResults, error)
	GetTvShow(id string) (TvShow, error)
}
