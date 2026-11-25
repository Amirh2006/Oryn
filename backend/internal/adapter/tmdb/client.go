package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Amirh2006/Oryn/backend/internal/config"
	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/Amirh2006/Oryn/backend/internal/platform/resilience"
)

const (
	providerName    = "tmdb"
	defaultMaxRetry = 3
)

// Client is a resilient TMDB API client implementing domain.MediaProvider
type Client struct {
	apiKey         string
	baseURL        string
	imageBaseURL   string
	httpClient     *http.Client
	circuitBreaker *resilience.CircuitBreaker
}

// NewClient constructs a new TMDB client
func NewClient(cfg config.TMDBConfig, cb *resilience.CircuitBreaker) (*Client, error) {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	if cfg.OutboundProxyURL != "" {
		proxyURL, err := url.Parse(cfg.OutboundProxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid OUTBOUND_PROXY_URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
		slog.Info("TMDB client configured with outbound proxy", slog.String("proxy", cfg.OutboundProxyURL))
	}

	if cb == nil {
		cb = resilience.NewCircuitBreaker(resilience.CircuitBreakerConfig{
			MaxFailures:       5,
			ResetTimeout:      30 * time.Second,
			HalfOpenSuccesses: 2,
		})
	}

	return &Client{
		apiKey:       cfg.APIKey,
		baseURL:      strings.TrimRight(cfg.BaseURL, "/"),
		imageBaseURL: strings.TrimRight(cfg.ImageBaseURL, "/"),
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   15 * time.Second,
		},
		circuitBreaker: cb,
	}, nil
}

// doRequest performs an HTTP request with exponential backoff and circuit breaker
func (c *Client) doRequest(ctx context.Context, endpoint string, query url.Values) ([]byte, error) {
	if c.apiKey == "" {
		return nil, domain.NewAppError(domain.ErrCodeExternalProvider, "TMDB_API_KEY is not configured", nil)
	}

	query.Set("api_key", c.apiKey)
	fullURL := fmt.Sprintf("%s%s?%s", c.baseURL, endpoint, query.Encode())

	var body []byte
	op := func() error {
		var lastErr error
		for attempt := 1; attempt <= defaultMaxRetry; attempt++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
			if err != nil {
				return err
			}
			req.Header.Set("Accept", "application/json")

			resp, err := c.httpClient.Do(req)
			if err != nil {
				lastErr = err
				c.backoff(ctx, attempt)
				continue
			}

			respBody, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				lastErr = err
				c.backoff(ctx, attempt)
				continue
			}

			// Transient rate limits or upstream 5xx errors -> retry
			if resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode >= 500 && resp.StatusCode <= 504) {
				lastErr = fmt.Errorf("upstream TMDB transient error: HTTP %d", resp.StatusCode)
				c.backoff(ctx, attempt)
				continue
			}

			if resp.StatusCode == http.StatusNotFound {
				return domain.NewAppError(domain.ErrCodeNotFound, "media not found on upstream provider", nil)
			}

			if resp.StatusCode >= 400 {
				return domain.NewAppError(domain.ErrCodeExternalProvider, fmt.Sprintf("TMDB API returned HTTP %d: %s", resp.StatusCode, string(respBody)), nil)
			}

			body = respBody
			return nil
		}
		return domain.WrapError(domain.ErrCodeExternalProvider, "TMDB request failed after max retries", lastErr)
	}

	if err := c.circuitBreaker.Execute(op); err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) backoff(ctx context.Context, attempt int) {
	delay := time.Duration(100*(1<<attempt))*time.Millisecond + time.Duration(rand.Intn(50))*time.Millisecond
	select {
	case <-time.After(delay):
	case <-ctx.Done():
	}
}

// BuildImageURL resolves a poster or backdrop path to a full CDN URL
func (c *Client) BuildImageURL(size, path string) string {
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return fmt.Sprintf("%s/%s%s", c.imageBaseURL, size, path)
}

