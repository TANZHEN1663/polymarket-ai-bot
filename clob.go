package polymarket

import (
	"context"
	"fmt"
	"time"
)

type CLOBService struct {
	client *Client
}

func NewCLOBService(client *Client) *CLOBService {
	return &CLOBService{client: client}
}

type OrderBook struct {
	MarketID    string    `json:"market_id"`
	Bids        []Order   `json:"bids"`
	Asks        []Order   `json:"asks"`
	LastUpdated time.Time `json:"last_updated"`
}

type Order struct {
	OrderID   string  `json:"order_id"`
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
	Side      string  `json:"side"`
	Outcome   string  `json:"outcome"`
	Timestamp time.Time `json:"timestamp"`
}

type MarketPrice struct {
	MarketID  string  `json:"market_id"`
	YesPrice  float64 `json:"yes_price"`
	NoPrice   float64 `json:"no_price"`
	LastPrice float64 `json:"last_price"`
	Volume24h float64 `json:"volume_24h"`
	Change24h float64 `json:"change_24h"`
}

type MidpointPrice struct {
	MarketID   string  `json:"market_id"`
	Midpoint   float64 `json:"midpoint"`
	BidPrice   float64 `json:"bid_price"`
	AskPrice   float64 `json:"ask_price"`
	Spread     float64 `json:"spread"`
	LastUpdated time.Time `json:"last_updated"`
}

type Spread struct {
	MarketID   string  `json:"market_id"`
	BidPrice   float64 `json:"bid_price"`
	AskPrice   float64 `json:"ask_price"`
	Spread     float64 `json:"spread"`
	SpreadPercent float64 `json:"spread_percent"`
}

type TradePrice struct {
	MarketID  string    `json:"market_id"`
	Price     float64   `json:"price"`
	Side      string    `json:"side"`
	Size      float64   `json:"size"`
	Timestamp time.Time `json:"timestamp"`
}

type PriceHistory struct {
	MarketID  string                 `json:"market_id"`
	Prices    []PriceHistoryPoint    `json:"prices"`
	Interval  string                 `json:"interval"`
}

type PriceHistoryPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
}

type FeeRate struct {
	MarketID    string  `json:"market_id"`
	MakerFee    float64 `json:"maker_fee"`
	TakerFee    float64 `json:"taker_fee"`
	VolumeTier  string  `json:"volume_tier"`
}

type TickSize struct {
	MarketID  string  `json:"market_id"`
	TickSize  float64 `json:"tick_size"`
	MinSize   float64 `json:"min_size"`
	MaxSize   float64 `json:"max_size"`
}

type ServerTime struct {
	ISO   string `json:"iso"`
	Epoch int64  `json:"epoch"`
}

type CreateOrder struct {
	MarketID    string  `json:"market_id"`
	Side        string  `json:"side"`
	Outcome     string  `json:"outcome"`
	Count       float64 `json:"count"`
	Price       float64 `json:"price"`
	OrderType   string  `json:"order_type"`
	Expiration  int64   `json:"expiration,omitempty"`
	GoodTill    int64   `json:"good_till,omitempty"`
}

type OrderResponse struct {
	OrderID   string `json:"order_id"`
	Status    string `json:"status"`
	MarketID  string `json:"market_id"`
	Side      string `json:"side"`
	Outcome   string `json:"outcome"`
	Count     float64 `json:"count"`
	Price     float64 `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

type CancelOrderRequest struct {
	OrderID string `json:"order_id"`
}

type CancelOrdersRequest struct {
	OrderIDs []string `json:"order_ids"`
}

type MarketOrdersRequest struct {
	MarketID string `json:"market_id"`
	Side     string `json:"side,omitempty"`
}

type UserOrdersParams struct {
	MarketID   string `json:"market_id,omitempty"`
	ActiveOnly bool   `json:"active_only,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
}

type UserOrdersResponse struct {
	Orders     []OrderResponse `json:"orders"`
	NextCursor string          `json:"next_cursor,omitempty"`
	HasMore    bool            `json:"has_more,omitempty"`
}

type HeartbeatRequest struct {
	Signature string `json:"signature"`
	Timestamp int64  `json:"timestamp"`
}

type HeartbeatResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type OrderScoringRequest struct {
	OrderID string `json:"order_id"`
}

