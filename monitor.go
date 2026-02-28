package polymarket

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

type Monitor struct {
	config      *LoggingConfig
	logger      *log.Logger
	fileLogger  *log.Logger
	mu          sync.RWMutex
	metrics     *Metrics
	alerts      []Alert
	alertChan   chan Alert
	eventChan   chan MonitorEvent
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

type Metrics struct {
	StartTime       time.Time            `json:"start_time"`
	TotalRequests   int64                `json:"total_requests"`
	FailedRequests  int64                `json:"failed_requests"`
	TotalTrades     int64                `json:"total_trades"`
	ProfitableTrades int64               `json:"profitable_trades"`
	TotalPnL        float64              `json:"total_pnl"`
	DailyPnL        float64              `json:"daily_pnl"`
	WinRate         float64              `json:"win_rate"`
	SharpeRatio     float64              `json:"sharpe_ratio"`
	MaxDrawdown     float64              `json:"max_drawdown"`
	CurrentDrawdown float64              `json:"current_drawdown"`
	OpenPositions   int                  `json:"open_positions"`
	APIErrors       map[string]int64     `json:"api_errors"`
	LatencyStats    *LatencyStats        `json:"latency_stats"`
	LastUpdated     time.Time            `json:"last_updated"`
}

type LatencyStats struct {
	Average   float64 `json:"average"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	P50       float64 `json:"p50"`
	P95       float64 `json:"p95"`
	P99       float64 `json:"p99"`
}

type Alert struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Resolved  bool      `json:"resolved"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type MonitorEvent struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type LogEntry struct {
	Timestamp time.Time            `json:"timestamp"`
	Level     string               `json:"level"`
	Message   string               `json:"message"`
	Module    string               `json:"module,omitempty"`
	Error     string               `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

func NewMonitor(config *LoggingConfig) *Monitor {
	monitor := &Monitor{
		config:    config,
		metrics:   &Metrics{StartTime: time.Now(), APIErrors: make(map[string]int64)},
		alertChan: make(chan Alert, 100),
		eventChan: make(chan MonitorEvent, 1000),
		stopChan:  make(chan struct{}),
	}

	var output io.Writer = os.Stdout
	if config.Output == "file" && config.File != "" {
		file, err := os.OpenFile(config.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Failed to open log file: %v, using stdout", err)
		} else {
			output = file
		}
	}

	monitor.logger = log.New(output, "", 0)

	return monitor
}

func (m *Monitor) Start() {
	m.wg.Add(2)
	go m.processAlerts()
	go m.processEvents()
}

func (m *Monitor) Stop() {
	close(m.stopChan)
	m.wg.Wait()
}

func (m *Monitor) processAlerts() {
	defer m.wg.Done()

	for {
		select {
		case alert := <-m.alertChan:
			m.handleAlert(alert)
		case <-m.stopChan:
			return
		}
	}
}

func (m *Monitor) processEvents() {
	defer m.wg.Done()

	for {
		select {
		case event := <-m.eventChan:
			m.handleEvent(event)
		case <-m.stopChan:
			return
		}
	}
}

func (m *Monitor) handleAlert(alert Alert) {
	m.mu.Lock()
	m.alerts = append(m.alerts, alert)
	m.mu.Unlock()

	if alert.Severity == "HIGH" || alert.Severity == "CRITICAL" {
		m.Error(fmt.Sprintf("[ALERT] %s: %s", alert.Type, alert.Message), "monitor", nil)
	} else {
		m.Info(fmt.Sprintf("[ALERT] %s: %s", alert.Type, alert.Message), "monitor")
	}
}

func (m *Monitor) handleEvent(event MonitorEvent) {
	if m.config.Level == "debug" {
		m.Debug(fmt.Sprintf("[EVENT] %s", event.Type), "monitor", event.Data)
	}
}

func (m *Monitor) Debug(message string, module string, metadata ...map[string]interface{}) {
	if m.config.Level != "debug" {
		return
	}
	m.log("DEBUG", message, module, nil, metadata...)
}

func (m *Monitor) Info(message string, module string, metadata ...map[string]interface{}) {
	m.log("INFO", message, module, nil, metadata...)
}

func (m *Monitor) Warn(message string, module string, err error, metadata ...map[string]interface{}) {
	m.log("WARN", message, module, err, metadata...)
}

func (m *Monitor) Error(message string, module string, err error, metadata ...map[string]interface{}) {
	m.log("ERROR", message, module, err, metadata...)
}

func (m *Monitor) log(level, message, module string, err error, metadata ...map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Module:    module,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if len(metadata) > 0 {
		entry.Metadata = metadata[0]
	}

	if m.config.Format == "json" {
		data, _ := json.Marshal(entry)
		m.logger.Println(string(data))
	} else {
		m.logger.Printf("[%s] %s - %s: %s", level, entry.Timestamp.Format(time.RFC3339), module, message)
		if err != nil {
			m.logger.Printf("  Error: %v", err)
		}
	}
}

func (m *Monitor) RecordRequest(endpoint string, latency time.Duration, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics.TotalRequests++
	if !success {
		m.metrics.FailedRequests++
		if endpoint != "" {
			m.metrics.APIErrors[endpoint]++
		}
	}

	if m.metrics.LatencyStats == nil {
		m.metrics.LatencyStats = &LatencyStats{}
	}

	latencyMs := float64(latency.Nanoseconds()) / 1e6
	m.updateLatencyStats(latencyMs)

	m.metrics.LastUpdated = time.Now()
}

func (m *Monitor) updateLatencyStats(latency float64) {
	stats := m.metrics.LatencyStats

	if stats.Min == 0 || latency < stats.Min {
		stats.Min = latency
	}
	if latency > stats.Max {
		stats.Max = latency
	}

	count := m.metrics.TotalRequests
	if count == 0 {
		count = 1
	}

	stats.Average = (stats.Average*float64(count-1) + latency) / float64(count)

	stats.P50 = stats.Average * 0.8
	stats.P95 = stats.Average * 1.5
	stats.P99 = stats.Average * 2.0
}

func (m *Monitor) RecordTrade(trade *TradeExecution, pnl float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics.TotalTrades++
	if pnl > 0 {
		m.metrics.ProfitableTrades++
	}
	m.metrics.TotalPnL += pnl
	m.metrics.DailyPnL += pnl

	if m.metrics.TotalTrades > 0 {
		m.metrics.WinRate = float64(m.metrics.ProfitableTrades) / float64(m.metrics.TotalTrades) * 100
	}

	m.metrics.LastUpdated = time.Now()

	m.Info(fmt.Sprintf("Trade executed: %s %s %.2f @ %.2f, PnL: %.2f",
		trade.Strategy, trade.Side, trade.Size, trade.Price, pnl), "trading", map[string]interface{}{
		"market_id": trade.MarketID,
		"order_id":  trade.OrderID,
		"pnl":       pnl,
	})
}

func (m *Monitor) SendAlert(alertType, severity, message string, metadata map[string]interface{}) {
	alert := Alert{
		ID:        fmt.Sprintf("%s_%d", alertType, time.Now().UnixNano()),
		Type:      alertType,
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	select {
	case m.alertChan <- alert:
	default:
		m.Warn("Alert channel full, dropping alert", "monitor", nil, map[string]interface{}{
			"alert_type": alertType,
		})
	}
}

func (m *Monitor) RecordEvent(eventType string, data map[string]interface{}) {
	event := MonitorEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	select {
	case m.eventChan <- event:
	default:
		m.Warn("Event channel full, dropping event", "monitor", nil, nil)
	}
}

func (m *Monitor) GetMetrics() *Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := *m.metrics
	metrics.LastUpdated = time.Now()

	return &metrics
}

func (m *Monitor) GetAlerts(limit int) []Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.alerts) {
		limit = len(m.alerts)
	}

	start := len(m.alerts) - limit
	if start < 0 {
		start = 0
	}

	alerts := make([]Alert, limit)
	copy(alerts, m.alerts[start:])

	return alerts
}

func (m *Monitor) ResolveAlert(alertID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.alerts {
		if m.alerts[i].ID == alertID {
			m.alerts[i].Resolved = true
			break
		}
	}
}

func (m *Monitor) UpdateRiskMetrics(riskMetrics *RiskMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if riskMetrics != nil {
		m.metrics.CurrentDrawdown = riskMetrics.CurrentDrawdown
		if riskMetrics.CurrentDrawdown > m.metrics.MaxDrawdown {
			m.metrics.MaxDrawdown = riskMetrics.CurrentDrawdown
		}

		if riskMetrics.CurrentDrawdown > 0.1 {
			m.SendAlert("HIGH_DRAWDOWN", "HIGH",
				fmt.Sprintf("Current drawdown %.2f%% exceeds threshold", riskMetrics.CurrentDrawdown*100),
				map[string]interface{}{
					"drawdown": riskMetrics.CurrentDrawdown,
					"threshold": 0.1,
				})
		}
	}
}

func (m *Monitor) ExportMetrics() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.metrics, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (m *Monitor) GetUptime() time.Duration {
	return time.Since(m.metrics.StartTime)
}

func (m *Monitor) GetSystemInfo() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]interface{}{
		"go_version":    runtime.Version(),
		"os":            runtime.GOOS,
		"arch":          runtime.GOARCH,
		"num_cpu":       runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
		"alloc_bytes":   memStats.Alloc,
		"total_alloc":   memStats.TotalAlloc,
		"sys_bytes":     memStats.Sys,
		"gc_pause_ns":   memStats.PauseTotalNs,
	}
}

func (m *Monitor) LogStartup(config *Config) {
	m.Info("Polymarket AI Trading Bot starting", "main", map[string]interface{}{
		"version":         "1.0.0",
		"environment":     config.Polymarket.Environment,
		"trading_mode":    config.Trading.Mode,
		"ai_provider":     config.AI.Provider,
		"ai_model":        config.AI.Model,
		"total_capital":   config.Risk.TotalCapital,
		"strategies":      config.Trading.AllowedStrategies,
		"scan_interval":   config.Trading.ScanInterval,
	})
}

func (m *Monitor) LogShutdown() {
	metrics := m.GetMetrics()
	m.Info("Polymarket AI Trading Bot shutting down", "main", map[string]interface{}{
		"uptime":            m.GetUptime().String(),
		"total_requests":    metrics.TotalRequests,
		"failed_requests":   metrics.FailedRequests,
		"total_trades":      metrics.TotalTrades,
		"total_pnl":         metrics.TotalPnL,
		"win_rate":          metrics.WinRate,
		"sharpe_ratio":      metrics.SharpeRatio,
		"max_drawdown":      metrics.MaxDrawdown,
	})
}
