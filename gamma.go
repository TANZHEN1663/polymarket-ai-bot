package polymarket

import (
	"context"
	"time"
)

type GammaService struct {
	client *Client
}

func NewGammaService(client *Client) *GammaService {
	return &GammaService{client: client}
}

type Event struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Subtitle    string    `json:"subtitle"`
	Category    string    `json:"category"`
	Status      string    `json:"status"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Markets     []Market  `json:"markets"`
	ImageURL    string    `json:"image_url"`
	Description string    `json:"description"`
}

type Market struct {
	ID              string   `json:"id"`
	Slug            string   `json:"slug"`
	Title           string   `json:"title"`
	Subtitle        string   `json:"subtitle"`
	EventID         string   `json:"event_id"`
	Status          string   `json:"status"`
	OutcomeType     string   `json:"outcome_type"`
	Outcomes        []string `json:"outcomes"`
	YesPrice        float64  `json:"yes_price"`
	NoPrice         float64  `json:"no_price"`
	LastPrice       float64  `json:"last_price"`
	Volume          float64  `json:"volume"`
	OpenInterest    float64  `json:"open_interest"`
	Liquidity       float64  `json:"liquidity"`
	ExpirationTime  time.Time `json:"expiration_time"`
	ExpirationValue string   `json:"expiration_value"`
	Result          string   `json:"result"`
	CanClose        bool     `json:"can_close"`
	ImageURL        string   `json:"image_url"`
}

type Tag struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

type Series struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

type GammaComment struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
	Likes     int       `json:"likes"`
	MarketID  string    `json:"market_id"`
}

type Sport struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Categories  []string `json:"categories"`
	ImageURL    string `json:"image_url"`
}

type ListEventsParams struct {
	Category  string `json:"category,omitempty"`
	Search    string `json:"search,omitempty"`
	Status    string `json:"status,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Cursor    string `json:"cursor,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	OrderBy   string `json:"order_by,omitempty"`
}

type ListEventsResponse struct {
	Events     []Event  `json:"events"`
	NextCursor string   `json:"next_cursor,omitempty"`
	HasMore    bool     `json:"has_more,omitempty"`
}

