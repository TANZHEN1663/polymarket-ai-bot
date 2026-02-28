# Polymarket AI 交易机器人 - 架构文档

## 📁 项目结构

```
polymarket/
├── client.go              # API 客户端核心（Gamma、Data、CLOB）
├── gamma.go               # Gamma API 服务（市场数据发现）
├── data.go                # Data API 服务（用户数据和持仓）
├── clob.go                # CLOB API 服务（订单簿和交易）
├── analyzer.go            # 市场数据分析引擎
├── predictor.go           # AI 预测模型
├── strategy.go            # 交易策略引擎
├── risk.go                # 风险管理系统
├── config.go              # 配置管理
├── monitor.go             # 监控和日志系统
├── bot.go                 # 机器人主引擎
├── main.go                # 程序入口
├── polymarket_test.go     # 单元测试
├── README.md              # 使用文档
├── ARCHITECTURE.md        # 架构文档（本文件）
├── start.sh               # Linux/Mac启动脚本
└── start.bat              # Windows 启动脚本
```

## 🏗️ 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────┐
│                    Polymarket AI Bot                     │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────┐  │
│  │   Gamma API  │    │   Data API   │    │ CLOB API │  │
│  │  (市场发现)   │    │  (用户数据)   │    │ (交易)   │  │
│  └──────┬───────┘    └──────┬───────┘    └────┬─────┘  │
│         │                   │                  │         │
│         └───────────────────┼──────────────────┘         │
│                             │                            │
│                    ┌────────▼────────┐                   │
│                    │   API Client    │                   │
│                    │   (client.go)   │                   │
│                    └────────┬────────┘                   │
│                             │                            │
│         ┌───────────────────┼───────────────────┐       │
│         │                   │                   │       │
│  ┌──────▼──────┐    ┌──────▼──────┐    ┌──────▼──────┐ │
│  │  Analyzer   │    │  Predictor  │    │  Strategy   │ │
│  │ (市场分析)   │    │  (AI 预测)   │    │  (策略引擎)  │ │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘ │
│         │                   │                   │       │
│         │            ┌──────▼──────┐           │       │
│         │            │   LLM API   │           │       │
│         │            │ (Qwen 等)   │           │       │
│         │            └─────────────┘           │       │
│         │                                      │       │
│         └──────────────────┬───────────────────┘       │
│                            │                           │
│                   ┌────────▼────────┐                  │
│                   │ Risk Manager    │                  │
│                   │ (风险管理)      │                  │
│                   └────────┬────────┘                  │
│                            │                           │
│         ┌──────────────────┼──────────────────┐        │
│         │                  │                  │        │
│  ┌──────▼──────┐   ┌───────▼───────┐  ┌──────▼──────┐ │
│  │   Monitor   │   │ Trade Executor│  │  Position   │ │
│  │  (监控日志)  │   │  (交易执行)    │  │  Manager    │ │
│  └─────────────┘   └───────────────┘  └─────────────┘ │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

## 🔄 数据流

### 1. 市场扫描流程

```
定时器触发 (每 60 秒)
    │
    ▼
Gamma API - ListMarkets
    │
    ▼
过滤 (流动性、交易量)
    │
    ▼
对每个市场执行:
    ├─► Analyzer.AnalyzeMarket()
    │   ├─ 获取订单簿
    │   ├─ 获取价格
    │   ├─ 流动性分析
    │   ├─ 订单簿不平衡分析
    │   └─ 计算公平价值
    │
    ├─► Predictor.Predict()
    │   ├─ 构建 AI 提示词
    │   ├─ 调用 LLM
    │   ├─ 解析预测结果
    │   └─ 计算置信度
    │
    └─► 生成交易信号
        │
        ▼
    风险评估
        │
        ▼
    执行交易 (如果启用)
```

### 2. 交易执行流程

```
交易信号
    │
    ▼
RiskManager.CheckTradeLimits()
    ├─ 检查置信度
    ├─ 检查仓位限制
    ├─ 检查可用资金
    └─ 检查市场暴露
    │
    ▼
Strategy.Execute()
    ├─ 计算仓位大小 (Kelly 公式)
    ├─ 创建订单
    └─ 提交到 CLOB API
    │
    ▼
更新持仓记录
    │
    ▼
记录交易日志
    │
    ▼
更新风险指标
```

### 3. 风险管理流程

```
持续监控
    │
    ├─► 每 10 秒检查:
    │   ├─ 当前回撤
    │   ├─ 日 PnL
    │   ├─ 持仓数量
    │   └─ 资金利用率
    │
    ├─► 触及阈值时:
    │   ├─ 发出告警
    │   ├─ 暂停交易
    │   └─ 通知用户
    │
    └─► 每日重置:
        ├─ 重置日 PnL
        ├─ 重置交易计数
        └─ 恢复交易 (如果因日亏损暂停)
```

## 📊 核心模块详解

### 1. API 客户端层 (client.go, gamma.go, data.go, clob.go)

**职责**: 
- 封装 Polymarket 三个 API 的所有端点
- 处理 HTTP 请求、重试、错误处理
- 提供类型安全的响应解析

**关键结构**:
```go
type Client struct {
    gammaClient  *resty.Client
    dataClient   *resty.Client
    clobClient   *resty.Client
}

type GammaService struct{ client *Client }
type DataService struct{ client *Client }
type CLOBService struct{ client *Client }
```

### 2. 市场分析引擎 (analyzer.go)

**职责**:
- 订单簿分析（流动性、买卖压力）
- 交易量分析（趋势、活动级别）
- 价格形态识别
- 公平价值计算