// NormalizeItem maps a raw TMDB item into internal domain.MediaItem
func (c *Client) NormalizeItem(raw RawTMDBItem, defaultType domain.MediaType) domain.MediaItem {
	mediaType := defaultType
	if raw.MediaType == "movie" {
		mediaType = domain.MediaTypeMovie
	} else if raw.MediaType == "tv" {
		mediaType = domain.MediaTypeTV
	}

	title := raw.Title
	if title == "" {
		title = raw.Name
	}

	origTitle := raw.OriginalTitle
	if origTitle == "" {
		origTitle = raw.OriginalName
	}

	dateStr := raw.ReleaseDate
	if dateStr == "" {
		dateStr = raw.FirstAirDate
	}

	year := 0
	if len(dateStr) >= 4 {
		if y, err := strconv.Atoi(dateStr[:4]); err == nil {
			year = y
		}
	}

	return domain.MediaItem{
		ID:               fmt.Sprintf("%s-%d", mediaType, raw.ID),
		ProviderID:       raw.ID,
		ProviderName:     providerName,
		MediaType:        mediaType,
		Title:            title,
		OriginalTitle:    origTitle,
		Overview:         raw.Overview,
		PosterURL:        c.BuildImageURL("w342", raw.PosterPath),
		BackdropURL:      c.BuildImageURL("w780", raw.BackdropPath),
		ReleaseDate:      dateStr,
		ReleaseYear:      year,
		Rating:           raw.VoteAverage,
		VoteCount:        raw.VoteCount,
		Popularity:       raw.Popularity,
		GenreIDs:         raw.GenreIDs,
		OriginalLanguage: raw.OriginalLanguage,
	}
}

// NormalizeDetails maps raw details to domain.MediaDetails
func (c *Client) NormalizeDetails(raw RawTMDBDetails, mediaType domain.MediaType) domain.MediaDetails {
	item := c.NormalizeItem(RawTMDBItem{
		ID:               raw.ID,
		Title:            raw.Title,
		Name:             raw.Name,
		OriginalTitle:    raw.OriginalTitle,
		OriginalName:     raw.OriginalName,
		Overview:         raw.Overview,
		PosterPath:       raw.PosterPath,
		BackdropPath:     raw.BackdropPath,
		ReleaseDate:      raw.ReleaseDate,
		FirstAirDate:     raw.FirstAirDate,
		VoteAverage:      raw.VoteAverage,
		VoteCount:        raw.VoteCount,
		Popularity:       raw.Popularity,
		OriginalLanguage: raw.OriginalLanguage,
	}, mediaType)

	var genres []domain.Genre
	for _, g := range raw.Genres {
		genres = append(genres, domain.Genre{ID: g.ID, Name: g.Name})
	}
	item.Genres = genres

	var countries []string
	for _, pc := range raw.ProductionCountries {
		countries = append(countries, pc.Name)
	}

	var languages []string
	for _, sl := range raw.SpokenLanguages {
		languages = append(languages, sl.EnglishName)
	}

	var cast []domain.CastMember
	var crew []domain.CrewMember
	if raw.Credits != nil {
		for i, cm := range raw.Credits.Cast {
			if i >= 15 {
				break
			}
			cast = append(cast, domain.CastMember{
				ID:         cm.ID,
				Name:       cm.Name,
				Character:  cm.Character,
				ProfileURL: c.BuildImageURL("w185", cm.ProfilePath),
				Order:      cm.Order,
			})
		}
		for _, cr := range raw.Credits.Crew {
			if cr.Job == "Director" || cr.Job == "Writer" || cr.Job == "Screenplay" || cr.Job == "Producer" {
				crew = append(crew, domain.CrewMember{
					ID:         cr.ID,
					Name:       cr.Name,
					Job:        cr.Job,
					Department: cr.Department,
					ProfileURL: c.BuildImageURL("w185", cr.ProfilePath),
				})
			}
		}
	}

	trailerURL := ""
	if raw.Videos != nil {
		for _, v := range raw.Videos.Results {
			if v.Site == "YouTube" && (v.Type == "Trailer" || v.Type == "Teaser") {
				trailerURL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.Key)
				if v.Official {
					break
				}
			}
		}
	}

	return domain.MediaDetails{
		MediaItem:           item,
		Tagline:             raw.Tagline,
		Runtime:             raw.Runtime,
		NumberOfSeasons:     raw.NumberOfSeasons,
		NumberOfEpisodes:    raw.NumberOfEpisodes,
		Status:              raw.Status,
		Budget:              raw.Budget,
		Revenue:             raw.Revenue,
		ProductionCountries: countries,
		SpokenLanguages:     languages,
		TrailerURL:          trailerURL,
		Cast:                cast,
		Crew:                crew,
	}
}

