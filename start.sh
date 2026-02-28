#!/bin/bash

# Polymarket AI Trading Bot - Quick Start Script
# This script helps you set up and run the Polymarket trading bot

set -e

echo "🚀 Polymarket AI Trading Bot - Setup & Launch"
echo "=============================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    echo "   Download from: https://golang.org/dl/"
    exit 1
fi

echo "✅ Go version: $(go version)"
echo ""

# Check if config file exists
CONFIG_FILE="polymarket_config.json"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "📝 Creating configuration file..."
    if [ -f "polymarket_config.example.json" ]; then
        cp polymarket_config.example.json $CONFIG_FILE
        echo "✅ Created $CONFIG_FILE from example"
        echo ""
        echo "⚠️  IMPORTANT: Edit $CONFIG_FILE and add your API keys:"
        echo "   - Polymarket CLOB API credentials"
        echo "   - AI Provider API key (Qwen, Claude, etc.)"
        echo ""
        echo "After editing, run this script again."
        exit 0
    else
        echo "❌ Example config file not found!"
        exit 1
    fi
fi

# Check if dependencies are installed
echo "📦 Checking dependencies..."
if [ ! -f "go.mod" ]; then
    echo "❌ go.mod not found. Please run this script from the polymarket directory."
    exit 1
fi

go mod tidy
echo "✅ Dependencies installed"
echo ""

# Run tests (optional)
if [ "$1" == "--test" ]; then
    echo "🧪 Running tests..."
    go test -v ./...
    echo "✅ Tests completed"
    echo ""
fi

# Show configuration status
echo "📋 Configuration Status:"
echo "------------------------"
if grep -q '"api_key": ""' $CONFIG_FILE; then
    echo "⚠️  WARNING: Polymarket API key is empty!"
fi
if grep -q '"api_secret": ""' $CONFIG_FILE; then
    echo "⚠️  WARNING: Polymarket API secret is empty!"
fi
if grep -q '"api_key": ""' $CONFIG_FILE | grep -A5 '"ai":' $CONFIG_FILE | grep -q '"api_key": ""'; then
    echo "⚠️  WARNING: AI API key is empty!"
fi
echo ""

# Check trading mode
TRADING_MODE=$(grep -o '"mode": "[^"]*"' $CONFIG_FILE | cut -d'"' -f4)
echo "📊 Trading Mode: $TRADING_MODE"
if [ "$TRADING_MODE" == "live" ]; then
    echo "⚠️  WARNING: Running in LIVE mode with real money!"
    echo "   Press Ctrl+C now if you want to review settings first."
    sleep 3
else
    echo "✅ Running in PAPER mode (simulated trading)"
fi
echo ""

# Build and run
echo "🚀 Starting Polymarket AI Trading Bot..."
echo "========================================"
echo ""

go run main.go -config=$CONFIG_FILE
