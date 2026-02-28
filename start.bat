@echo off
REM Polymarket AI Trading Bot - Quick Start Script for Windows
REM This script helps you set up and run the Polymarket trading bot

echo.
echo ============================================
echo   Polymarket AI Trading Bot - Setup ^& Launch
echo ============================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Go is not installed. Please install Go 1.21+ first.
    echo Download from: https://golang.org/dl/
    pause
    exit /b 1
)

echo [OK] Go version:
go version
echo.

REM Check if config file exists
set CONFIG_FILE=polymarket_config.json
if not exist "%CONFIG_FILE%" (
    echo [INFO] Creating configuration file...
    if exist "polymarket_config.example.json" (
        copy polymarket_config.example.json %CONFIG_FILE%
        echo [OK] Created %CONFIG_FILE% from example
        echo.
        echo [IMPORTANT] Edit %CONFIG_FILE% and add your API keys:
        echo   - Polymarket CLOB API credentials
        echo   - AI Provider API key (Qwen, Claude, etc.)
        echo.
        echo After editing, run this script again.
        pause
        exit /b 0
    ) else if exist "..\polymarket_config.example.json" (
        copy ..\polymarket_config.example.json %CONFIG_FILE%
        echo [OK] Created %CONFIG_FILE% from example
        echo.
        echo [IMPORTANT] Edit %CONFIG_FILE% and add your API keys:
        echo   - Polymarket CLOB API credentials
        echo   - AI Provider API key (Qwen, Claude, etc.)
        echo.
        echo After editing, run this script again.
        pause
        exit /b 0
    ) else (
        echo [ERROR] Example config file not found!
        echo Please run this script from the polymarket directory.
        pause
        exit /b 1
    )
)

REM Check if dependencies are installed
echo [INFO] Checking dependencies...
if not exist "go.mod" (
    if exist "..\go.mod" (
        echo [INFO] Found go.mod in parent directory
    ) else (
        echo [ERROR] go.mod not found. Please run this script from the polymarket directory.
        pause
        exit /b 1
    )
)

call go mod tidy
echo [OK] Dependencies installed
echo.

REM Run tests if requested
if "%1"=="--test" (
    echo [INFO] Running tests...
    call go test -v ./...
    echo [OK] Tests completed
    echo.
)

REM Show configuration status
echo [INFO] Configuration Status:
echo ------------------------
findstr /C:"\"api_key\": \"\"" %CONFIG_FILE% >nul
if %ERRORLEVEL% EQU 0 (
    echo [WARNING] Polymarket API key is empty!
)
findstr /C:"\"api_secret\": \"\"" %CONFIG_FILE% >nul
if %ERRORLEVEL% EQU 0 (
    echo [WARNING] Polymarket API secret is empty!
)
echo.

REM Check trading mode
findstr /C:"\"mode\": \"live\"" %CONFIG_FILE% >nul
if %ERRORLEVEL% EQU 0 (
    echo [WARNING] Running in LIVE mode with real money!
    echo    Press Ctrl+C now if you want to review settings first.
    timeout /t 3 /nobreak >nul
) else (
    echo [OK] Running in PAPER mode (simulated trading)
)
echo.

REM Build and run
echo [INFO] Starting Polymarket AI Trading Bot...
echo ========================================
echo.

call go run main.go -config=%CONFIG_FILE%
