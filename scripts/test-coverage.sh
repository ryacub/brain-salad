#!/bin/bash

set -e

echo "================================"
echo "Brain Salad Test Coverage Report"
echo "================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create coverage directory
COVERAGE_DIR="coverage"
mkdir -p "$COVERAGE_DIR"

echo -e "${BLUE}Step 1: Running all tests with coverage...${NC}"
go test ./... -coverprofile="$COVERAGE_DIR/coverage.out" -covermode=atomic

echo ""
echo -e "${BLUE}Step 2: Generating HTML coverage report...${NC}"
go tool cover -html="$COVERAGE_DIR/coverage.out" -o "$COVERAGE_DIR/coverage.html"

echo ""
echo -e "${BLUE}Step 3: Analyzing coverage by package...${NC}"
echo ""
go tool cover -func="$COVERAGE_DIR/coverage.out" | tee "$COVERAGE_DIR/coverage-summary.txt"

echo ""
echo -e "${BLUE}Step 4: Overall coverage summary...${NC}"
TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_DIR/coverage.out" | tail -n 1 | awk '{print $NF}')
echo -e "Total Coverage: ${GREEN}$TOTAL_COVERAGE${NC}"

# Extract percentage as number for comparison
COVERAGE_NUM=$(echo "$TOTAL_COVERAGE" | sed 's/%//')

echo ""
echo -e "${BLUE}Step 5: Coverage by critical packages...${NC}"
echo ""

# Analyze coverage for key packages
PACKAGES=(
    "internal/llm"
    "internal/llm/processing"
    "internal/llm/quality"
    "internal/llm/cache"
    "internal/llm/client"
    "internal/scoring"
    "internal/api"
)

for pkg in "${PACKAGES[@]}"; do
    PKG_COVERAGE=$(go tool cover -func="$COVERAGE_DIR/coverage.out" | grep "$pkg" | grep -v "test" | awk '{sum+=$NF; count++} END {if(count>0) print sum/count; else print 0}')

    if [ ! -z "$PKG_COVERAGE" ]; then
        # Color code based on coverage percentage
        if (( $(echo "$PKG_COVERAGE >= 90" | bc -l) )); then
            COLOR=$GREEN
        elif (( $(echo "$PKG_COVERAGE >= 70" | bc -l) )); then
            COLOR=$YELLOW
        else
            COLOR=$RED
        fi

        printf "  %-35s ${COLOR}%.1f%%${NC}\n" "$pkg:" "$PKG_COVERAGE"
    fi
done

echo ""
echo -e "${BLUE}Step 6: Identifying untested files...${NC}"
echo ""

# Find files with 0% coverage
UNTESTED=$(go tool cover -func="$COVERAGE_DIR/coverage.out" | grep "0.0%" || true)

if [ ! -z "$UNTESTED" ]; then
    echo -e "${YELLOW}Files with 0% coverage:${NC}"
    echo "$UNTESTED"
else
    echo -e "${GREEN}All files have some test coverage!${NC}"
fi

echo ""
echo -e "${BLUE}Step 7: Coverage goals check...${NC}"
echo ""

# Check if we meet the 90% target
if (( $(echo "$COVERAGE_NUM >= 90" | bc -l) )); then
    echo -e "${GREEN}✓ Coverage goal met: $TOTAL_COVERAGE >= 90%${NC}"
    EXIT_CODE=0
elif (( $(echo "$COVERAGE_NUM >= 80" | bc -l) )); then
    echo -e "${YELLOW}⚠ Coverage approaching goal: $TOTAL_COVERAGE (target: 90%)${NC}"
    EXIT_CODE=0
else
    echo -e "${RED}✗ Coverage below goal: $TOTAL_COVERAGE < 90%${NC}"
    EXIT_CODE=1
fi

echo ""
echo -e "${BLUE}Results saved to:${NC}"
echo "  - HTML Report:    $COVERAGE_DIR/coverage.html"
echo "  - Coverage Data:  $COVERAGE_DIR/coverage.out"
echo "  - Summary:        $COVERAGE_DIR/coverage-summary.txt"

echo ""
echo "To view the HTML report, run:"
echo "  open $COVERAGE_DIR/coverage.html"
echo ""

echo -e "${GREEN}Coverage analysis complete!${NC}"

exit $EXIT_CODE
