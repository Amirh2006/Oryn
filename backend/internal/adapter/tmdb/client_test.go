package tmdb

import (
	"testing"

	"github.com/Amirh2006/Oryn/backend/internal/config"
	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTMDB_NormalizeItem(t *testing.T) {
	client, err := NewClient(config.TMDBConfig{
		APIKey:       "dummy-key",
		BaseURL:      "https://api.themoviedb.org/3",
		ImageBaseURL: "https://image.tmdb.org/t/p",
	}, nil)
	require.NoError(t, err)

	raw := RawTMDBItem{
		ID:               550,
		Title:            "Fight Club",
		OriginalTitle:    "Fight Club",
		Overview:         "An insomniac office worker...",
		PosterPath:       "/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg",
		BackdropPath:     "/fCayJrkYRaCRCTh8GqN30f8oyQF.jpg",
		ReleaseDate:      "1999-10-15",
		VoteAverage:      8.433,
		VoteCount:        26000,
		Popularity:       61.4,
		GenreIDs:         []int{18, 53},
		OriginalLanguage: "en",
		MediaType:        "movie",
	}

	item := client.NormalizeItem(raw, domain.MediaTypeMovie)

	assert.Equal(t, "movie-550", item.ID)
	assert.Equal(t, 550, item.ProviderID)
	assert.Equal(t, "tmdb", item.ProviderName)
	assert.Equal(t, domain.MediaTypeMovie, item.MediaType)
	assert.Equal(t, "Fight Club", item.Title)
	assert.Equal(t, 1999, item.ReleaseYear)
	assert.Equal(t, "https://image.tmdb.org/t/p/w342/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg", item.PosterURL)
	assert.Equal(t, "https://image.tmdb.org/t/p/w780/fCayJrkYRaCRCTh8GqN30f8oyQF.jpg", item.BackdropURL)
	assert.Equal(t, 8.433, item.Rating)
	assert.Equal(t, 26000, item.VoteCount)
}

func TestTMDB_NormalizeDetails_GracefulMissingFields(t *testing.T) {
	client, err := NewClient(config.TMDBConfig{
		APIKey:       "dummy-key",
		BaseURL:      "https://api.themoviedb.org/3",
		ImageBaseURL: "https://image.tmdb.org/t/p",
	}, nil)
	require.NoError(t, err)

	raw := RawTMDBDetails{
		ID:               1399,
		Name:             "Game of Thrones",
		OriginalName:     "Game of Thrones",
		Overview:         "Seven noble families fight for control...",
		FirstAirDate:     "2011-04-17",
		VoteAverage:      8.4,
		VoteCount:        22000,
		NumberOfSeasons:  8,
		NumberOfEpisodes: 73,
		Credits: &RawTMDBCredits{
			Cast: []RawTMDBCast{
				{ID: 1229, Name: "Emilia Clarke", Character: "Daenerys Targaryen", Order: 0},
			},
			Crew: []RawTMDBCrew{
				{ID: 9813, Name: "David Benioff", Job: "Director", Department: "Directing"},
			},
		},
		Videos: &RawTMDBVideosResponse{
			Results: []RawTMDBVideo{
				{Key: "KPLWWIOCOOQ", Site: "YouTube", Type: "Trailer", Official: true},
			},
		},
	}

	details := client.NormalizeDetails(raw, domain.MediaTypeTV)

	assert.Equal(t, "tv-1399", details.ID)
	assert.Equal(t, domain.MediaTypeTV, details.MediaType)
	assert.Equal(t, "Game of Thrones", details.Title)
	assert.Equal(t, 2011, details.ReleaseYear)
	assert.Equal(t, 8, details.NumberOfSeasons)
	assert.Equal(t, 73, details.NumberOfEpisodes)
	assert.Len(t, details.Cast, 1)
	assert.Equal(t, "Emilia Clarke", details.Cast[0].Name)
	assert.Len(t, details.Crew, 1)
	assert.Equal(t, "David Benioff", details.Crew[0].Name)
	assert.Equal(t, "https://www.youtube.com/watch?v=KPLWWIOCOOQ", details.TrailerURL)
}
