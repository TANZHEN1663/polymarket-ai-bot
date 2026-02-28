# 🚀 Polymarket AI 交易机器人 - 快速启动指南

## ✅ 配置文件已创建

配置文件 `polymarket_config.json` 已经创建在你的 polymarket 目录中！

## 📝 下一步操作

### 1. 编辑配置文件

打开 `polymarket_config.json` 文件，填入你的 API 密钥：

```json
{
  "polymarket": {
    "api_key": "你的 Polymarket API Key",
    "api_secret": "你的 Polymarket API Secret",
    "passphrase": "你的 Polymarket Passphrase"
  },
  "ai": {
    "provider": "qwen",
    "model": "qwen-plus",
    "api_key": "你的 Qwen API Key"
  }
}
```

### 2. 获取 API 密钥

#### Polymarket API 密钥
1. 访问：https://clob.polymarket.com
2. 注册/登录账户
3. 在设置中获取 API Key、Secret 和 Passphrase

#### AI API 密钥（推荐使用 Qwen）
1. 访问阿里云百炼平台
2. 注册并创建 API Key
3. 选择 Qwen-plus 模型

### 3. 运行机器人

**方式一：直接运行（推荐）**
```powershell
cd c:\Users\a5414\nofx\polymarket
go run main.go
```

**方式二：使用启动脚本**
```powershell
cd c:\Users\a5414\nofx\polymarket
.\start.bat
```

## 🎯 运行模式说明

### 模拟模式（新手推荐）
配置文件中保持以下设置：
```json
{
  "trading": {
    "enabled": false,
    "mode": "paper"
  }
}
```
- ✅ 不会实际下单
- ✅ 只生成交易信号
- ✅ 用于测试和观察

### 实盘模式（有经验用户）
```json
{
  "trading": {
    "enabled": true,
    "mode": "live"
  }
}
```
- ⚠️ 会使用真实资金交易
- ⚠️ 确保已充分测试
- ⚠️ 建议从小资金开始

## 📊 常用命令

```powershell
# 查看机器人状态
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

## 🔧 默认配置说明

### 风险管理（保守配置）
- 总资金：10,000 USDC
- 单一仓位最大：500 USDC (5%)
- 每日最大亏损：500 USDC
- 最大回撤：15%
- Kelly 系数：0.25 (保守)
- 最低置信度：50%

### 交易筛选
- 最小流动性：5,000 USDC
- 最小 24h 交易量：10,000 USDC
- 扫描间隔：60 秒
- 最大持仓数量：10 个

## 📈 预期表现

### 模拟测试建议
1. **第 1 周**: 观察交易信号，验证策略逻辑
2. **第 2 周**: 调整参数，优化配置
3. **第 3 周**: 小资金实盘测试（1000-5000 USDC）
4. **第 4 周**: 根据表现逐步增加资金

### 关键指标目标
- 胜率：> 60%
- 夏普比率：> 1.5
- 最大回撤：< 15%
- 盈利因子：> 2.0

## ⚠️ 重要提示

1. **始终从模拟模式开始** - 至少测试 1-2 周
2. **使用可承受损失的资金** - 预测市场风险很高
3. **密切监控** - 每天检查交易记录和指标
4. **定期审查** - 每周回顾表现，调整参数
5. **保持学习** - 阅读文档了解策略原理

## 🛡️ 安全特性

### 自动风控
- ✅ 触及日亏损限制自动停止
- ✅ 触及最大回撤自动清仓
- ✅ 单一市场仓位限制
- ✅ 流动性筛选（避免高滑点）
- ✅ 置信度过滤

### 手动控制
- 随时可以暂停/恢复交易
- 紧急情况下立即停止

## 📚 学习资源

### 项目文档
- [README.md](README.md) - 完整使用指南
- [ARCHITECTURE.md](ARCHITECTURE.md) - 架构设计
- 代码注释 - 每个模块都有详细说明

### Polymarket 官方
- API 文档：https://docs.polymarket.com
- Gamma API: https://gamma-api.polymarket.com
- CLOB API: https://clob.polymarket.com

## 🆘 常见问题

### Q: 需要多少资金开始？
A: 建议：
- 模拟测试：无要求
- 实盘测试：1000-5000 USDC
- 正式运行：10000+ USDC

### Q: 如何知道机器人是否正常工作？
A: 检查以下几点：
1. 日志中有市场扫描信息
2. 生成交易信号
3. 风险指标正常
4. 使用 -status 命令查看状态

### Q: 多久能看到收益？
A: 这取决于：
- 市场条件
- 配置参数
- 风险偏好
建议至少观察 1 个月再做评估

### Q: 如何优化策略？
A: 可以调整：
1. AI 模型和温度参数
2. 风险配置（仓位大小、Kelly 系数）
3. 交易筛选条件（流动性、交易量）
4. 启用/禁用特定策略

## 🎉 开始你的交易之旅

现在你已经准备好了！记住：
1. 耐心测试
2. 谨慎投资
3. 持续学习
4. 享受过程

**祝你好运！** 🚀

---

**免责声明**: 本软件仅供教育和研究目的，不构成投资建议。预测市场存在高风险，可能导致本金损失。请谨慎交易，只使用你能承受损失的资金。
