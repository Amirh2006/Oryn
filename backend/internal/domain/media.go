package domain

import (
	"context"
)

// MediaType represents the type of media item
type MediaType string

const (
	MediaTypeMovie MediaType = "movie"
	MediaTypeTV    MediaType = "tv"
	MediaTypeAll   MediaType = "all"
)

// Genre represents a media category
type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// MediaItem represents a normalized lightweight media item for lists, cards, and feeds
type MediaItem struct {
	ID               string    `json:"id"`
	ProviderID       int       `json:"providerId"`
	ProviderName     string    `json:"providerName"`
	MediaType        MediaType `json:"mediaType"`
	Title            string    `json:"title"`
	OriginalTitle    string    `json:"originalTitle,omitempty"`
	Overview         string    `json:"overview"`
	PosterURL        string    `json:"posterUrl,omitempty"`
	BackdropURL      string    `json:"backdropUrl,omitempty"`
	ReleaseDate      string    `json:"releaseDate,omitempty"`
	ReleaseYear      int       `json:"releaseYear,omitempty"`
	Rating           float64   `json:"rating"`
	VoteCount        int       `json:"voteCount"`
	Popularity       float64   `json:"popularity"`
	GenreIDs         []int     `json:"genreIds,omitempty"`
	Genres           []Genre   `json:"genres,omitempty"`
	OriginalLanguage string    `json:"originalLanguage,omitempty"`
}

// CastMember represents an actor/performer
type CastMember struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Character  string `json:"character"`
	ProfileURL string `json:"profileUrl,omitempty"`
	Order      int    `json:"order"`
}

// CrewMember represents a director, writer, or key crew member
type CrewMember struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Job        string `json:"job"`
	Department string `json:"department"`
	ProfileURL string `json:"profileUrl,omitempty"`
}

// MediaDetails represents rich detailed media information for single item views
type MediaDetails struct {
	MediaItem
	Tagline             string       `json:"tagline,omitempty"`
	Runtime             int          `json:"runtime,omitempty"` // runtime in minutes (for movies)
	NumberOfSeasons     int          `json:"numberOfSeasons,omitempty"` // for TV
	NumberOfEpisodes    int          `json:"numberOfEpisodes,omitempty"` // for TV
	Status              string       `json:"status,omitempty"`
	Budget              int64        `json:"budget,omitempty"`
	Revenue             int64        `json:"revenue,omitempty"`
	ProductionCountries []string     `json:"productionCountries,omitempty"`
	SpokenLanguages     []string     `json:"spokenLanguages,omitempty"`
	TrailerURL          string       `json:"trailerUrl,omitempty"`
	Cast                []CastMember `json:"cast,omitempty"`
	Crew                []CrewMember `json:"crew,omitempty"`
}

// MediaFilter specifies search and discovery parameters
type MediaFilter struct {
	MediaType        MediaType `json:"mediaType"`
	Genres           []int     `json:"genres,omitempty"`
	ReleaseYear      int       `json:"releaseYear,omitempty"`
	RatingMin        float64   `json:"ratingMin,omitempty"`
	RatingMax        float64   `json:"ratingMax,omitempty"`
	RuntimeMin       int       `json:"runtimeMin,omitempty"`
	RuntimeMax       int       `json:"runtimeMax,omitempty"`
	OriginalLanguage string    `json:"originalLanguage,omitempty"`
	SortBy           string    `json:"sortBy,omitempty"` // popularity.desc, vote_average.desc, primary_release_date.desc
}

// PaginatedResult is a generic pagination container
type PaginatedResult[T any] struct {
	Items        []T   `json:"items"`
	Page         int   `json:"page"`
	TotalPages   int   `json:"totalPages"`
	TotalResults int64 `json:"totalResults"`
}

// MediaProvider defines the interface for interacting with media catalog APIs
type MediaProvider interface {
	GetTrending(ctx context.Context, mediaType MediaType, timeWindow string, page int) (*PaginatedResult[MediaItem], error)
	GetPopular(ctx context.Context, mediaType MediaType, page int) (*PaginatedResult[MediaItem], error)
	GetTopRated(ctx context.Context, mediaType MediaType, page int) (*PaginatedResult[MediaItem], error)
	SearchMedia(ctx context.Context, query string, mediaType MediaType, page int) (*PaginatedResult[MediaItem], error)
	DiscoverMedia(ctx context.Context, filter MediaFilter, page int) (*PaginatedResult[MediaItem], error)
	GetMediaDetails(ctx context.Context, mediaType MediaType, id string) (*MediaDetails, error)
	GetSimilarMedia(ctx context.Context, mediaType MediaType, id string, page int) (*PaginatedResult[MediaItem], error)
	GetGenres(ctx context.Context, mediaType MediaType) ([]Genre, error)
}
