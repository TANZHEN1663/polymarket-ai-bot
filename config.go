package polymarket

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Polymarket PolymarketConfig `json:"polymarket"`
	Risk       RiskConfig       `json:"risk"`
	AI         AIConfig         `json:"ai"`
	Trading    TradingConfig    `json:"trading"`
	Logging    LoggingConfig    `json:"logging"`
}

type PolymarketConfig struct {
	APIKey       string `json:"api_key"`
	APISecret    string `json:"api_secret"`
	Passphrase   string `json:"passphrase"`
	Timeout      int    `json:"timeout"`
	RetryCount   int    `json:"retry_count"`
	RetryWait    int    `json:"retry_wait"`
	Environment  string `json:"environment"`
}

type AIConfig struct {
	Provider       string  `json:"provider"`
	Model          string  `json:"model"`
	APIKey         string  `json:"api_key"`
	BaseURL        string  `json:"base_url"`
	Temperature    float64 `json:"temperature"`
	MaxTokens      int     `json:"max_tokens"`
	Timeout        int     `json:"timeout"`
	CacheEnabled   bool    `json:"cache_enabled"`
	CacheTTL       int     `json:"cache_ttl"`
}

type TradingConfig struct {
	Enabled              bool     `json:"enabled"`
	Mode                 string   `json:"mode"`
	AllowedStrategies    []string `json:"allowed_strategies"`
	AllowedMarkets       []string `json:"allowed_markets"`
	BlockedMarkets       []string `json:"blocked_markets"`
	MinLiquidity         float64  `json:"min_liquidity"`
	MinVolume24h         float64  `json:"min_volume_24h"`
	MaxSlippage          float64  `json:"max_slippage"`
	AutoCompound         bool     `json:"auto_compound"`
	RebalanceThreshold   float64  `json:"rebalance_threshold"`
	ScanInterval         int      `json:"scan_interval"`
	MaxOpenPositions     int      `json:"max_open_positions"`
}

