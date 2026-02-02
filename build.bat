@echo off
REM Build script for TaskVault on Windows

setlocal enabledelayedexpansion

echo.
echo 🔨 TaskVault Build for Windows
echo ==============================

REM Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ Go 1.21+ required but not found
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
echo ✓ Go %GO_VERSION%

REM Download dependencies
echo 📦 Downloading dependencies...
go mod download
if errorlevel 1 (
    echo ❌ Dependency download failed
    exit /b 1
)

REM Run tests
echo 🧪 Running tests...
go test ./... -v -race
if errorlevel 1 (
    echo ⚠️  Some tests failed
)

REM Build CLI
echo 🏗️  Building CLI...
go build -o taskvault.exe ./cmd/taskvault
if errorlevel 1 (
    echo ❌ Build failed
    exit /b 1
)

echo ✓ Built: taskvault.exe
echo.
echo Build successful!
echo Run: taskvault.exe --help
