package mediadata

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
	SearchMovie(query string, page int) (MovieResults, error)
	GetMovie(id string) (Movie, error)
	GetMovieDetails(id string) (MovieDetails, error)
}

type TvShowClient interface {
	SearchTvShow(query string, page int) (TvShowResults, error)
	GetTvShow(id string) (TvShow, error)
	GetTvShowDetails(id string) (TvShowDetails, error)
}