type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	File       string `json:"file"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
}

func LoadConfig() (*Config, error) {
	configPath := os.Getenv("POLYMARKET_CONFIG")
	if configPath == "" {
		configPath = "polymarket_config.json"
	}

	config := &Config{}

	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	config.loadFromEnv()

	config.validate()

	return config, nil
}

func (c *Config) loadFromEnv() {
	if apiKey := os.Getenv("POLYMARKET_API_KEY"); apiKey != "" {
		c.Polymarket.APIKey = apiKey
	}
	if apiSecret := os.Getenv("POLYMARKET_API_SECRET"); apiSecret != "" {
		c.Polymarket.APISecret = apiSecret
	}
	if passphrase := os.Getenv("POLYMARKET_PASSPHRASE"); passphrase != "" {
		c.Polymarket.Passphrase = passphrase
	}

	if timeout := os.Getenv("POLYMARKET_TIMEOUT"); timeout != "" {
		if val, err := strconv.Atoi(timeout); err == nil {
			c.Polymarket.Timeout = val
		}
	}

	if provider := os.Getenv("AI_PROVIDER"); provider != "" {
		c.AI.Provider = provider
	}
	if model := os.Getenv("AI_MODEL"); model != "" {
		c.AI.Model = model
	}
	if apiKey := os.Getenv("AI_API_KEY"); apiKey != "" {
		c.AI.APIKey = apiKey
	}
	if baseURL := os.Getenv("AI_BASE_URL"); baseURL != "" {
		c.AI.BaseURL = baseURL
	}

	if enabled := os.Getenv("TRADING_ENABLED"); enabled != "" {
		c.Trading.Enabled = enabled == "true" || enabled == "1"
	}
	if mode := os.Getenv("TRADING_MODE"); mode != "" {
		c.Trading.Mode = mode
	}

	if totalCapital := os.Getenv("RISK_TOTAL_CAPITAL"); totalCapital != "" {
		if val, err := strconv.ParseFloat(totalCapital, 64); err == nil {
			c.Risk.TotalCapital = val
		}
	}
	if maxPosition := os.Getenv("RISK_MAX_POSITION"); maxPosition != "" {
		if val, err := strconv.ParseFloat(maxPosition, 64); err == nil {
			c.Risk.MaxPositionSize = val
		}
	}
	if dailyLoss := os.Getenv("RISK_DAILY_LOSS_LIMIT"); dailyLoss != "" {
		if val, err := strconv.ParseFloat(dailyLoss, 64); err == nil {
			c.Risk.DailyLossLimit = val
		}
	}

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		c.Logging.Level = level
	}
	if output := os.Getenv("LOG_OUTPUT"); output != "" {
		c.Logging.Output = output
	}
}

func (c *Config) validate() error {
	if c.Polymarket.Timeout == 0 {
		c.Polymarket.Timeout = 30
	}
	if c.Polymarket.RetryCount == 0 {
		c.Polymarket.RetryCount = 3
	}
	if c.Polymarket.RetryWait == 0 {
		c.Polymarket.RetryWait = 1
	}
	if c.Polymarket.Environment == "" {
		c.Polymarket.Environment = "production"
	}

	if c.AI.Provider == "" {
		c.AI.Provider = "qwen"
	}
	if c.AI.Model == "" {
		c.AI.Model = "qwen-plus"
	}
	if c.AI.Temperature == 0 {
		c.AI.Temperature = 0.3
	}
	if c.AI.MaxTokens == 0 {
		c.AI.MaxTokens = 2000
	}
	if c.AI.Timeout == 0 {
		c.AI.Timeout = 30
	}
	if c.AI.CacheTTL == 0 {
		c.AI.CacheTTL = 1800
	}

	if c.Trading.Mode == "" {
		c.Trading.Mode = "live"
	}
	if len(c.Trading.AllowedStrategies) == 0 {
		c.Trading.AllowedStrategies = []string{"VALUE_BET", "ARBITRAGE", "MOMENTUM"}
	}
	if c.Trading.MinLiquidity == 0 {
		c.Trading.MinLiquidity = 5000
	}
	if c.Trading.MinVolume24h == 0 {
		c.Trading.MinVolume24h = 10000
	}
	if c.Trading.MaxSlippage == 0 {
		c.Trading.MaxSlippage = 0.05
	}
	if c.Trading.ScanInterval == 0 {
		c.Trading.ScanInterval = 60
	}
	if c.Trading.MaxOpenPositions == 0 {
		c.Trading.MaxOpenPositions = 10
	}

	if c.Risk.TotalCapital == 0 {
		c.Risk.TotalCapital = 10000
	}
	if c.Risk.MaxPositionSize == 0 {
		c.Risk.MaxPositionSize = 500
	}
	if c.Risk.MaxPositionPercent == 0 {
		c.Risk.MaxPositionPercent = 0.05
	}
	if c.Risk.DailyLossLimit == 0 {
		c.Risk.DailyLossLimit = 500
	}
	if c.Risk.MaxDrawdown == 0 {
		c.Risk.MaxDrawdown = 0.15
	}
	if c.Risk.KellyMultiplier == 0 {
		c.Risk.KellyMultiplier = 0.25
	}
	if c.Risk.MinConfidence == 0 {
		c.Risk.MinConfidence = 50.0
	}

	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}

	return nil
}

func (c *Config) SaveToFile(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func DefaultConfig() *Config {
	return &Config{
		Polymarket: PolymarketConfig{
			Timeout:     30,
			RetryCount:  3,
			RetryWait:   1,
			Environment: "production",
		},
		Risk: RiskConfig{
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
		},
		AI: AIConfig{
			Provider:     "qwen",
			Model:        "qwen-plus",
			Temperature:  0.3,
			MaxTokens:    2000,
			Timeout:      30,
			CacheEnabled: true,
			CacheTTL:     1800,
		},
		Trading: TradingConfig{
			Enabled:            false,
			Mode:               "paper",
			MinLiquidity:       5000,
			MinVolume24h:       10000,
			MaxSlippage:        0.05,
			AutoCompound:       false,
			ScanInterval:       60,
			MaxOpenPositions:   10,
			AllowedStrategies:  []string{"VALUE_BET", "ARBITRAGE", "MOMENTUM"},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     30,
		},
	}
}

func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.Polymarket.Timeout) * time.Second
}

func (c *Config) GetRetryCount() int {
	return c.Polymarket.RetryCount
}

func (c *Config) GetRetryWait() time.Duration {
	return time.Duration(c.Polymarket.RetryWait) * time.Second
}

func (c *Config) IsPaperTrading() bool {
	return c.Trading.Mode == "paper" || !c.Trading.Enabled
}

func (c *Config) IsStrategyAllowed(strategy string) bool {
	for _, allowed := range c.Trading.AllowedStrategies {
		if allowed == strategy {
			return true
		}
	}
	return false
}

func (c *Config) IsMarketAllowed(marketID string) bool {
	for _, blocked := range c.Trading.BlockedMarkets {
		if blocked == marketID {
			return false
		}
	}

	if len(c.Trading.AllowedMarkets) == 0 {
		return true
	}

	for _, allowed := range c.Trading.AllowedMarkets {
		if allowed == marketID {
			return true
		}
	}

	return false
}
