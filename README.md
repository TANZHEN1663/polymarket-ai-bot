# Polymarket AI 交易机器人

一个基于人工智能的 Polymarket 预测市场自动化交易系统，具有高胜率的交易策略和严格的风险管理。

## 🎯 核心特性

### 1. **多策略交易引擎**
- **价值投注 (Value Betting)**: 识别市场定价错误，当 AI 预测概率与市场隐含概率存在显著差异时入场
- **套利策略 (Arbitrage)**: 利用订单簿不平衡和市场价差进行套利
- **动量策略 (Momentum)**: 基于价格趋势和交易量变化捕捉动量机会
- **对冲策略 (Hedging)**: 自动对冲现有仓位，降低风险敞口

### 2. **AI 驱动的概率预测**
- 集成多种 LLM 模型（Qwen、Claude、DeepSeek、OpenAI 等）
- 基于市场数据、历史趋势、外部因素进行综合概率评估
- 智能推理和关键因素分析
- 概率校准和置信度评估

### 3. **严格的风险管理**
- **Kelly 公式**优化仓位大小
- **最大回撤控制**（默认 15%）
- **每日亏损限制**（默认 500 USDC）
- **仓位分散限制**（单一市场不超过 20%）
- **流动性风险评估**（避免高滑点）
- **置信度阈值**（最低 50% 置信度）

### 4. **实时市场监控**
- 订单簿深度分析
- 交易量趋势检测
- 流动性评分系统
- 价格形态识别
- 市场情绪分析

### 5. **全面的监控和日志**
- 实时性能指标追踪
- 交易历史记录
- 风险指标监控
- 自动告警系统
- 结构化日志输出

## 📦 安装

### 前置要求
- Go 1.21+
- Polymarket CLOB API 密钥（https://clob.polymarket.com）
- LLM API 密钥（推荐使用 Qwen 或 DeepSeek）

### 1. 克隆项目
```bash
git clone <your-repo-url>
cd polymarket
```

### 2. 安装依赖
```bash
go mod download
```

### 3. 配置文件
复制并编辑配置文件：
```bash
cp polymarket_config.example.json polymarket_config.json
```

编辑 `polymarket_config.json`，填入你的 API 密钥：
```json
{
  "polymarket": {
    "api_key": "你的 Polymarket API Key",
    "api_secret": "你的 Polymarket API Secret",
    "passphrase": "你的 Polymarket Passphrase"
  },
  "ai": {
    "provider": "deepseek",
    "model": "deepseek-chat",
    "api_key": "你的 DeepSeek API Key"
  },
  "trading": {
    "enabled": false,
    "mode": "paper"
  }
}
```

### 4. 环境变量（可选）
```bash
export POLYMARKET_API_KEY="your-api-key"
export POLYMARKET_API_SECRET="your-api-secret"
export POLYMARKET_PASSPHRASE="your-passphrase"
export AI_API_KEY="your-llm-api-key"
```

## 🚀 使用方法

### 启动机器人
```bash
cd cmd
go run main.go -config=../polymarket_config.json
```

### 命令行选项
```bash
# 查看状态
go run main.go -status

# 查看交易指标
go run main.go -metrics

# 查看持仓
go run main.go -positions

# 暂停交易
go run main.go -pause

# 恢复交易
go run main.go -resume
```

## 📊 策略详解

### 价值投注策略 (VALUE_BET)
**原理**: 当 AI 预测的事件发生概率与市场隐含概率存在显著差异时下注。

**入场条件**:
- AI 概率与市场隐含概率差异 > 5%
- 置信度 > 50%
- 流动性 > 5000 USDC
- 24 小时交易量 > 10000 USDC

**仓位计算**: 使用 Kelly 公式的 25%（保守版本）

### 套利策略 (ARBITRAGE)
**原理**: 利用订单簿不平衡或相关市场间的价差进行套利。

**入场条件**:
- 订单簿不平衡度 > 20%
- 预期价值 > 3%
- 买卖价差足够覆盖手续费

### 动量策略 (MOMENTUM)
**原理**: 跟随市场趋势，在强势突破时入场。

**入场条件**:
- 价格趋势强度 > 2%
- 交易量放大 > 20%
- 置信度 > 60%

### 对冲策略 (HEDGE)
**原理**: 当现有仓位盈利达到一定比例时，自动对冲锁定利润。

