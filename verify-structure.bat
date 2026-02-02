@echo off
REM verify-project-structure.bat - Validate TaskVault project completeness (Windows)

setlocal enabledelayedexpansion

echo.
echo 🔍 TaskVault Project Structure Verification
echo ===========================================
echo.

set passed=0
set failed=0

REM Function-like macro to check files
setlocal enabledelayedexpansion

for %%F in (
    "go.mod|Go module definition"
    "go.sum|Go dependency checksums"
    "README.md|Main documentation"
    "LICENSE|MIT license"
    ".gitignore|Git ignore rules"
    "Makefile|Development Makefile"
    "build.sh|Unix build script"
    "build.bat|Windows build script"
) do (
    for /f "tokens=1,* delims=|" %%A in ("%%F") do (
        if exist "%%A" (
            echo ✓ %%B
            set /a passed+=1
        ) else (
            echo ✗ %%B (missing: %%A)
            set /a failed+=1
        )
    )
)

echo.
echo Documentation:
for %%F in (
    "ARCHITECTURE.md|Architecture & design document"
    "ROADMAP.md|Product roadmap"
    "CONTRIBUTING.md|Contribution guidelines"
    "BUSINESS_STRATEGY.md|Business & SaaS strategy"
) do (
    for /f "tokens=1,* delims=|" %%A in ("%%F") do (
        if exist "%%A" (
            echo ✓ %%B
            set /a passed+=1
        ) else (
            echo ✗ %%B (missing: %%A)
            set /a failed+=1
        )
    )
)

echo.
echo Source Code Directories:
for %%D in (
    "cmd\taskvault|CLI entry point"
    "internal\hash|Hash engine"
    "internal\storage|Storage layer"
    "internal\cache|Cache manager"
    "internal\audit|Audit logger"
    "internal\config|Configuration"
    "pkg\sdk|Go SDK"
    "examples|Example code"
) do (
    for /f "tokens=1,* delims=|" %%A in ("%%D") do (
        if exist "%%A\" (
            echo ✓ %%B
            set /a passed+=1
        ) else (
            echo ✗ %%B (missing: %%A)
            set /a failed+=1
        )
    )
)

echo.
echo =========================================
echo Passed: %passed% / Failed: %failed%
echo =========================================

if %failed% equ 0 (
    echo.
    echo ✓ Project structure complete!
    echo.
    echo Next steps:
    echo 1. cd taskvault
    echo 2. go mod tidy
    echo 3. go build ./cmd/taskvault
    echo 4. taskvault --help
    exit /b 0
) else (
    echo.
    echo ✗ Project structure incomplete
    exit /b 1
)
