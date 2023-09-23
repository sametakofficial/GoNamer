package mediadata

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

type TvShow struct {
	ID          string
	Title       string
	Overview    string
	FistAirDate string
	PosterURL   string
	Rating      float32
	RatingCount int64
}

type TvShowDetails struct {
	TvShow
	Genres []Genre
	Cast   []Person
	Studio []Studio
}

type TvShowResults struct {
	TvShows        []TvShow
	Totals         int64
	ResultsPerPage int64
}

type MovieClient interface {
	SearchMovie(query string, page int) (MovieResults, error)
	GetMovie(id string) (Movie, error)
}

type TvShowClient interface {
	SearchTvShow(query string, page int) (TvShowResults, error)
	GetTvShow(id string) (TvShow, error)
}