func (s *GammaService) ListEvents(ctx context.Context, params *ListEventsParams) (*ListEventsResponse, error) {
	endpoint := "/events"
	queryParams := make(map[string]string)

	if params != nil {
		if params.Category != "" {
			queryParams["category"] = params.Category
		}
		if params.Search != "" {
			queryParams["search"] = params.Search
		}
		if params.Status != "" {
			queryParams["status"] = params.Status
		}
		if params.Limit > 0 {
			queryParams["limit"] = string(rune(params.Limit))
		}
		if params.Cursor != "" {
			queryParams["cursor"] = params.Cursor
		}
		if params.SortBy != "" {
			queryParams["sort_by"] = params.SortBy
		}
		if params.OrderBy != "" {
			queryParams["order_by"] = params.OrderBy
		}
	}

	var resp ListEventsResponse
	err := s.client.get(ctx, s.client.gammaClient, endpoint, queryParams, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *GammaService) GetEventByID(ctx context.Context, eventID string) (*Event, error) {
	endpoint := "/events/" + eventID
	var event Event
	err := s.client.get(ctx, s.client.gammaClient, endpoint, nil, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *GammaService) GetEventBySlug(ctx context.Context, slug string) (*Event, error) {
	endpoint := "/events/slug/" + slug
	var event Event
	err := s.client.get(ctx, s.client.gammaClient, endpoint, nil, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *GammaService) GetEventTags(ctx context.Context, eventID string) ([]Tag, error) {
	endpoint := "/events/" + eventID + "/tags"
	var tags []Tag
	err := s.client.get(ctx, s.client.gammaClient, endpoint, nil, &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

type ListMarketsParams struct {
	EventID     string `json:"event_id,omitempty"`
	Category    string `json:"category,omitempty"`
	Search      string `json:"search,omitempty"`
	Status      string `json:"status,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Cursor      string `json:"cursor,omitempty"`
	SortBy      string `json:"sort_by,omitempty"`
	OrderBy     string `json:"order_by,omitempty"`
}

type ListMarketsResponse struct {
	Markets    []Market `json:"markets"`
	NextCursor string   `json:"next_cursor,omitempty"`
	HasMore    bool     `json:"has_more,omitempty"`
}

func (s *GammaService) ListMarkets(ctx context.Context, params *ListMarketsParams) (*ListMarketsResponse, error) {
	endpoint := "/markets"
	queryParams := make(map[string]string)

	if params != nil {
		if params.EventID != "" {
			queryParams["event_id"] = params.EventID
		}
		if params.Category != "" {
			queryParams["category"] = params.Category
		}
		if params.Search != "" {
			queryParams["search"] = params.Search
		}
		if params.Status != "" {
			queryParams["status"] = params.Status
		}
		if params.Limit > 0 {
			queryParams["limit"] = string(rune(params.Limit))
		}
		if params.Cursor != "" {
			queryParams["cursor"] = params.Cursor
		}
		if params.SortBy != "" {
			queryParams["sort_by"] = params.SortBy
		}
		if params.OrderBy != "" {
			queryParams["order_by"] = params.OrderBy
		}
	}

	var resp ListMarketsResponse
	err := s.client.get(ctx, s.client.gammaClient, endpoint, queryParams, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *GammaService) GetMarketByID(ctx context.Context, marketID string) (*Market, error) {
	endpoint := "/markets/" + marketID
	var market Market
	err := s.client.get(ctx, s.client.gammaClient, endpoint, nil, &market)
	if err != nil {
		return nil, err
	}
	return &market, nil
}

func (s *GammaService) GetMarketBySlug(ctx context.Context, slug string) (*Market, error) {
	endpoint := "/markets/slug/" + slug
	var market Market
	err := s.client.get(ctx, s.client.gammaClient, endpoint, nil, &market)
	if err != nil {
		return nil, err
	}
	return &market, nil
}

func (s *GammaService) GetMarketTagsByID(ctx context.Context, marketID string) ([]Tag, error) {
	endpoint := "/markets/" + marketID + "/tags"
	var tags []Tag
	err := s.client.get(ctx, s.client.gammaClient, endpoint, nil, &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (s *GammaService) GetTopHolders(ctx context.Context, marketIDs []string) (map[string][]string, error) {
	endpoint := "/markets/top-holders"
	params := map[string]interface{}{
		"market_ids": marketIDs,
	}
	var result map[string][]string
	err := s.client.post(ctx, s.client.gammaClient, endpoint, params, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *GammaService) GetOpenInterest(ctx context.Context, marketID string) (float64, error) {
	endpoint := "/markets/" + marketID + "/open-interest"
	var result struct {
		OpenInterest float64 `json:"open_interest"`
	}
	err := s.client.get(ctx, s.client.gammaClient, endpoint, nil, &result)
	if err != nil {
		return 0, err
	}
	return result.OpenInterest, nil
}

func (s *GammaService) GetLiveVolume(ctx context.Context, eventID string) (float64, error) {
	endpoint := "/events/" + eventID + "/volume"
	var result struct {
		Volume float64 `json:"volume"`
	}
	err := s.client.get(ctx, s.client.gammaClient, endpoint, nil, &result)
	if err != nil {
		return 0, err
	}
	return result.Volume, nil
}

func (s *GammaService) ListTags(ctx context.Context) ([]Tag, error) {
	endpoint := "/tags"
	var tags []Tag
	err := s.client.get(ctx, s.client.gammaClient, endpoint, nil, &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (s *GammaService) GetTagByID(ctx context.Context, tagID string) (*Tag, error) {
	endpoint := "/tags/" + tagID
	var tag Tag
	err := s.client.get(ctx, s.client.gammaClient, endpoint, nil, &tag)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (s *GammaService) GetTagBySlug(ctx context.Context, slug string) (*Tag, error) {
	endpoint := "/tags/slug/" + slug
	var tag Tag
	err := s.client.get(ctx, s.client.gammaClient, endpoint, nil, &tag)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (s *GammaService) ListComments(ctx context.Context, marketID string, limit int) ([]GammaComment, error) {
	endpoint := "/comments"
	params := map[string]string{
		"market_id": marketID,
	}
	if limit > 0 {
		params["limit"] = string(rune(limit))
	}
	var comments []GammaComment
	err := s.client.get(ctx, s.client.gammaClient, endpoint, params, &comments)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (s *GammaService) GetSports(ctx context.Context) ([]Sport, error) {
	endpoint := "/sports"
	var sports []Sport
	err := s.client.get(ctx, s.client.gammaClient, endpoint, nil, &sports)
	if err != nil {
		return nil, err
	}
	return sports, nil
}

func (s *GammaService) GetProfile(ctx context.Context, walletAddress string) (interface{}, error) {
	endpoint := "/profile/" + walletAddress
	var profile interface{}
	err := s.client.get(ctx, s.client.gammaClient, endpoint, nil, &profile)
	if err != nil {
		return nil, err
	}
	return profile, nil
}