type OrderScoringResponse struct {
	OrderID     string  `json:"order_id"`
	Score       float64 `json:"score"`
	Probability float64 `json:"probability"`
	Status      string  `json:"status"`
}

func (s *CLOBService) GetOrderBook(ctx context.Context, marketID string) (*OrderBook, error) {
	endpoint := "/book"
	params := map[string]string{
		"market_id": marketID,
	}
	var orderBook OrderBook
	err := s.client.get(ctx, s.client.clobClient, endpoint, params, &orderBook)
	if err != nil {
		return nil, err
	}
	return &orderBook, nil
}

type GetOrderBooksRequest struct {
	MarketIDs []string `json:"market_ids"`
}

func (s *CLOBService) GetOrderBooks(ctx context.Context, request *GetOrderBooksRequest) (map[string]OrderBook, error) {
	endpoint := "/book"
	var result map[string]OrderBook
	err := s.client.post(ctx, s.client.clobClient, endpoint, request, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *CLOBService) GetMarketPrice(ctx context.Context, marketID string) (*MarketPrice, error) {
	endpoint := "/price"
	params := map[string]string{
		"market_id": marketID,
	}
	var price MarketPrice
	err := s.client.get(ctx, s.client.clobClient, endpoint, params, &price)
	if err != nil {
		return nil, err
	}
	return &price, nil
}

type GetMarketPricesParams struct {
	MarketIDs []string `json:"market_ids,omitempty"`
}

func (s *CLOBService) GetMarketPrices(ctx context.Context, params *GetMarketPricesParams) ([]MarketPrice, error) {
	endpoint := "/prices"
	var prices []MarketPrice
	
	if params != nil && len(params.MarketIDs) > 0 {
		err := s.client.post(ctx, s.client.clobClient, endpoint, params, &prices)
		if err != nil {
			return nil, err
		}
	} else {
		err := s.client.get(ctx, s.client.clobClient, endpoint, nil, &prices)
		if err != nil {
			return nil, err
		}
	}
	
	return prices, nil
}

func (s *CLOBService) GetMidpointPrice(ctx context.Context, marketID string) (*MidpointPrice, error) {
	endpoint := "/midpoint"
	params := map[string]string{
		"market_id": marketID,
	}
	var midpoint MidpointPrice
	err := s.client.get(ctx, s.client.clobClient, endpoint, params, &midpoint)
	if err != nil {
		return nil, err
	}
	return &midpoint, nil
}

type GetMidpointPricesParams struct {
	MarketIDs []string `json:"market_ids,omitempty"`
}

func (s *CLOBService) GetMidpointPrices(ctx context.Context, params *GetMidpointPricesParams) ([]MidpointPrice, error) {
	endpoint := "/midpoints"
	var prices []MidpointPrice
	
	if params != nil && len(params.MarketIDs) > 0 {
		err := s.client.post(ctx, s.client.clobClient, endpoint, params, &prices)
		if err != nil {
			return nil, err
		}
	} else {
		err := s.client.get(ctx, s.client.clobClient, endpoint, nil, &prices)
		if err != nil {
			return nil, err
		}
	}
	
	return prices, nil
}

func (s *CLOBService) GetSpread(ctx context.Context, marketID string) (*Spread, error) {
	endpoint := "/spread"
	params := map[string]string{
		"market_id": marketID,
	}
	var spread Spread
	err := s.client.get(ctx, s.client.clobClient, endpoint, params, &spread)
	if err != nil {
		return nil, err
	}
	return &spread, nil
}

type GetSpreadsRequest struct {
	MarketIDs []string `json:"market_ids"`
}

func (s *CLOBService) GetSpreads(ctx context.Context, request *GetSpreadsRequest) ([]Spread, error) {
	endpoint := "/spreads"
	var spreads []Spread
	err := s.client.post(ctx, s.client.clobClient, endpoint, request, &spreads)
	if err != nil {
		return nil, err
	}
	return spreads, nil
}

func (s *CLOBService) GetLastTradePrice(ctx context.Context, marketID string) (*TradePrice, error) {
	endpoint := "/last-trade-price"
	params := map[string]string{
		"market_id": marketID,
	}
	var price TradePrice
	err := s.client.get(ctx, s.client.clobClient, endpoint, params, &price)
	if err != nil {
		return nil, err
	}
	return &price, nil
}

type GetLastTradePricesParams struct {
	MarketIDs []string `json:"market_ids,omitempty"`
}

func (s *CLOBService) GetLastTradePrices(ctx context.Context, params *GetLastTradePricesParams) ([]TradePrice, error) {
	endpoint := "/last-trade-prices"
	var prices []TradePrice
	
	if params != nil && len(params.MarketIDs) > 0 {
		err := s.client.post(ctx, s.client.clobClient, endpoint, params, &prices)
		if err != nil {
			return nil, err
		}
	} else {
		err := s.client.get(ctx, s.client.clobClient, endpoint, nil, &prices)
		if err != nil {
			return nil, err
		}
	}
	
	return prices, nil
}

type GetPriceHistoryParams struct {
	MarketID   string `json:"market_id"`
	Interval   string `json:"interval,omitempty"`
	StartTime  int64  `json:"start_time,omitempty"`
	EndTime    int64  `json:"end_time,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

func (s *CLOBService) GetPriceHistory(ctx context.Context, params *GetPriceHistoryParams) (*PriceHistory, error) {
	endpoint := "/price-history"
	queryParams := make(map[string]string)
	
	if params != nil {
		if params.MarketID != "" {
			queryParams["market_id"] = params.MarketID
		}
		if params.Interval != "" {
			queryParams["interval"] = params.Interval
		}
		if params.StartTime > 0 {
			queryParams["start_time"] = string(rune(params.StartTime))
		}
		if params.EndTime > 0 {
			queryParams["end_time"] = string(rune(params.EndTime))
		}
		if params.Limit > 0 {
			queryParams["limit"] = string(rune(params.Limit))
		}
	}
	
	var history PriceHistory
	err := s.client.get(ctx, s.client.clobClient, endpoint, queryParams, &history)
	if err != nil {
		return nil, err
	}
	return &history, nil
}

func (s *CLOBService) GetFeeRate(ctx context.Context, marketID string) (*FeeRate, error) {
	endpoint := "/fee-rate"
	params := map[string]string{
		"market_id": marketID,
	}
	var fee FeeRate
	err := s.client.get(ctx, s.client.clobClient, endpoint, params, &fee)
	if err != nil {
		return nil, err
	}
	return &fee, nil
}

func (s *CLOBService) GetTickSize(ctx context.Context, marketID string) (*TickSize, error) {
	endpoint := "/tick-size"
	params := map[string]string{
		"market_id": marketID,
	}
	var tickSize TickSize
	err := s.client.get(ctx, s.client.clobClient, endpoint, params, &tickSize)
	if err != nil {
		return nil, err
	}
	return &tickSize, nil
}

func (s *CLOBService) GetServerTime(ctx context.Context) (*ServerTime, error) {
	endpoint := "/time"
	var time ServerTime
	err := s.client.get(ctx, s.client.clobClient, endpoint, nil, &time)
	if err != nil {
		return nil, err
	}
	return &time, nil
}

func (s *CLOBService) CreateOrder(ctx context.Context, order *CreateOrder) (*OrderResponse, error) {
	endpoint := "/order"
	var response OrderResponse
	err := s.client.post(ctx, s.client.clobClient, endpoint, order, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *CLOBService) CancelOrder(ctx context.Context, orderID string) error {
	endpoint := "/order"
	params := map[string]string{
		"order_id": orderID,
	}
	return s.client.delete(ctx, s.client.clobClient, endpoint, params)
}

func (s *CLOBService) GetOrder(ctx context.Context, orderID string) (*OrderResponse, error) {
	endpoint := "/order/" + orderID
	var response OrderResponse
	err := s.client.get(ctx, s.client.clobClient, endpoint, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *CLOBService) CreateOrders(ctx context.Context, orders []CreateOrder) ([]OrderResponse, error) {
	endpoint := "/orders"
	var responses []OrderResponse
	err := s.client.post(ctx, s.client.clobClient, endpoint, orders, &responses)
	if err != nil {
		return nil, err
	}
	return responses, nil
}

func (s *CLOBService) GetUserOrders(ctx context.Context, params *UserOrdersParams) (*UserOrdersResponse, error) {
	endpoint := "/orders"
	queryParams := make(map[string]string)
	
	if params != nil {
		if params.MarketID != "" {
			queryParams["market_id"] = params.MarketID
		}
		if params.ActiveOnly {
			queryParams["active_only"] = "true"
		}
		if params.Limit > 0 {
			queryParams["limit"] = string(rune(params.Limit))
		}
		if params.Cursor != "" {
			queryParams["cursor"] = params.Cursor
		}
	}
	
	var response UserOrdersResponse
	err := s.client.get(ctx, s.client.clobClient, endpoint, queryParams, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *CLOBService) CancelOrders(ctx context.Context, orderIDs []string) error {
	endpoint := "/orders"
	body := map[string]interface{}{
		"order_ids": orderIDs,
	}
	return s.client.delete(ctx, s.client.clobClient, endpoint, body)
}

func (s *CLOBService) CancelAllOrders(ctx context.Context) error {
	endpoint := "/orders"
	return s.client.delete(ctx, s.client.clobClient, endpoint, nil)
}

func (s *CLOBService) CancelOrdersForMarket(ctx context.Context, marketID string) error {
	endpoint := "/orders/" + marketID
	return s.client.delete(ctx, s.client.clobClient, endpoint, nil)
}

func (s *CLOBService) GetOrderScoring(ctx context.Context, orderID string) (*OrderScoringResponse, error) {
	endpoint := "/order-scoring"
	params := map[string]string{
		"order_id": orderID,
	}
	var response OrderScoringResponse
	err := s.client.get(ctx, s.client.clobClient, endpoint, params, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *CLOBService) SendHeartbeat(ctx context.Context, heartbeat *HeartbeatRequest) (*HeartbeatResponse, error) {
	endpoint := "/heartbeat"
	var response HeartbeatResponse
	err := s.client.post(ctx, s.client.clobClient, endpoint, heartbeat, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *CLOBService) GetProfile(ctx context.Context, walletAddress string) (interface{}, error) {
	endpoint := "/profile/" + walletAddress
	var profile interface{}
	err := s.client.get(ctx, s.client.clobClient, endpoint, nil, &profile)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

type GetPositionsParams struct {
	UserID   string `json:"user_id,omitempty"`
	MarketID string `json:"market_id,omitempty"`
}

func (s *CLOBService) GetPositions(ctx context.Context, params *GetPositionsParams) ([]UserPosition, error) {
	endpoint := "/positions"
	queryParams := make(map[string]string)
	
	if params != nil {
		if params.UserID != "" {
			queryParams["user_id"] = params.UserID
		}
		if params.MarketID != "" {
			queryParams["market_id"] = params.MarketID
		}
	}
	
	var positions []UserPosition
	err := s.client.get(ctx, s.client.clobClient, endpoint, queryParams, &positions)
	if err != nil {
		return nil, err
	}
	return positions, nil
}

type GetClosedPositionsParams struct {
	UserID   string `json:"user_id,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Cursor   string `json:"cursor,omitempty"`
}

func (s *CLOBService) GetClosedPositions(ctx context.Context, params *GetClosedPositionsParams) (*GetUserPositionsResponse, error) {
	endpoint := "/positions/closed"
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
	}
	
	var response GetUserPositionsResponse
	err := s.client.get(ctx, s.client.clobClient, endpoint, queryParams, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *CLOBService) GetUserTrades(ctx context.Context, params *GetTradesParams) (*GetTradesResponse, error) {
	endpoint := "/trades"
	queryParams := make(map[string]string)
	
	if params != nil {
		if params.UserID != "" {
			queryParams["user_id"] = params.UserID
		}
		if params.MarketID != "" {
			queryParams["market_id"] = params.MarketID
		}
		if params.Limit > 0 {
			queryParams["limit"] = string(rune(params.Limit))
		}
		if params.Cursor != "" {
			queryParams["cursor"] = params.Cursor
		}
	}
	
	var response GetTradesResponse
	err := s.client.get(ctx, s.client.clobClient, endpoint, queryParams, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *CLOBService) DownloadAccountingSnapshot(ctx context.Context, userID string) ([]byte, error) {
	endpoint := "/accounting/" + userID
	resp, err := s.client.clobClient.R().
		SetContext(ctx).
		SetDoNotParseResponse(true).
		Get(endpoint)
	
	if err != nil {
		return nil, fmt.Errorf("failed to download snapshot: %w", err)
	}
	
	defer resp.RawBody().Close()
	body, err := readResponseBody(resp.RawResponse)
	if err != nil {
		return nil, err
	}
	
	return body, nil
}
