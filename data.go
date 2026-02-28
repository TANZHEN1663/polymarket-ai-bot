package polymarket

import (
	"context"
	"time"
)

type DataService struct {
	client *Client
}

func NewDataService(client *Client) *DataService {
	return &DataService{client: client}
}

type UserPosition struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	MarketID       string    `json:"market_id"`
	MarketTitle    string    `json:"market_title"`
	Outcome        string    `json:"outcome"`
	Position       float64   `json:"position"`
	AveragePrice   float64   `json:"average_price"`
	TotalCost      float64   `json:"total_cost"`
	CurrentValue   float64   `json:"current_value"`
	PnL            float64   `json:"pnl"`
	PnLPercent     float64   `json:"pnl_percent"`
	RealizedPnL    float64   `json:"realized_pnl"`
	LastUpdated    time.Time `json:"last_updated"`
	IsClosed       bool      `json:"is_closed"`
	ClosedAt       time.Time `json:"closed_at,omitempty"`
}

type Trade struct {
	ID          string    `json:"id"`
	MarketID    string    `json:"market_id"`
	MarketTitle string    `json:"market_title"`
	UserID      string    `json:"user_id"`
	Side        string    `json:"side"`
	Outcome     string    `json:"outcome"`
	Count       float64   `json:"count"`
	Price       float64   `json:"price"`
	Fees        float64   `json:"fees"`
	Timestamp   time.Time `json:"timestamp"`
	TxHash      string    `json:"tx_hash"`
}

type UserActivity struct {
	UserID              string    `json:"user_id"`
	TotalTrades         int       `json:"total_trades"`
	TotalVolume         float64   `json:"total_volume"`
	TotalMarkets        int       `json:"total_markets"`
	ActivePositions     int       `json:"active_positions"`
	ClosedPositions     int       `json:"closed_positions"`
	TotalPnL            float64   `json:"total_pnl"`
	WinRate             float64   `json:"win_rate"`
	LastActivity        time.Time `json:"last_activity"`
	AccountCreated      time.Time `json:"account_created"`
}

type HolderData struct {
	Address        string  `json:"address"`
	Position       float64 `json:"position"`
	Percentage     float64 `json:"percentage"`
	AveragePrice   float64 `json:"average_price"`
	CurrentValue   float64 `json:"current_value"`
}

type OpenInterestData struct {
	MarketID     string    `json:"market_id"`
	OpenInterest float64   `json:"open_interest"`
	YesOI        float64   `json:"yes_oi"`
	NoOI         float64   `json:"no_oi"`
	LastUpdated  time.Time `json:"last_updated"`
}

type LeaderboardEntry struct {
	Rank        int     `json:"rank"`
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	TotalPnL    float64 `json:"total_pnl"`
	WinRate     float64 `json:"win_rate"`
	TotalVolume float64 `json:"total_volume"`
	TradeCount  int     `json:"trade_count"`
}

type BuilderAnalytics struct {
	BuilderID       string    `json:"builder_id"`
	BuilderName     string    `json:"builder_name"`
	TotalMarkets    int       `json:"total_markets"`
	TotalVolume     float64   `json:"total_volume"`
	TotalFees       float64   `json:"total_fees"`
	ActiveMarkets   int       `json:"active_markets"`
	MarketShare     float64   `json:"market_share"`
	LastUpdated     time.Time `json:"last_updated"`
}

type GetUserPositionsParams struct {
	UserID     string `json:"user_id,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
	IsClosed   *bool  `json:"is_closed,omitempty"`
	MarketID   string `json:"market_id,omitempty"`
}

type GetUserPositionsResponse struct {
	Positions  []UserPosition `json:"positions"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more,omitempty"`
}

