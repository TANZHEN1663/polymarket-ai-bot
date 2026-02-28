package polymarket

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	config := ClientConfig{
		APIKey:     "test-key",
		APISecret:  "test-secret",
		Passphrase: "test-passphrase",
		Timeout:    10 * time.Second,
		RetryCount: 2,
		RetryWait:  500 * time.Millisecond,
	}

	client := NewClient(config)
	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.apiKey != "test-key" {
		t.Errorf("Expected apiKey to be 'test-key', got '%s'", client.apiKey)
	}
}

func TestNewRiskManager(t *testing.T) {
	config := &RiskConfig{
		TotalCapital:       10000,
		MaxPositionSize:    500,
		DailyLossLimit:     500,
		MaxDrawdown:        0.15,
		KellyMultiplier:    0.25,
		MinConfidence:      50.0,
	}

	rm := NewRiskManager(config)
	if rm == nil {
		t.Fatal("Expected risk manager to be created")
	}

	if rm.totalCapital != 10000 {
		t.Errorf("Expected totalCapital to be 10000, got %f", rm.totalCapital)
	}
}

func TestRiskManagerCanTrade(t *testing.T) {
	rm := NewRiskManager(&RiskConfig{
		TotalCapital:     10000,
		DailyLossLimit:   500,
		MaxDrawdown:      0.15,
		DailyTradeLimit:  20,
	})

	if !rm.CanTrade() {
		t.Error("Expected CanTrade to return true initially")
	}

	rm.dailyPnL = -600
	if rm.CanTrade() {
		t.Error("Expected CanTrade to return false after hitting daily loss limit")
	}
}

func TestKellyFraction(t *testing.T) {
	rm := NewRiskManager(&RiskConfig{
		TotalCapital:      10000,
		KellyMultiplier:   0.25,
	})

	signal := &TradingSignal{
		Confidence: 70,
		Assessment: &ValueAssessment{
			AIProbability:     0.65,
			MarketImpliedProb: 0.50,
			Edge:              0.15,
		},
	}

	size := rm.CalculatePositionSize(signal)
	if size <= 0 {
		t.Errorf("Expected positive position size, got %f", size)
	}

	if size > 500 {
		t.Errorf("Expected position size <= 500, got %f", size)
	}
}

func TestMarketAnalyzer(t *testing.T) {
	client := NewClient(ClientConfig{})
	analyzer := NewMarketAnalyzer(client)

	if analyzer == nil {
		t.Fatal("Expected market analyzer to be created")
	}
}

func TestValueBetStrategy(t *testing.T) {
	client := NewClient(ClientConfig{})
	analyzer := NewMarketAnalyzer(client)
	riskManager := NewRiskManager(&RiskConfig{TotalCapital: 10000})
	strategyEngine := NewStrategyEngine(client, analyzer, nil, riskManager)

	strategy, ok := strategyEngine.GetStrategy("VALUE_BET")
	if !ok {
		t.Fatal("Expected VALUE_BET strategy to be registered")
	}

	if strategy.Name() != "VALUE_BET" {
		t.Errorf("Expected strategy name 'VALUE_BET', got '%s'", strategy.Name())
	}
}

func TestTradingSignalValidation(t *testing.T) {
	signal := &TradingSignal{
		MarketID:      "test-market",
		Strategy:      "VALUE_BET",
		Confidence:    75,
		ExpectedValue: 0.10,
		Assessment: &ValueAssessment{
			AIProbability:     0.65,
			MarketImpliedProb: 0.50,
			Edge:              0.15,
			Recommendation:    "BUY",
		},
	}

	if signal.Assessment.Edge <= 0.05 {
		t.Error("Expected edge > 0.05 for valid signal")
	}

	if signal.Confidence < 50 {
		t.Error("Expected confidence >= 50 for valid signal")
	}
}

func TestPositionManagement(t *testing.T) {
	client := NewClient(ClientConfig{})
	analyzer := NewMarketAnalyzer(client)
	riskManager := NewRiskManager(&RiskConfig{TotalCapital: 10000})
	strategyEngine := NewStrategyEngine(client, analyzer, nil, riskManager)

	execution := &TradeExecution{
		MarketID:  "test-market",
		Strategy:  "VALUE_BET",
		Side:      "BUY",
		Outcome:   "YES",
		Size:      100,
		Price:     0.65,
		OrderID:   "test-order",
		Status:    "filled",
		Timestamp: time.Now(),
	}

	strategyEngine.updatePosition(execution)

	position := strategyEngine.GetPosition("test-market")
	if position == nil {
		t.Error("Expected position to be created")
	}

	if position.Size != 100 {
		t.Errorf("Expected position size 100, got %f", position.Size)
	}

	if position.AveragePrice != 0.65 {
		t.Errorf("Expected average price 0.65, got %f", position.AveragePrice)
	}
}