**触发条件**:
- 仓位盈利 > 50%
- 市场出现反转信号

## ⚙️ 配置说明

### 风险管理配置
```json
{
  "risk": {
    "total_capital": 10000,        // 总资金 (USDC)
    "max_position_size": 500,      // 单一仓位最大规模
    "max_position_percent": 0.05,  // 单一仓位最大占比 (5%)
    "daily_loss_limit": 500,       // 每日最大亏损
    "max_drawdown": 0.15,          // 最大回撤 (15%)
    "kelly_multiplier": 0.25,      // Kelly 系数 (0.25=保守)
    "min_confidence": 50.0,        // 最低置信度 (%)
    "max_liquidity_slippage": 0.05 // 最大滑点 (5%)
  }
}
```

### AI 配置
```json
{
  "ai": {
    "provider": "deepseek",         // AI 提供商
    "model": "deepseek-chat",       // 模型名称
    "temperature": 0.3,             // 温度参数 (低=保守)
    "max_tokens": 2000,             // 最大输出长度
    "cache_enabled": true,          // 启用缓存
    "cache_ttl": 1800               // 缓存时间 (秒)
  }
}
```

### 交易配置
```json
{
  "trading": {
    "enabled": false,               // 是否启用交易
    "mode": "paper",                // 模式：paper/live
    "scan_interval": 60,            // 市场扫描间隔 (秒)
    "max_open_positions": 10,       // 最大持仓数量
    "min_liquidity": 5000,          // 最小流动性
    "min_volume_24h": 10000,        // 最小 24h 交易量
    "allowed_strategies": [         // 启用的策略
      "VALUE_BET",
      "ARBITRAGE",
      "MOMENTUM"
    ]
  }
}
```

## 📈 性能指标

### 关键指标说明
- **Win Rate (胜率)**: 盈利交易占比
- **Sharpe Ratio (夏普比率)**: 风险调整后收益
- **Max Drawdown (最大回撤)**: 历史最大亏损幅度
- **Profit Factor (盈利因子)**: 总盈利/总亏损
- **Expected Value (期望值)**: 平均每笔交易的预期收益

### 查看实时指标
```bash
go run main.go -metrics
```

## 🛡️ 风险管理

### 自动风控措施
1. **交易暂停**: 触及日亏损限制时自动停止
2. **回撤控制**: 触及最大回撤时清仓并停止
3. **仓位限制**: 单一市场不超过总资金的 20%
4. **流动性筛选**: 避开低流动性市场（高滑点风险）
5. **置信度过滤**: 只交易高置信度机会

### 手动控制
```bash
# 紧急暂停
go run main.go -pause

# 恢复交易
go run main.go -resume
```

## 📝 日志示例

```json
{"timestamp":"2026-02-28T12:00:00Z","level":"INFO","message":"Trading signal: VALUE_BET","module":"signal","metadata":{"strength":0.12,"confidence":75,"direction":"YES"}}
{"timestamp":"2026-02-28T12:00:01Z","level":"INFO","message":"Order executed: ORD123456","module":"executor","metadata":{"market_id":"MKT789","strategy":"VALUE_BET","side":"BUY","size":250,"price":0.65}}
```

## ⚠️ 风险提示

1. **预测市场风险**: Polymarket 是高风险的去中心化预测市场，可能导致本金全部损失
2. **模型风险**: AI 预测不保证准确性，历史表现不代表未来
3. **技术风险**: API 故障、网络延迟等技术问题可能导致损失
4. **流动性风险**: 低流动性市场可能无法及时平仓
5. **监管风险**: 预测市场在某些司法管辖区可能不受监管

**重要**: 只使用你能承受损失的资金进行交易！

## 📚 API 参考

### Polymarket API
- **Gamma API**: https://gamma-api.polymarket.com - 市场数据
- **Data API**: https://data-api.polymarket.com - 用户数据
- **CLOB API**: https://clob.polymarket.com - 订单簿和交易

### 文档
- 官方文档：https://docs.polymarket.com/api-reference/introduction

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 MIT 许可证。

## 📞 支持

如有问题，请提交 Issue 或联系开发者。

---

**免责声明**: 本软件仅供教育和研究目的，不构成投资建议。使用本软件进行交易存在重大风险，用户需自行承担所有风险。
