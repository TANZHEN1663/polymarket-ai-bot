package polymarket

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

type StrategyEngine struct {
	client       *Client
	analyzer     *MarketAnalyzer
	predictor    *AIPredictor
	riskManager  *RiskManager
	strategies   map[string]TradingStrategy
	mu           sync.RWMutex
	positions    map[string]*Position
	tradeHistory []Trade
}

type TradingStrategy interface {
	Name() string
	Execute(ctx context.Context, signal *TradingSignal) (*TradeExecution, error)
	Validate(signal *TradingSignal) bool
}

type TradingSignal struct {
	MarketID      string  `json:"market_id"`
	MarketTitle   string  `json:"market_title"`
	Strategy      string  `json:"strategy"`
	Direction     string  `json:"direction"`
	Strength      float64 `json:"strength"`
	Confidence    float64 `json:"confidence"`
	ExpectedValue float64 `json:"expected_value"`
	CurrentPrice  float64 `json:"current_price"`
	FairValue     float64 `json:"fair_value"`
	Opportunity   *MarketOpportunity `json:"opportunity"`
	Prediction    *AIPrediction      `json:"prediction"`
	Assessment    *ValueAssessment   `json:"assessment"`
	Timestamp     time.Time          `json:"timestamp"`
}

type TradeExecution struct {
	MarketID   string  `json:"market_id"`
	Strategy   string  `json:"strategy"`
	Side       string  `json:"side"`
	Outcome    string  `json:"outcome"`
	Size       float64 `json:"size"`
	Price      float64 `json:"price"`
	OrderID    string  `json:"order_id"`
	Status     string  `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
	Reasoning  string  `json:"reasoning"`
}

type Position struct {
	MarketID      string    `json:"market_id"`
	Outcome       string    `json:"outcome"`
	Size          float64   `json:"size"`
	AveragePrice  float64   `json:"average_price"`
	CurrentValue  float64   `json:"current_value"`
	PnL           float64   `json:"pnl"`
	PnLPercent    float64   `json:"pnl_percent"`
	OpenedAt      time.Time `json:"opened_at"`
	LastUpdated   time.Time `json:"last_updated"`
	Strategy      string    `json:"strategy"`
}

type ValueBetStrategy struct {
	engine *StrategyEngine
}

type ArbitrageStrategy struct {
	engine *StrategyEngine
}

type MomentumStrategy struct {
	engine *StrategyEngine
}

type HedgeStrategy struct {
	engine *StrategyEngine
}

func NewStrategyEngine(client *Client, analyzer *MarketAnalyzer, predictor *AIPredictor, riskManager *RiskManager) *StrategyEngine {
	engine := &StrategyEngine{
		client:      client,
		analyzer:    analyzer,
		predictor:   predictor,
		riskManager: riskManager,
		strategies:  make(map[string]TradingStrategy),
		positions:   make(map[string]*Position),
		tradeHistory: make([]Trade, 0),
	}

	engine.RegisterStrategy(&ValueBetStrategy{engine: engine})
	engine.RegisterStrategy(&ArbitrageStrategy{engine: engine})
	engine.RegisterStrategy(&MomentumStrategy{engine: engine})
	engine.RegisterStrategy(&HedgeStrategy{engine: engine})

	return engine
}

func (se *StrategyEngine) RegisterStrategy(strategy TradingStrategy) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.strategies[strategy.Name()] = strategy
}

func (se *StrategyEngine) GetStrategy(name string) (TradingStrategy, bool) {
	se.mu.RLock()
	defer se.mu.RUnlock()
	strategy, ok := se.strategies[name]
	return strategy, ok
}

func (se *StrategyEngine) ProcessSignal(ctx context.Context, signal *TradingSignal) (*TradeExecution, error) {
	if !se.riskManager.CanTrade() {
		return nil, fmt.Errorf("trading halted by risk manager")
	}

	strategy, ok := se.GetStrategy(signal.Strategy)
	if !ok {
		return nil, fmt.Errorf("strategy %s not found", signal.Strength)
	}

	if !strategy.Validate(signal) {
		return nil, fmt.Errorf("signal validation failed")
	}

	if !se.riskManager.CheckTradeLimits(signal) {
		return nil, fmt.Errorf("trade exceeds risk limits")
	}

	execution, err := strategy.Execute(ctx, signal)
	if err != nil {
		return nil, fmt.Errorf("strategy execution failed: %w", err)
	}

	se.riskManager.RecordTrade(execution)
	se.updatePosition(execution)

	return execution, nil
}

func (s *ValueBetStrategy) Name() string {
	return "VALUE_BET"
}

func (s *ValueBetStrategy) Validate(signal *TradingSignal) bool {
	if signal.Assessment == nil {
		return false
	}

	if math.Abs(signal.Assessment.Edge) < 0.05 {
		return false
	}

	if signal.Confidence < 50 {
		return false
	}

	return true
}

func (s *ValueBetStrategy) Execute(ctx context.Context, signal *TradingSignal) (*TradeExecution, error) {
	outcome := "YES"
	if signal.Assessment.AIProbability < signal.Assessment.MarketImpliedProb {
		outcome = "NO"
	}

	size := s.calculateBetSize(signal)
	price := signal.CurrentPrice

	order := &CreateOrder{
		MarketID:  signal.MarketID,
		Side:      "BUY",
		Outcome:   outcome,
		Count:     size,
		Price:     price,
		OrderType: "LIMIT",
	}

	clobService := NewCLOBService(s.engine.client)
	resp, err := clobService.CreateOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	return &TradeExecution{
		MarketID:  signal.MarketID,
		Strategy:  s.Name(),
		Side:      "BUY",
		Outcome:   outcome,
		Size:      size,
		Price:     price,
		OrderID:   resp.OrderID,
		Status:    resp.Status,
		Timestamp: time.Now(),
		Reasoning: fmt.Sprintf("Edge: %.2f%%, Confidence: %.2f%%", signal.Assessment.EdgePercent, signal.Confidence),
	}, nil
}

func (s *ValueBetStrategy) calculateBetSize(signal *TradingSignal) float64 {
	edge := signal.Assessment.Edge
	confidence := signal.Confidence / 100.0

	baseSize := 100.0

	sizeMultiplier := edge * 10 * confidence
	size := baseSize * sizeMultiplier

	if size > 1000 {
		size = 1000
	}

	return size
}

func (s *ArbitrageStrategy) Name() string {
	return "ARBITRAGE"
}

func (s *ArbitrageStrategy) Validate(signal *TradingSignal) bool {
	if signal.Opportunity == nil {
		return false
	}

	if signal.Opportunity.ExpectedValue < 0.03 {
		return false
	}

	return true
}

func (s *ArbitrageStrategy) Execute(ctx context.Context, signal *TradingSignal) (*TradeExecution, error) {
	outcome := signal.Opportunity.RecommendedBet
	size := s.calculateBetSize(signal)

	price := signal.CurrentPrice
	if outcome == "NO" {
		price = 100 - price
	}

	order := &CreateOrder{
		MarketID:  signal.MarketID,
		Side:      "BUY",
		Outcome:   outcome,
		Count:     size,
		Price:     price,
		OrderType: "LIMIT",
	}

	clobService := NewCLOBService(s.engine.client)
	resp, err := clobService.CreateOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	return &TradeExecution{
		MarketID:  signal.MarketID,
		Strategy:  s.Name(),
		Side:      "BUY",
		Outcome:   outcome,
		Size:      size,
		Price:     price,
		OrderID:   resp.OrderID,
		Status:    resp.Status,
		Timestamp: time.Now(),
		Reasoning: signal.Opportunity.Reasoning,
	}, nil
}

func (s *ArbitrageStrategy) calculateBetSize(signal *TradingSignal) float64 {
	expectedValue := signal.Opportunity.ExpectedValue
	confidence := signal.Opportunity.Confidence / 100.0

	baseSize := 200.0
	size := baseSize * expectedValue * 10 * confidence

	if size > 2000 {
		size = 2000
	}

	return size
}

func (s *MomentumStrategy) Name() string {
	return "MOMENTUM"
}

func (s *MomentumStrategy) Validate(signal *TradingSignal) bool {
	if signal.Strength < 0.02 {
		return false
	}

	if signal.Confidence < 60 {
		return false
	}

	return true
}

func (s *MomentumStrategy) Execute(ctx context.Context, signal *TradingSignal) (*TradeExecution, error) {
	outcome := "YES"
	if signal.Direction == "DOWN" {
		outcome = "NO"
	}

	size := s.calculateBetSize(signal)
	price := signal.CurrentPrice

	order := &CreateOrder{
		MarketID:  signal.MarketID,
		Side:      "BUY",
		Outcome:   outcome,
		Count:     size,
		Price:     price,
		OrderType: "MARKET",
	}

	clobService := NewCLOBService(s.engine.client)
	resp, err := clobService.CreateOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	return &TradeExecution{
		MarketID:  signal.MarketID,
		Strategy:  s.Name(),
		Side:      "BUY",
		Outcome:   outcome,
		Size:      size,
		Price:     price,
		OrderID:   resp.OrderID,
		Status:    resp.Status,
		Timestamp: time.Now(),
		Reasoning: fmt.Sprintf("Momentum signal: %.2f%% strength", signal.Strength*100),
	}, nil
}

func (s *MomentumStrategy) calculateBetSize(signal *TradingSignal) float64 {
	strength := math.Abs(signal.Strength)
	confidence := signal.Confidence / 100.0

	baseSize := 150.0
	size := baseSize * strength * 100 * confidence

	if size > 1500 {
		size = 1500
	}

	return size
}

func (s *HedgeStrategy) Name() string {
	return "HEDGE"
}

func (s *HedgeStrategy) Validate(signal *TradingSignal) bool {
	position := s.engine.GetPosition(signal.MarketID)
	if position == nil {
		return false
	}

	if position.PnL > position.Size*0.5 {
		return true
	}

	return false
}

func (s *HedgeStrategy) Execute(ctx context.Context, signal *TradingSignal) (*TradeExecution, error) {
	position := s.engine.GetPosition(signal.MarketID)
	if position == nil {
		return nil, fmt.Errorf("no position to hedge")
	}

	hedgeOutcome := "YES"
	if position.Outcome == "YES" {
		hedgeOutcome = "NO"
	}

	hedgeSize := position.Size * 0.5
	price := signal.CurrentPrice

	order := &CreateOrder{
		MarketID:  signal.MarketID,
		Side:      "BUY",
		Outcome:   hedgeOutcome,
		Count:     hedgeSize,
		Price:     price,
		OrderType: "LIMIT",
	}

	clobService := NewCLOBService(s.engine.client)
	resp, err := clobService.CreateOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	return &TradeExecution{
		MarketID:  signal.MarketID,
		Strategy:  s.Name(),
		Side:      "BUY",
		Outcome:   hedgeOutcome,
		Size:      hedgeSize,
		Price:     price,
		OrderID:   resp.OrderID,
		Status:    resp.Status,
		Timestamp: time.Now(),
		Reasoning: fmt.Sprintf("Hedge existing %s position", position.Outcome),
	}, nil
}

func (s *HedgeStrategy) calculateBetSize(signal *TradingSignal) float64 {
	return 0
}

func (se *StrategyEngine) updatePosition(execution *TradeExecution) {
	se.mu.Lock()
	defer se.mu.Unlock()

	key := execution.MarketID + "_" + execution.Outcome

	if existing, ok := se.positions[key]; ok {
		totalSize := existing.Size + execution.Size
		totalCost := existing.Size*existing.AveragePrice + execution.Size*execution.Price
		existing.AveragePrice = totalCost / totalSize
		existing.Size = totalSize
		existing.LastUpdated = time.Now()
	} else {
		se.positions[key] = &Position{
			MarketID:     execution.MarketID,
			Outcome:      execution.Outcome,
			Size:         execution.Size,
			AveragePrice: execution.Price,
			CurrentValue: execution.Size * execution.Price,
			OpenedAt:     time.Now(),
			LastUpdated:  time.Now(),
			Strategy:     execution.Strategy,
		}
	}
}

func (se *StrategyEngine) GetPosition(marketID string) *Position {
	se.mu.RLock()
	defer se.mu.RUnlock()

	for key, pos := range se.positions {
		if pos.MarketID == marketID {
			_ = key
			return pos
		}
	}
	return nil
}

func (se *StrategyEngine) GetAllPositions() map[string]*Position {
	se.mu.RLock()
	defer se.mu.RUnlock()

	positions := make(map[string]*Position)
	for k, v := range se.positions {
		positions[k] = v
	}
	return positions
}

func (se *StrategyEngine) ClosePosition(ctx context.Context, marketID, outcome string) error {
	se.mu.Lock()
	key := marketID + "_" + outcome
	_, ok := se.positions[key]
	if !ok {
		se.mu.Unlock()
		return fmt.Errorf("position not found")
	}
	delete(se.positions, key)
	se.mu.Unlock()

	clobService := NewCLOBService(se.client)
	
	orders, err := clobService.GetUserOrders(ctx, &UserOrdersParams{
		MarketID:   marketID,
		ActiveOnly: true,
	})
	if err != nil {
		return err
	}

	for _, order := range orders.Orders {
		if order.Outcome == outcome {
			err := clobService.CancelOrder(ctx, order.OrderID)
			if err != nil {
				continue
			}
		}
	}

	return nil
}

func (se *StrategyEngine) GetTradeHistory() []Trade {
	se.mu.RLock()
	defer se.mu.RUnlock()

	history := make([]Trade, len(se.tradeHistory))
	copy(history, se.tradeHistory)
	return history
}

func (se *StrategyEngine) CalculateTotalPnL() float64 {
	se.mu.RLock()
	defer se.mu.RUnlock()

	totalPnL := 0.0
	for _, position := range se.positions {
		totalPnL += position.PnL
	}
	return totalPnL
}