// GetTrending fetches trending movies or TV shows
func (c *Client) GetTrending(ctx context.Context, mediaType domain.MediaType, timeWindow string, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	if timeWindow != "day" && timeWindow != "week" {
		timeWindow = "week"
	}
	endpointType := "all"
	if mediaType == domain.MediaTypeMovie {
		endpointType = "movie"
	} else if mediaType == domain.MediaTypeTV {
		endpointType = "tv"
	}

	endpoint := fmt.Sprintf("/trending/%s/%s", endpointType, timeWindow)
	query := url.Values{"page": []string{strconv.Itoa(page)}}

	body, err := c.doRequest(ctx, endpoint, query)
	if err != nil {
		return nil, err
	}

	var raw RawTMDBPaginatedResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, domain.WrapError(domain.ErrCodeExternalProvider, "failed to parse TMDB trending response", err)
	}

	var items []domain.MediaItem
	for _, itm := range raw.Results {
		items = append(items, c.NormalizeItem(itm, mediaType))
	}

	return &domain.PaginatedResult[domain.MediaItem]{
		Items:        items,
		Page:         raw.Page,
		TotalPages:   raw.TotalPages,
		TotalResults: raw.TotalResults,
	}, nil
}

// GetPopular fetches popular items
func (c *Client) GetPopular(ctx context.Context, mediaType domain.MediaType, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	endpoint := "/movie/popular"
	if mediaType == domain.MediaTypeTV {
		endpoint = "/tv/popular"
	}

	query := url.Values{"page": []string{strconv.Itoa(page)}}
	body, err := c.doRequest(ctx, endpoint, query)
	if err != nil {
		return nil, err
	}

	var raw RawTMDBPaginatedResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, domain.WrapError(domain.ErrCodeExternalProvider, "failed to parse TMDB popular response", err)
	}

	var items []domain.MediaItem
	for _, itm := range raw.Results {
		items = append(items, c.NormalizeItem(itm, mediaType))
	}

	return &domain.PaginatedResult[domain.MediaItem]{
		Items:        items,
		Page:         raw.Page,
		TotalPages:   raw.TotalPages,
		TotalResults: raw.TotalResults,
	}, nil
}

// GetTopRated fetches top rated items
func (c *Client) GetTopRated(ctx context.Context, mediaType domain.MediaType, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	endpoint := "/movie/top_rated"
	if mediaType == domain.MediaTypeTV {
		endpoint = "/tv/top_rated"
	}

	query := url.Values{"page": []string{strconv.Itoa(page)}}
	body, err := c.doRequest(ctx, endpoint, query)
	if err != nil {
		return nil, err
	}

	var raw RawTMDBPaginatedResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, domain.WrapError(domain.ErrCodeExternalProvider, "failed to parse TMDB top-rated response", err)
	}

	var items []domain.MediaItem
	for _, itm := range raw.Results {
		items = append(items, c.NormalizeItem(itm, mediaType))
	}

	return &domain.PaginatedResult[domain.MediaItem]{
		Items:        items,
		Page:         raw.Page,
		TotalPages:   raw.TotalPages,
		TotalResults: raw.TotalResults,
	}, nil
}

// SearchMedia searches movies or series by text query
func (c *Client) SearchMedia(ctx context.Context, q string, mediaType domain.MediaType, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	endpoint := "/search/multi"
	if mediaType == domain.MediaTypeMovie {
		endpoint = "/search/movie"
	} else if mediaType == domain.MediaTypeTV {
		endpoint = "/search/tv"
	}

	query := url.Values{
		"query": []string{q},
		"page":  []string{strconv.Itoa(page)},
	}

	body, err := c.doRequest(ctx, endpoint, query)
	if err != nil {
		return nil, err
	}

	var raw RawTMDBPaginatedResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, domain.WrapError(domain.ErrCodeExternalProvider, "failed to parse TMDB search response", err)
	}

	var items []domain.MediaItem
	for _, itm := range raw.Results {
		// In multi search, ignore person results
		if itm.MediaType == "person" {
			continue
		}
		items = append(items, c.NormalizeItem(itm, mediaType))
	}

	return &domain.PaginatedResult[domain.MediaItem]{
		Items:        items,
		Page:         raw.Page,
		TotalPages:   raw.TotalPages,
		TotalResults: raw.TotalResults,
	}, nil
}