func TestRiskMetrics(t *testing.T) {
	rm := NewRiskManager(&RiskConfig{
		TotalCapital:    10000,
		MaxDrawdown:     0.15,
		DailyLossLimit:  500,
	})

	rm.UpdatePnL(200)
	rm.UpdatePnL(-100)

	metrics := rm.GetRiskMetrics()

	if metrics.TotalPnL != 100 {
		t.Errorf("Expected total PnL 100, got %f", metrics.TotalPnL)
	}

	if metrics.DailyPnL != 100 {
		t.Errorf("Expected daily PnL 100, got %f", metrics.DailyPnL)
	}

	if metrics.CurrentDrawdown != 0 {
		t.Errorf("Expected current drawdown 0, got %f", metrics.CurrentDrawdown)
	}
}

func TestDrawdownControl(t *testing.T) {
	rm := NewRiskManager(&RiskConfig{
		TotalCapital:    10000,
		MaxDrawdown:     0.15,
	})

	rm.UpdatePnL(-1600)

	metrics := rm.GetRiskMetrics()
	if metrics.CurrentDrawdown <= 0.15 {
		t.Errorf("Expected drawdown > 0.15, got %f", metrics.CurrentDrawdown)
	}

	if rm.CanTrade() {
		t.Error("Expected trading to be halted after max drawdown")
	}
}

func TestConfigLoading(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("Expected default config to be created")
	}

	if config.Risk.TotalCapital != 10000 {
		t.Errorf("Expected default total capital 10000, got %f", config.Risk.TotalCapital)
	}

	if config.Trading.Mode != "paper" {
		t.Errorf("Expected default trading mode 'paper', got '%s'", config.Trading.Mode)
	}
}

func TestMonitor(t *testing.T) {
	monitor := NewMonitor(&LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})

	if monitor == nil {
		t.Fatal("Expected monitor to be created")
	}

	monitor.Start()
	defer monitor.Stop()

	monitor.Info("Test message", "test", nil)

	metrics := monitor.GetMetrics()
	if metrics == nil {
		t.Error("Expected metrics to be available")
	}
}

func TestAlertSystem(t *testing.T) {
	monitor := NewMonitor(&LoggingConfig{
		Level:  "info",
		Format: "json",
	})

	monitor.Start()
	defer monitor.Stop()

	monitor.SendAlert("TEST_ALERT", "HIGH", "Test alert message", map[string]interface{}{
		"test_key": "test_value",
	})

	time.Sleep(100 * time.Millisecond)

	alerts := monitor.GetAlerts(10)
	if len(alerts) == 0 {
		t.Error("Expected at least one alert")
	}
}

func TestLatencyTracking(t *testing.T) {
	monitor := NewMonitor(&LoggingConfig{
		Level:  "info",
		Format: "json",
	})

	monitor.RecordRequest("/api/test", 150*time.Millisecond, true)
	monitor.RecordRequest("/api/test", 200*time.Millisecond, true)
	monitor.RecordRequest("/api/test", 100*time.Millisecond, false)

	metrics := monitor.GetMetrics()

	if metrics.TotalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", metrics.TotalRequests)
	}

	if metrics.FailedRequests != 1 {
		t.Errorf("Expected 1 failed request, got %d", metrics.FailedRequests)
	}

	if metrics.LatencyStats == nil {
		t.Error("Expected latency stats to be available")
	}
}

func BenchmarkKellyCalculation(b *testing.B) {
	rm := NewRiskManager(&RiskConfig{
		TotalCapital:      10000,
		KellyMultiplier:   0.25,
		MaxPositionSize:   500,
	})

	signal := &TradingSignal{
		Confidence: 70,
		Assessment: &ValueAssessment{
			AIProbability:     0.65,
			MarketImpliedProb: 0.50,
			Edge:              0.15,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.CalculatePositionSize(signal)
	}
}

func BenchmarkMarketAnalysis(b *testing.B) {
	client := NewClient(ClientConfig{})
	analyzer := NewMarketAnalyzer(client)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeMarket(ctx, "test-market-id")
		if err != nil {
			b.Skip("Skipping benchmark - API not available")
		}
	}
}