func (s *DataService) GetUserPositions(ctx context.Context, params *GetUserPositionsParams) (*GetUserPositionsResponse, error) {
	endpoint := "/positions"
	queryParams := make(map[string]string)

	if params != nil {
		if params.UserID != "" {
			queryParams["user_id"] = params.UserID
		}
		if params.Limit > 0 {
			queryParams["limit"] = string(rune(params.Limit))
		}
		if params.Cursor != "" {
			queryParams["cursor"] = params.Cursor
		}
		if params.IsClosed != nil {
			if *params.IsClosed {
				queryParams["is_closed"] = "true"
			} else {
				queryParams["is_closed"] = "false"
			}
		}
		if params.MarketID != "" {
			queryParams["market_id"] = params.MarketID
		}
	}

	var resp GetUserPositionsResponse
	err := s.client.get(ctx, s.client.dataClient, endpoint, queryParams, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

type GetTradesParams struct {
	UserID     string    `json:"user_id,omitempty"`
	MarketID   string    `json:"market_id,omitempty"`
	StartTime  time.Time `json:"start_time,omitempty"`
	EndTime    time.Time `json:"end_time,omitempty"`
	Limit      int       `json:"limit,omitempty"`
	Cursor     string    `json:"cursor,omitempty"`
}

type GetTradesResponse struct {
	Trades     []Trade `json:"trades"`
	NextCursor string  `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more,omitempty"`
}

func (s *DataService) GetTrades(ctx context.Context, params *GetTradesParams) (*GetTradesResponse, error) {
	endpoint := "/trades"
	queryParams := make(map[string]string)

	if params != nil {
		if params.UserID != "" {
			queryParams["user_id"] = params.UserID
		}
		if params.MarketID != "" {
			queryParams["market_id"] = params.MarketID
		}
		if !params.StartTime.IsZero() {
			queryParams["start_time"] = params.StartTime.Format(time.RFC3339)
		}
		if !params.EndTime.IsZero() {
			queryParams["end_time"] = params.EndTime.Format(time.RFC3339)
		}
		if params.Limit > 0 {
			queryParams["limit"] = string(rune(params.Limit))
		}
		if params.Cursor != "" {
			queryParams["cursor"] = params.Cursor
		}
	}

	var resp GetTradesResponse
	err := s.client.get(ctx, s.client.dataClient, endpoint, queryParams, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *DataService) GetUserActivity(ctx context.Context, userID string) (*UserActivity, error) {
	endpoint := "/activity/" + userID
	var activity UserActivity
	err := s.client.get(ctx, s.client.dataClient, endpoint, nil, &activity)
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

func (s *DataService) GetTopHolders(ctx context.Context, marketID string) ([]HolderData, error) {
	endpoint := "/holders/" + marketID
	var holders []HolderData
	err := s.client.get(ctx, s.client.dataClient, endpoint, nil, &holders)
	if err != nil {
		return nil, err
	}
	return holders, nil
}

func (s *DataService) GetOpenInterest(ctx context.Context, marketID string) (*OpenInterestData, error) {
	endpoint := "/open-interest/" + marketID
	var oi OpenInterestData
	err := s.client.get(ctx, s.client.dataClient, endpoint, nil, &oi)
	if err != nil {
		return nil, err
	}
	return &oi, nil
}

type GetLeaderboardParams struct {
	Period   string `json:"period,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
	SortBy   string `json:"sort_by,omitempty"`
}

type GetLeaderboardResponse struct {
	Entries []LeaderboardEntry `json:"entries"`
	Total   int                `json:"total"`
}

func (s *DataService) GetLeaderboard(ctx context.Context, params *GetLeaderboardParams) (*GetLeaderboardResponse, error) {
	endpoint := "/leaderboard"
	queryParams := make(map[string]string)

	if params != nil {
		if params.Period != "" {
			queryParams["period"] = params.Period
		}
		if params.Limit > 0 {
			queryParams["limit"] = string(rune(params.Limit))
		}
		if params.Offset > 0 {
			queryParams["offset"] = string(rune(params.Offset))
		}
		if params.SortBy != "" {
			queryParams["sort_by"] = params.SortBy
		}
	}

	var resp GetLeaderboardResponse
	err := s.client.get(ctx, s.client.dataClient, endpoint, queryParams, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *DataService) GetTotalPositionValue(ctx context.Context, userID string) (float64, error) {
	endpoint := "/positions/" + userID + "/total-value"
	var result struct {
		TotalValue float64 `json:"total_value"`
	}
	err := s.client.get(ctx, s.client.dataClient, endpoint, nil, &result)
	if err != nil {
		return 0, err
	}
	return result.TotalValue, nil
}

func (s *DataService) GetTotalMarketsTraded(ctx context.Context, userID string) (int, error) {
	endpoint := "/users/" + userID + "/total-markets"
	var result struct {
		TotalMarkets int `json:"total_markets"`
	}
	err := s.client.get(ctx, s.client.dataClient, endpoint, nil, &result)
	if err != nil {
		return 0, err
	}
	return result.TotalMarkets, nil
}

func (s *DataService) GetPositionsForMarket(ctx context.Context, userID, marketID string) ([]UserPosition, error) {
	endpoint := "/positions/" + userID + "/" + marketID
	var positions []UserPosition
	err := s.client.get(ctx, s.client.dataClient, endpoint, nil, &positions)
	if err != nil {
		return nil, err
	}
	return positions, nil
}

type GetBuilderLeaderboardParams struct {
	Period string `json:"period,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type BuilderLeaderboardEntry struct {
	Rank          int     `json:"rank"`
	BuilderID     string  `json:"builder_id"`
	BuilderName   string  `json:"builder_name"`
	TotalVolume   float64 `json:"total_volume"`
	TotalFees     float64 `json:"total_fees"`
	MarketCount   int     `json:"market_count"`
}

type GetBuilderLeaderboardResponse struct {
	Entries []BuilderLeaderboardEntry `json:"entries"`
}

func (s *DataService) GetBuilderLeaderboard(ctx context.Context, params *GetBuilderLeaderboardParams) (*GetBuilderLeaderboardResponse, error) {
	endpoint := "/builders/leaderboard"
	queryParams := make(map[string]string)

	if params != nil {
		if params.Period != "" {
			queryParams["period"] = params.Period
		}
		if params.Limit > 0 {
			queryParams["limit"] = string(rune(params.Limit))
		}
	}

	var resp GetBuilderLeaderboardResponse
	err := s.client.get(ctx, s.client.dataClient, endpoint, queryParams, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

type VolumeTimeSeries struct {
	Timestamp time.Time `json:"timestamp"`
	Volume    float64   `json:"volume"`
}

type GetBuilderVolumeParams struct {
	BuilderID string    `json:"builder_id,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Interval  string    `json:"interval,omitempty"`
}

type GetBuilderVolumeResponse struct {
	Data []VolumeTimeSeries `json:"data"`
}

func (s *DataService) GetBuilderVolumeTimeSeries(ctx context.Context, params *GetBuilderVolumeResponse) (*GetBuilderVolumeResponse, error) {
	endpoint := "/builders/volume"
	queryParams := make(map[string]string)

	if params != nil {
		if !params.Data[0].Timestamp.IsZero() {
		}
	}

	var resp GetBuilderVolumeResponse
	err := s.client.get(ctx, s.client.dataClient, endpoint, queryParams, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