// DiscoverMedia supports complex multi-criteria filtering
func (c *Client) DiscoverMedia(ctx context.Context, f domain.MediaFilter, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	endpoint := "/discover/movie"
	if f.MediaType == domain.MediaTypeTV {
		endpoint = "/discover/tv"
	}

	query := url.Values{
		"page": []string{strconv.Itoa(page)},
	}

	if f.SortBy != "" {
		query.Set("sort_by", f.SortBy)
	} else {
		query.Set("sort_by", "popularity.desc")
	}

	if len(f.Genres) > 0 {
		var genreStrs []string
		for _, g := range f.Genres {
			genreStrs = append(genreStrs, strconv.Itoa(g))
		}
		query.Set("with_genres", strings.Join(genreStrs, ","))
	}

	if f.ReleaseYear > 0 {
		if f.MediaType == domain.MediaTypeTV {
			query.Set("first_air_date_year", strconv.Itoa(f.ReleaseYear))
		} else {
			query.Set("primary_release_year", strconv.Itoa(f.ReleaseYear))
		}
	}

	if f.RatingMin > 0 {
		query.Set("vote_average.gte", fmt.Sprintf("%.1f", f.RatingMin))
	}
	if f.RatingMax > 0 {
		query.Set("vote_average.lte", fmt.Sprintf("%.1f", f.RatingMax))
	}
	if f.RuntimeMin > 0 {
		query.Set("with_runtime.gte", strconv.Itoa(f.RuntimeMin))
	}
	if f.RuntimeMax > 0 {
		query.Set("with_runtime.lte", strconv.Itoa(f.RuntimeMax))
	}
	if f.OriginalLanguage != "" {
		query.Set("with_original_language", f.OriginalLanguage)
	}

	body, err := c.doRequest(ctx, endpoint, query)
	if err != nil {
		return nil, err
	}

	var raw RawTMDBPaginatedResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, domain.WrapError(domain.ErrCodeExternalProvider, "failed to parse TMDB discover response", err)
	}

	var items []domain.MediaItem
	for _, itm := range raw.Results {
		items = append(items, c.NormalizeItem(itm, f.MediaType))
	}

	return &domain.PaginatedResult[domain.MediaItem]{
		Items:        items,
		Page:         raw.Page,
		TotalPages:   raw.TotalPages,
		TotalResults: raw.TotalResults,
	}, nil
}

// GetMediaDetails retrieves rich details including credits and videos
func (c *Client) GetMediaDetails(ctx context.Context, mediaType domain.MediaType, id string) (*domain.MediaDetails, error) {
	// Strip prefix if given as "movie-123"
	cleanID := id
	if strings.Contains(id, "-") {
		parts := strings.Split(id, "-")
		cleanID = parts[len(parts)-1]
	}

	endpoint := fmt.Sprintf("/movie/%s", cleanID)
	if mediaType == domain.MediaTypeTV {
		endpoint = fmt.Sprintf("/tv/%s", cleanID)
	}

	query := url.Values{
		"append_to_response": []string{"credits,videos"},
	}

	body, err := c.doRequest(ctx, endpoint, query)
	if err != nil {
		return nil, err
	}

	var raw RawTMDBDetails
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, domain.WrapError(domain.ErrCodeExternalProvider, "failed to parse TMDB details response", err)
	}

	details := c.NormalizeDetails(raw, mediaType)
	return &details, nil
}

// GetSimilarMedia finds similar titles
func (c *Client) GetSimilarMedia(ctx context.Context, mediaType domain.MediaType, id string, page int) (*domain.PaginatedResult[domain.MediaItem], error) {
	cleanID := id
	if strings.Contains(id, "-") {
		parts := strings.Split(id, "-")
		cleanID = parts[len(parts)-1]
	}

	endpoint := fmt.Sprintf("/movie/%s/similar", cleanID)
	if mediaType == domain.MediaTypeTV {
		endpoint = fmt.Sprintf("/tv/%s/similar", cleanID)
	}

	query := url.Values{"page": []string{strconv.Itoa(page)}}
	body, err := c.doRequest(ctx, endpoint, query)
	if err != nil {
		return nil, err
	}

	var raw RawTMDBPaginatedResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, domain.WrapError(domain.ErrCodeExternalProvider, "failed to parse TMDB similar response", err)
	}

	var items []domain.MediaItem
	for _, itm := range raw.Results {
		items = append(items, c.NormalizeItem(itm, mediaType))
	}

	return &domain.PaginatedResult[domain.MediaItem]{
		Items:        items,
		Page:         raw.Page,
		TotalPages:   raw.TotalPages,
		TotalResults: raw.TotalResults,
	}, nil
}

// GetGenres retrieves official genre lists
func (c *Client) GetGenres(ctx context.Context, mediaType domain.MediaType) ([]domain.Genre, error) {
	endpoint := "/genre/movie/list"
	if mediaType == domain.MediaTypeTV {
		endpoint = "/genre/tv/list"
	}

	body, err := c.doRequest(ctx, endpoint, url.Values{})
	if err != nil {
		return nil, err
	}

	var raw RawTMDBGenreListResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, domain.WrapError(domain.ErrCodeExternalProvider, "failed to parse TMDB genres response", err)
	}

	var genres []domain.Genre
	for _, g := range raw.Genres {
		genres = append(genres, domain.Genre{ID: g.ID, Name: g.Name})
	}

	return genres, nil
}
