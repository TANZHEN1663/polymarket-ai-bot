package polymarket

import (
	"math"
	"sync"
	"time"
)

type RiskManager struct {
	config           *RiskConfig
	mu               sync.RWMutex
	totalCapital     float64
	usedCapital      float64
	totalPnL         float64
	dailyPnL         float64
	dailyLossLimit   float64
	maxDrawdown      float64
	peakCapital      float64
	currentDrawdown  float64
	tradingHalted    bool
	haltReason       string
	tradeCount       int
	dailyTradeCount  int
	lastResetDate    time.Time
	tradeHistory     []TradeRecord
	positionLimits   map[string]float64
}

type RiskConfig struct {
	TotalCapital      float64 `json:"total_capital"`
	MaxPositionSize   float64 `json:"max_position_size"`
	MaxPositionPercent float64 `json:"max_position_percent"`
	DailyLossLimit    float64 `json:"daily_loss_limit"`
	MaxDrawdown       float64 `json:"max_drawdown"`
	MaxTradeCount     int     `json:"max_trade_count"`
	DailyTradeLimit   int     `json:"daily_trade_limit"`
	KellyMultiplier   float64 `json:"kelly_multiplier"`
	MinConfidence     float64 `json:"min_confidence"`
	MaxLiquiditySlippage float64 `json:"max_liquidity_slippage"`
	DiversificationLimit float64 `json:"diversification_limit"`
}

type TradeRecord struct {
	MarketID    string    `json:"market_id"`
	Strategy    string    `json:"strategy"`
	Side        string    `json:"side"`
	Size        float64   `json:"size"`
	Price       float64   `json:"price"`
	PnL         float64   `json:"pnl"`
	Timestamp   time.Time `json:"timestamp"`
}

type PositionLimits struct {
	MaxPerMarket    float64 `json:"max_per_market"`
	MaxPerOutcome   float64 `json:"max_per_outcome"`
	MaxTotalExposure float64 `json:"max_total_exposure"`
	MaxCorrelatedExposure float64 `json:"max_correlated_exposure"`
}

func NewRiskManager(config *RiskConfig) *RiskManager {
	if config == nil {
		config = &RiskConfig{
			TotalCapital:         10000,
			MaxPositionSize:      500,
			MaxPositionPercent:   0.05,
			DailyLossLimit:       500,
			MaxDrawdown:          0.15,
			MaxTradeCount:        100,
			DailyTradeLimit:      20,
			KellyMultiplier:      0.25,
			MinConfidence:        50.0,
			MaxLiquiditySlippage: 0.05,
			DiversificationLimit: 0.2,
		}
	}

	rm := &RiskManager{
		config:         config,
		totalCapital:   config.TotalCapital,
		peakCapital:    config.TotalCapital,
		dailyLossLimit: config.DailyLossLimit,
		maxDrawdown:    config.MaxDrawdown,
		positionLimits: make(map[string]float64),
		lastResetDate:  time.Now(),
		tradeHistory:   make([]TradeRecord, 0),
	}

	if config.DailyTradeLimit > 0 {
		rm.dailyTradeCount = 0
	}

	return rm
}

func (rm *RiskManager) CanTrade() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if rm.tradingHalted {
		return false
	}

	if rm.dailyPnL <= -rm.dailyLossLimit {
		return false
	}

	if rm.currentDrawdown > rm.maxDrawdown {
		return false
	}

	if rm.dailyTradeCount >= rm.config.DailyTradeLimit {
		return false
	}

	if rm.usedCapital >= rm.totalCapital*0.9 {
		return false
	}

	return true
}

func (rm *RiskManager) CheckTradeLimits(signal *TradingSignal) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if signal.Confidence < rm.config.MinConfidence {
		return false
	}

	estimatedSize := rm.CalculatePositionSize(signal)
	if estimatedSize > rm.config.MaxPositionSize {
		return false
	}

	if estimatedSize > rm.totalCapital*rm.config.MaxPositionPercent {
		return false
	}

	remainingCapital := rm.totalCapital - rm.usedCapital
	if estimatedSize > remainingCapital {
		return false
	}

	existingExposure := rm.positionLimits[signal.MarketID]
	if existingExposure+estimatedSize > rm.config.DiversificationLimit*rm.totalCapital {
		return false
	}

	return true
}

