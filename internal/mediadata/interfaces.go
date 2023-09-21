package mediadata

type Movie struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Overview    string `json:"overview"`
	ReleaseDate string `json:"release_date"`
}

type MovieResults struct {
	Movies         []Movie `json:"movies"`
	Totals         int64   `json:"totals"`
	ResultsPerPage int64   `json:"results_per_page"`
}

type MovieClient interface {
	SearchMovie(query string, page int) (MovieResults, error)
	GetMovie(id int) (Movie, error)
}
