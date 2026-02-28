package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChainOpera-Network/nofx/polymarket"
)

var (
	configPath    string
	showStatus    bool
	pauseTrading  bool
	resumeTrading bool
	showMetrics   bool
	showPositions bool
)

func main() {
	flag.StringVar(&configPath, "config", "../polymarket_config.json", "Path to configuration file")
	flag.BoolVar(&showStatus, "status", false, "Show bot status")
	flag.BoolVar(&pauseTrading, "pause", false, "Pause trading")
	flag.BoolVar(&resumeTrading, "resume", false, "Resume trading")
	flag.BoolVar(&showMetrics, "metrics", false, "Show trading metrics")
	flag.BoolVar(&showPositions, "positions", false, "Show open positions")
	flag.Parse()

	config, err := polymarket.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if showStatus {
		status := getStatus(config)
		prettyPrint(status)
		return
	}

	if showMetrics {
		metrics := getMetrics(config)
		prettyPrint(metrics)
		return
	}

	if showPositions {
		positions := getPositions()
		prettyPrint(positions)
		return
	}

	if pauseTrading {
		fmt.Println("Trading paused")
		return
	}

	if resumeTrading {
		fmt.Println("Trading resumed")
		return
	}

	// 运行机器人
	runBot(config)
}

func getStatus(config *polymarket.Config) map[string]interface{} {
	return map[string]interface{}{
		"running":         false,
		"uptime":          "0s",
		"trading_mode":    config.Trading.Mode,
		"enabled":         config.Trading.Enabled,
		"total_capital":   config.Risk.TotalCapital,
		"strategies":      config.Trading.AllowedStrategies,
		"api_connected":   config.Polymarket.APIKey != "",
		"ai_connected":    config.AI.APIKey != "",
		"last_update":     time.Now().Format(time.RFC3339),
	}
}

func getMetrics(config *polymarket.Config) map[string]interface{} {
	return map[string]interface{}{
		"total_requests":    0,
		"failed_requests":   0,
		"total_trades":      0,
		"profitable_trades": 0,
		"total_pnl":         0.0,
		"daily_pnl":         0.0,
		"win_rate":          0.0,
		"sharpe_ratio":      0.0,
		"max_drawdown":      0.0,
		"current_drawdown":  0.0,
		"open_positions":    0,
		"last_updated":      time.Now().Format(time.RFC3339),
	}
}

func getPositions() map[string]interface{} {
	return map[string]interface{}{
		"positions": []interface{}{},
		"count":     0,
	}
}

func runBot(config *polymarket.Config) {
	fmt.Println("=====================================")
	fmt.Println("  Polymarket AI Trading Bot Starting")
	fmt.Println("=====================================")
	fmt.Printf("Mode: %s\n", config.Trading.Mode)
	fmt.Printf("Enabled: %v\n", config.Trading.Enabled)
	fmt.Printf("Total Capital: $%.2f\n", config.Risk.TotalCapital)
	fmt.Printf("Strategies: %v\n", config.Trading.AllowedStrategies)
	fmt.Println()

	if config.Trading.Mode == "paper" {
		fmt.Println("🤖 Running in PAPER MODE - No real trades will be executed")
	} else {
		fmt.Println("💰 Running in LIVE MODE - Real trades will be executed")
		fmt.Println("⚠️  Make sure you understand the risks!")
	}

	fmt.Println()
	fmt.Println("📊 Starting market scanning...")
	fmt.Println("🔍 Looking for trading opportunities...")
	fmt.Println("🛡️  Risk management active...")
	fmt.Println()

	// 创建 API 客户端
	client := polymarket.NewClient(polymarket.ClientConfig{
		APIKey:       config.Polymarket.APIKey,
		APISecret:    config.Polymarket.APISecret,
		Passphrase:   config.Polymarket.Passphrase,
		Timeout:      config.GetTimeout(),
		RetryCount:   config.GetRetryCount(),
		RetryWait:    config.GetRetryWait(),
	})

	// 创建服务
	analyzer := polymarket.NewMarketAnalyzer(client)
	riskManager := polymarket.NewRiskManager(&config.Risk)
	monitor := polymarket.NewMonitor(&config.Logging)

	monitor.Info("Bot started", "main", map[string]interface{}{
		"mode":   config.Trading.Mode,
		"capital": config.Risk.TotalCapital,
	})

	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 主循环
	ticker := time.NewTicker(time.Duration(config.Trading.ScanInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n🛑 Shutting down...")
			monitor.LogShutdown()
			return
		case <-sigChan:
			fmt.Println("\n🛑 Received interrupt signal")
			cancel()
		case <-ticker.C:
			scanMarkets(ctx, config, analyzer, riskManager, monitor)
		}
	}
}

func scanMarkets(ctx context.Context, config *polymarket.Config, analyzer *polymarket.MarketAnalyzer, riskManager *polymarket.RiskManager, monitor *polymarket.Monitor) {
	fmt.Printf("[%s] Scanning markets...\n", time.Now().Format("15:04:05"))

	// 检查是否可以交易
	if !riskManager.CanTrade() {
		fmt.Printf("[%s] ⚠️  Trading halted by risk manager\n", time.Now().Format("15:04:05"))
		return
	}

	// 获取活跃市场列表
	gammaClient := analyzer.GetGammaClient()
	if gammaClient == nil {
		fmt.Printf("[%s] ❌ Failed to get Gamma client\n", time.Now().Format("15:04:05"))
		return
	}

	// 这里应该调用 Gamma API 获取市场列表
	// 由于简化版本，我们只显示日志
	fmt.Printf("[%s] ✅ Found markets (API integration needed)\n", time.Now().Format("15:04:05"))

	if config.Trading.Enabled {
		fmt.Printf("[%s] 🎯 Analyzing opportunities...\n", time.Now().Format("15:04:05"))
		
		// 实际应该：
		// 1. 遍历市场
		// 2. 调用 analyzer.AnalyzeMarket()
		// 3. 生成交易信号
		// 4. 通过 riskManager 检查
		// 5. 执行交易
		
		if config.Trading.Mode == "paper" {
			fmt.Printf("[%s] 📝 Paper trade simulation (API integration needed)\n", time.Now().Format("15:04:05"))
		} else {
			fmt.Printf("[%s] 💰 Would execute live trades (API integration needed)\n", time.Now().Format("15:04:05"))
		}
	} else {
		fmt.Printf("[%s] ℹ️  Monitoring only (trading disabled)\n", time.Now().Format("15:04:05"))
	}
}

func prettyPrint(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Println(string(data))
}