func (rm *RiskManager) CalculatePositionSize(signal *TradingSignal) float64 {
	if signal.Assessment != nil && signal.Assessment.Edge > 0 {
		kellyFraction := rm.calculateKellyFraction(signal)
		kellySize := rm.totalCapital * kellyFraction

		maxSize := rm.config.MaxPositionSize
		if kellySize > maxSize {
			kellySize = maxSize
		}

		return kellySize
	}

	baseSize := rm.totalCapital * 0.01

	confidenceMultiplier := signal.Confidence / 100.0
	size := baseSize * confidenceMultiplier

	if size > rm.config.MaxPositionSize {
		size = rm.config.MaxPositionSize
	}

	return size
}

func (rm *RiskManager) calculateKellyFraction(signal *TradingSignal) float64 {
	var winProbability float64
	var winLossRatio float64

	if signal.Assessment != nil {
		winProbability = signal.Assessment.AIProbability
		edge := signal.Assessment.Edge
		if edge > 0 {
			winLossRatio = edge / (1 - winProbability)
		} else {
			winLossRatio = 1.0
		}
	} else {
		winProbability = signal.Confidence / 100.0
		winLossRatio = 1.0
	}

	if winLossRatio <= 0 {
		return 0
	}

	kellyFraction := (winProbability*winLossRatio - (1 - winProbability)) / winLossRatio

	if kellyFraction < 0 {
		return 0
	}

	kellyFraction *= rm.config.KellyMultiplier

	if kellyFraction > 0.25 {
		kellyFraction = 0.25
	}

	return kellyFraction
}

func (rm *RiskManager) RecordTrade(execution *TradeExecution) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.usedCapital += execution.Size * execution.Price
	rm.tradeCount++
	rm.dailyTradeCount++

	rm.tradeHistory = append(rm.tradeHistory, TradeRecord{
		MarketID:  execution.MarketID,
		Strategy:  execution.Strategy,
		Side:      execution.Side,
		Size:      execution.Size,
		Price:     execution.Price,
		Timestamp: execution.Timestamp,
	})

	existingExposure := rm.positionLimits[execution.MarketID]
	rm.positionLimits[execution.MarketID] = existingExposure + execution.Size*execution.Price
}

func (rm *RiskManager) UpdatePnL(pnl float64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.totalPnL += pnl
	rm.dailyPnL += pnl

	newCapital := rm.totalCapital + rm.totalPnL

	if newCapital > rm.peakCapital {
		rm.peakCapital = newCapital
	}

	rm.currentDrawdown = (rm.peakCapital - newCapital) / rm.peakCapital

	if rm.currentDrawdown > rm.maxDrawdown {
		rm.HaltTrading("Maximum drawdown exceeded")
	}

	if rm.dailyPnL <= -rm.dailyLossLimit {
		rm.HaltTrading("Daily loss limit reached")
	}
}

func (rm *RiskManager) HaltTrading(reason string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.tradingHalted = true
	rm.haltReason = reason
}

func (rm *RiskManager) ResumeTrading() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.tradingHalted = false
	rm.haltReason = ""
}

func (rm *RiskManager) IsTradingHalted() (bool, string) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.tradingHalted, rm.haltReason
}

func (rm *RiskManager) ResetDailyStats() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.dailyPnL = 0
	rm.dailyTradeCount = 0
	rm.lastResetDate = time.Now()

	if rm.tradingHalted && rm.haltReason == "Daily loss limit reached" {
		rm.tradingHalted = false
		rm.haltReason = ""
	}
}

func (rm *RiskManager) GetRiskMetrics() *RiskMetrics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	currentCapital := rm.totalCapital + rm.totalPnL

	return &RiskMetrics{
		TotalCapital:     rm.totalCapital,
		CurrentCapital:   currentCapital,
		UsedCapital:      rm.usedCapital,
		AvailableCapital: currentCapital - rm.usedCapital,
		TotalPnL:         rm.totalPnL,
		DailyPnL:         rm.dailyPnL,
		PeakCapital:      rm.peakCapital,
		CurrentDrawdown:  rm.currentDrawdown,
		MaxDrawdown:      rm.maxDrawdown,
		TradingHalted:    rm.tradingHalted,
		HaltReason:       rm.haltReason,
		TradeCount:       rm.tradeCount,
		DailyTradeCount:  rm.dailyTradeCount,
		PositionCount:    len(rm.positionLimits),
		RiskUtilization:  rm.usedCapital / currentCapital,
	}
}

