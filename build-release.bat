@echo off
REM Cross-platform build script for TaskVault (Windows)
REM Generates binaries for Windows, Linux, and macOS

setlocal

set VERSION=0.1.0
set BUILD_DIR=dist
set BINARY_NAME=taskvault

echo.
echo 🚀 TaskVault Multi-Platform Build
echo ==================================
echo Version: %VERSION%
echo.

REM Create build directory
if not exist "%BUILD_DIR%" mkdir "%BUILD_DIR%"

REM Build for Windows (amd64)
echo 🔨 Building for Windows (amd64)...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-X main.version=%VERSION%" -o "%BUILD_DIR%\%BINARY_NAME%-windows-amd64.exe" .\cmd\taskvault
echo ✓ %BUILD_DIR%\%BINARY_NAME%-windows-amd64.exe

REM Build for Linux (amd64)
echo 🔨 Building for Linux (amd64)...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-X main.version=%VERSION%" -o "%BUILD_DIR%\%BINARY_NAME%-linux-amd64" .\cmd\taskvault
echo ✓ %BUILD_DIR%\%BINARY_NAME%-linux-amd64

REM Build for macOS (amd64)
echo 🔨 Building for macOS (Intel)...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-X main.version=%VERSION%" -o "%BUILD_DIR%\%BINARY_NAME%-darwin-amd64" .\cmd\taskvault
echo ✓ %BUILD_DIR%\%BINARY_NAME%-darwin-amd64

REM Build for macOS (arm64)
echo 🔨 Building for macOS (Apple Silicon)...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags "-X main.version=%VERSION%" -o "%BUILD_DIR%\%BINARY_NAME%-darwin-arm64" .\cmd\taskvault
echo ✓ %BUILD_DIR%\%BINARY_NAME%-darwin-arm64

echo.
echo ✅ Build complete! Binaries in %BUILD_DIR%\
echo.
dir /B "%BUILD_DIR%"

endlocal
