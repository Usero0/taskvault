#!/bin/bash
# verify-project-structure.sh - Validate TaskVault project completeness

set -e

echo "🔍 TaskVault Project Structure Verification"
echo "==========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

passed=0
failed=0

check_file() {
    local file=$1
    local description=$2
    
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC} $description"
        ((passed++))
    else
        echo -e "${RED}✗${NC} $description (missing: $file)"
        ((failed++))
    fi
}

check_dir() {
    local dir=$1
    local description=$2
    
    if [ -d "$dir" ]; then
        echo -e "${GREEN}✓${NC} $description"
        ((passed++))
    else
        echo -e "${RED}✗${NC} $description (missing: $dir)"
        ((failed++))
    fi
}

echo "Core Files:"
check_file "go.mod" "Go module definition"
check_file "go.sum" "Go dependency checksums"
check_file "README.md" "Main documentation"
check_file "LICENSE" "MIT license"
check_file ".gitignore" "Git ignore rules"
check_file "Makefile" "Development Makefile"
check_file "build.sh" "Unix build script"
check_file "build.bat" "Windows build script"

echo ""
echo "Documentation:"
check_file "ARCHITECTURE.md" "Architecture & design document"
check_file "ROADMAP.md" "Product roadmap"
check_file "CONTRIBUTING.md" "Contribution guidelines"
check_file "BUSINESS_STRATEGY.md" "Business & SaaS strategy"

echo ""
echo "Source Code Directories:"
check_dir "cmd/taskvault" "CLI entry point"
check_dir "internal/hash" "Hash engine"
check_dir "internal/storage" "Storage layer"
check_dir "internal/cache" "Cache manager"
check_dir "internal/audit" "Audit logger"
check_dir "internal/config" "Configuration"
check_dir "pkg/sdk" "Go SDK"
check_dir "examples" "Example code"

echo ""
echo "Source Code Files:"
check_file "cmd/taskvault/main.go" "CLI main"
check_file "internal/hash/engine.go" "Hash engine implementation"
check_file "internal/storage/store.go" "Storage implementation"
check_file "internal/cache/manager.go" "Cache manager implementation"
check_file "internal/audit/logger.go" "Audit logger implementation"
check_file "internal/config/config.go" "Configuration"
check_file "pkg/sdk/client.go" "Go SDK client"

echo ""
echo "Tests & Examples:"
check_file "internal/hash/engine_test.go" "Hash engine tests"
check_file "examples/sdk_examples.go" "SDK examples"
check_file "examples/integration_example.go" "Integration example"
check_file "cmd/examples/main.go" "CLI examples"

echo ""
echo "CI/CD & Configuration:"
check_file ".taskvault/config.example.yaml" "Example configuration"
check_file ".github/workflows/test.yml" "GitHub Actions workflow"

echo ""
echo "========================================="
echo -e "Passed: ${GREEN}$passed${NC} | Failed: ${RED}$failed${NC}"
echo "========================================="

if [ $failed -eq 0 ]; then
    echo -e "${GREEN}✓ Project structure complete!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. cd taskvault"
    echo "2. go mod tidy"
    echo "3. go build ./cmd/taskvault"
    echo "4. ./taskvault --help"
    exit 0
else
    echo -e "${RED}✗ Project structure incomplete${NC}"
    exit 1
fi