type RiskMetrics struct {
	TotalCapital      float64 `json:"total_capital"`
	CurrentCapital    float64 `json:"current_capital"`
	UsedCapital       float64 `json:"used_capital"`
	AvailableCapital  float64 `json:"available_capital"`
	TotalPnL          float64 `json:"total_pnl"`
	DailyPnL          float64 `json:"daily_pnl"`
	PeakCapital       float64 `json:"peak_capital"`
	CurrentDrawdown   float64 `json:"current_drawdown"`
	MaxDrawdown       float64 `json:"max_drawdown"`
	TradingHalted     bool    `json:"trading_halted"`
	HaltReason        string  `json:"halt_reason"`
	TradeCount        int     `json:"trade_count"`
	DailyTradeCount   int     `json:"daily_trade_count"`
	PositionCount     int     `json:"position_count"`
	RiskUtilization   float64 `json:"risk_utilization"`
}

func (rm *RiskManager) GetTradeHistory(limit int) []TradeRecord {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if limit <= 0 || limit > len(rm.tradeHistory) {
		limit = len(rm.tradeHistory)
	}

	start := len(rm.tradeHistory) - limit
	if start < 0 {
		start = 0
	}

	history := make([]TradeRecord, limit)
	copy(history, rm.tradeHistory[start:])

	return history
}

func (rm *RiskManager) GetPositionExposure(marketID string) float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.positionLimits[marketID]
}

func (rm *RiskManager) GetTotalExposure() float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	total := 0.0
	for _, exposure := range rm.positionLimits {
		total += exposure
	}
	return total
}

func (rm *RiskManager) UpdateCapital(newCapital float64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.totalCapital = newCapital
	if newCapital > rm.peakCapital {
		rm.peakCapital = newCapital
	}
}

func (rm *RiskManager) GetConfig() *RiskConfig {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.config
}

func (rm *RiskManager) UpdateConfig(config *RiskConfig) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.config = config
}

func (rm *RiskManager) CalculateSharpeRatio() float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if len(rm.tradeHistory) < 2 {
		return 0
	}

	var returns []float64
	for i := 1; i < len(rm.tradeHistory); i++ {
		pnl := rm.tradeHistory[i].PnL
		cost := rm.tradeHistory[i].Size * rm.tradeHistory[i].Price
		if cost > 0 {
			returns = append(returns, pnl/cost)
		}
	}

	if len(returns) == 0 {
		return 0
	}

	mean := 0.0
	for _, r := range returns {
		mean += r
	}
	mean /= float64(len(returns))

	variance := 0.0
	for _, r := range returns {
		diff := r - mean
		variance += diff * diff
	}
	variance /= float64(len(returns))

	stdDev := math.Sqrt(variance)

	if stdDev == 0 {
		return 0
	}

	riskFreeRate := 0.02 / 252
	sharpe := (mean - riskFreeRate) / stdDev

	return sharpe * math.Sqrt(252)
}

func (rm *RiskManager) GetWinRate() float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if len(rm.tradeHistory) == 0 {
		return 0
	}

	wins := 0
	for _, trade := range rm.tradeHistory {
		if trade.PnL > 0 {
			wins++
		}
	}

	return float64(wins) / float64(len(rm.tradeHistory)) * 100
}

func (rm *RiskManager) GetProfitFactor() float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	grossProfit := 0.0
	grossLoss := 0.0

	for _, trade := range rm.tradeHistory {
		if trade.PnL > 0 {
			grossProfit += trade.PnL
		} else {
			grossLoss += math.Abs(trade.PnL)
		}
	}

	if grossLoss == 0 {
		if grossProfit > 0 {
			return math.Inf(1)
		}
		return 0
	}

	return grossProfit / grossLoss
}
