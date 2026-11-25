package tmdb

// RawTMDBItem represents a raw movie or TV item from TMDB API
type RawTMDBItem struct {
	ID               int      `json:"id"`
	Title            string   `json:"title"`             // for movies
	Name             string   `json:"name"`              // for TV series
	OriginalTitle    string   `json:"original_title"`    // for movies
	OriginalName     string   `json:"original_name"`     // for TV series
	Overview         string   `json:"overview"`
	PosterPath       string   `json:"poster_path"`
	BackdropPath     string   `json:"backdrop_path"`
	ReleaseDate      string   `json:"release_date"`      // for movies
	FirstAirDate     string   `json:"first_air_date"`    // for TV series
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
	Popularity       float64  `json:"popularity"`
	GenreIDs         []int    `json:"genre_ids"`
	OriginalLanguage string   `json:"original_language"`
	MediaType        string   `json:"media_type"`
}

// RawTMDBPaginatedResponse represents a paginated TMDB response
type RawTMDBPaginatedResponse struct {
	Page         int           `json:"page"`
	Results      []RawTMDBItem `json:"results"`
	TotalPages   int           `json:"total_pages"`
	TotalResults int64         `json:"total_results"`
}

// RawTMDBGenre represents a genre object
type RawTMDBGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type RawTMDBGenreListResponse struct {
	Genres []RawTMDBGenre `json:"genres"`
}

// RawTMDBCast represents cast member
type RawTMDBCast struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
	Order       int    `json:"order"`
}

// RawTMDBCrew represents crew member
type RawTMDBCrew struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Job         string `json:"job"`
	Department  string `json:"department"`
	ProfilePath string `json:"profile_path"`
}

// RawTMDBCredits represents credits payload
type RawTMDBCredits struct {
	Cast []RawTMDBCast `json:"cast"`
	Crew []RawTMDBCrew `json:"crew"`
}

// RawTMDBVideo represents a video/trailer entry
type RawTMDBVideo struct {
	Key     string `json:"key"`
	Site    string `json:"site"`
	Type    string `json:"type"`
	Official bool  `json:"official"`
}

type RawTMDBVideosResponse struct {
	Results []RawTMDBVideo `json:"results"`
}

// RawTMDBProductionCountry represents production country
type RawTMDBProductionCountry struct {
	Iso31661 string `json:"iso_3166_1"`
	Name     string `json:"name"`
}

// RawTMDBSpokenLanguage represents spoken language
type RawTMDBSpokenLanguage struct {
	EnglishName string `json:"english_name"`
	Iso6391     string `json:"iso_639_1"`
	Name        string `json:"name"`
}

// RawTMDBDetails represents comprehensive media details
type RawTMDBDetails struct {
	ID                  int                        `json:"id"`
	Title               string                     `json:"title"`
	Name                string                     `json:"name"`
	OriginalTitle       string                     `json:"original_title"`
	OriginalName        string                     `json:"original_name"`
	Overview            string                     `json:"overview"`
	Tagline             string                     `json:"tagline"`
	PosterPath          string                     `json:"poster_path"`
	BackdropPath        string                     `json:"backdrop_path"`
	ReleaseDate         string                     `json:"release_date"`
	FirstAirDate        string                     `json:"first_air_date"`
	VoteAverage         float64                    `json:"vote_average"`
	VoteCount           int                        `json:"vote_count"`
	Popularity          float64                    `json:"popularity"`
	Genres              []RawTMDBGenre             `json:"genres"`
	OriginalLanguage    string                     `json:"original_language"`
	Status              string                     `json:"status"`
	Runtime             int                        `json:"runtime"`               // Movies
	NumberOfSeasons     int                        `json:"number_of_seasons"`     // TV
	NumberOfEpisodes    int                        `json:"number_of_episodes"`    // TV
	Budget              int64                      `json:"budget"`
	Revenue             int64                      `json:"revenue"`
	ProductionCountries []RawTMDBProductionCountry `json:"production_countries"`
	SpokenLanguages     []RawTMDBSpokenLanguage    `json:"spoken_languages"`
	Credits             *RawTMDBCredits            `json:"credits,omitempty"`
	Videos              *RawTMDBVideosResponse     `json:"videos,omitempty"`
}