**关键方法**:
```go
func (ma *MarketAnalyzer) AnalyzeMarket() *MarketOpportunity
func (ma *MarketAnalyzer) AnalyzeLiquidity() *LiquidityAnalysis
func (ma *MarketAnalyzer) AnalyzeOrderBookImbalance() *OrderBookImbalance
func (ma *MarketAnalyzer) CalculateFairValue() float64
```

### 3. AI 预测模型 (predictor.go)

**职责**:
- 构建智能提示词（包含市场数据、历史趋势）
- 调用 LLM 进行概率预测
- 解析和校准 AI 输出
- 缓存预测结果（减少 API 调用）

**提示词结构**:
```
市场信息
  - 当前价格、隐含概率
  - 交易量、流动性
  - 到期时间

历史数据
  - 价格历史
  - 类似事件结果

外部因素
  - 新闻情绪
  - 社交媒体

分析要求
  - 预测结果 (YES/NO)
  - 概率评估 (0-100%)
  - 置信度
  - 推理过程
  - 关键因素
  - 风险评估
```

### 4. 交易策略引擎 (strategy.go)

**策略接口**:
```go
type TradingStrategy interface {
    Name() string
    Execute(ctx context.Context, signal *TradingSignal) (*TradeExecution, error)
    Validate(signal *TradingSignal) bool
}
```

**内置策略**:
- **ValueBetStrategy**: 价值投注
- **ArbitrageStrategy**: 套利
- **MomentumStrategy**: 动量
- **HedgeStrategy**: 对冲

### 5. 风险管理系统 (risk.go)

**核心功能**:
- Kelly 公式仓位计算
- 回撤监控和控制
- 日亏损限制
- 仓位分散限制
- 流动性风险评估

**关键指标**:
```go
type RiskMetrics struct {
    TotalCapital      float64  // 总资金
    CurrentCapital    float64  // 当前资金
    UsedCapital       float64  // 已用资金
    TotalPnL          float64  // 总盈亏
    DailyPnL          float64  // 日盈亏
    CurrentDrawdown   float64  // 当前回撤
    WinRate           float64  // 胜率
    SharpeRatio       float64  // 夏普比率
}
```

### 6. 监控系统 (monitor.go)

**职责**:
- 结构化日志记录
- 性能指标追踪
- 告警系统
- 事件记录

**日志级别**:
- DEBUG: 详细调试信息
- INFO: 一般信息（交易信号、订单执行）
- WARN: 警告（接近限制）
- ERROR: 错误（API 失败、执行失败）
- ALERT: 严重告警（触及风控阈值）

## 🔐 安全机制

### 1. API 密钥管理
- 密钥从配置文件或环境变量加载
- 不在日志中输出
- 支持加密存储（未来功能）

### 2. 交易保护
- 默认模拟模式（Paper Trading）
- 需要显式启用实盘模式
- 多层风险评估

### 3. 错误处理
- 所有 API 调用带重试机制
- 优雅的错误恢复
- 详细的错误日志

## 📈 性能优化

### 1. 缓存策略
- AI 预测结果缓存（30 分钟 TTL）
- 市场价格缓存（5 秒 TTL）
- 减少重复 API 调用

### 2. 并发控制
- 市场扫描使用 goroutine 池
- 限制并发 API 请求数
- 避免速率限制

### 3. 资源管理
- 上下文超时控制
- 优雅关闭
- 内存泄漏防护

## 🔧 配置系统

### 配置优先级
1. 配置文件（polymarket_config.json）
2. 环境变量
3. 默认值

### 热重载支持
部分配置支持运行时更新：
- 风险参数
- 交易开关
- 日志级别

## 🧪 测试策略

### 单元测试
- API 客户端模拟
- 策略逻辑验证
- 风险计算测试

### 集成测试
- 模拟市场扫描
- 端到端交易流程
- 风险评估验证

### 性能测试
- 并发市场扫描
- 延迟基准测试
- 内存使用分析

## 📦 部署选项

### 1. 本地运行
```bash
go run main.go -config=polymarket_config.json
```

### 2. Docker 部署（未来）
```dockerfile
FROM golang:1.21
COPY . /app
RUN go build -o bot .
CMD ["./bot"]
```

### 3. 云服务部署
- Railway
- Heroku
- AWS/GCP/Azure

## 🚀 扩展指南

### 添加新策略
1. 实现 `TradingStrategy` 接口
2. 在 `StrategyEngine` 注册
3. 添加到配置文件的 `allowed_strategies`

### 添加新数据源
1. 创建新的 Service 结构
2. 实现数据获取方法
3. 在 `Analyzer` 中集成

### 添加新 AI 提供商
1. 实现 `mcp.Client` 接口
2. 在 `createLLMClient` 中添加 case
3. 更新配置选项

## 📝 最佳实践

1. **始终从模拟模式开始**
2. **逐步增加资金**
3. **监控前几周的表现**
4. **定期审查和调整参数**
5. **保持日志记录**
6. **设置告警通知**
7. **定期备份配置**
8. **关注 Polymarket API 变更**

## 🆘 故障排查

### 常见问题
1. **API 密钥错误**: 检查配置文件格式
2. **连接超时**: 增加 timeout 配置
3. **交易未执行**: 检查 risk.CheckTradeLimits
4. **AI 预测失败**: 验证 API 密钥和网络

### 调试技巧
1. 启用 DEBUG 日志级别
2. 检查监控指标
3. 查看 API 响应原始数据
4. 使用 -status 命令查看状态

---

**版本**: 1.0.0  
**最后更新**: 2026-02-28  
**维护者**: Polymarket AI Bot Team
