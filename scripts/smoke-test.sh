#!/bin/bash
#
# Smoke Test Script for Telos Idea Matrix
# Tests all critical functionality after deployment
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
WEB_URL="${WEB_URL:-http://localhost:3000}"
TIMEOUT=10
FAILED_TESTS=0
PASSED_TESTS=0

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
    ((PASSED_TESTS++))
}

log_fail() {
    echo -e "${RED}[✗]${NC} $1"
    ((FAILED_TESTS++))
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Test functions
test_api_health() {
    log_info "Testing API health endpoint..."

    if curl -f -s --max-time "$TIMEOUT" "$API_URL/health" > /dev/null 2>&1; then
        log_success "API health check passed"
    else
        log_fail "API health check failed"
    fi
}

test_api_version() {
    log_info "Testing API version endpoint..."

    response=$(curl -s --max-time "$TIMEOUT" "$API_URL/version" 2>/dev/null)

    if echo "$response" | grep -q "version"; then
        log_success "API version endpoint working"
    else
        log_fail "API version endpoint failed"
    fi
}

test_frontend_health() {
    log_info "Testing frontend accessibility..."

    if curl -f -s --max-time "$TIMEOUT" "$WEB_URL" > /dev/null 2>&1; then
        log_success "Frontend is accessible"
    else
        log_fail "Frontend is not accessible"
    fi
}

test_api_ideas_list() {
    log_info "Testing API ideas list endpoint..."

    response=$(curl -s --max-time "$TIMEOUT" "$API_URL/ideas" 2>/dev/null)

    if echo "$response" | grep -q -E '\[|\{'; then
        log_success "Ideas list endpoint working"
    else
        log_fail "Ideas list endpoint failed"
    fi
}

test_api_create_idea() {
    log_info "Testing API create idea endpoint..."

    test_idea='{"description":"Smoke test idea","tags":["test"]}'

    response=$(curl -s --max-time "$TIMEOUT" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$test_idea" \
        "$API_URL/ideas" 2>/dev/null)

    if echo "$response" | grep -q "id"; then
        idea_id=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        log_success "Create idea endpoint working (ID: $idea_id)"

        # Clean up: delete the test idea
        if [ -n "$idea_id" ]; then
            curl -s --max-time "$TIMEOUT" -X DELETE "$API_URL/ideas/$idea_id" > /dev/null 2>&1
        fi
    else
        log_fail "Create idea endpoint failed"
    fi
}

test_api_telos_config() {
    log_info "Testing API telos configuration endpoint..."

    response=$(curl -s --max-time "$TIMEOUT" "$API_URL/telos" 2>/dev/null)

    if echo "$response" | grep -q -E 'goals|strategies|stack'; then
        log_success "Telos config endpoint working"
    else
        log_fail "Telos config endpoint failed"
    fi
}

test_database_persistence() {
    log_info "Testing database persistence..."

    # Create an idea
    test_idea='{"description":"Persistence test","tags":["smoke-test"]}'
    response=$(curl -s --max-time "$TIMEOUT" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$test_idea" \
        "$API_URL/ideas" 2>/dev/null)

    if echo "$response" | grep -q "id"; then
        idea_id=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

        # Retrieve the idea
        sleep 1
        retrieve_response=$(curl -s --max-time "$TIMEOUT" "$API_URL/ideas/$idea_id" 2>/dev/null)

        if echo "$retrieve_response" | grep -q "Persistence test"; then
            log_success "Database persistence working"
        else
            log_fail "Database persistence failed"
        fi

        # Clean up
        if [ -n "$idea_id" ]; then
            curl -s --max-time "$TIMEOUT" -X DELETE "$API_URL/ideas/$idea_id" > /dev/null 2>&1
        fi
    else
        log_fail "Could not create idea for persistence test"
    fi
}

test_monitoring_endpoints() {
    log_info "Testing monitoring endpoints..."

    # Test Prometheus
    if curl -f -s --max-time "$TIMEOUT" "http://localhost:9090/-/healthy" > /dev/null 2>&1; then
        log_success "Prometheus is healthy"
    else
        log_warn "Prometheus health check failed (may not be running)"
    fi

    # Test Grafana
    if curl -f -s --max-time "$TIMEOUT" "http://localhost:3001/api/health" > /dev/null 2>&1; then
        log_success "Grafana is healthy"
    else
        log_warn "Grafana health check failed (may not be running)"
    fi
}

test_error_handling() {
    log_info "Testing API error handling..."

    # Test 404 error
    response_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time "$TIMEOUT" "$API_URL/nonexistent" 2>/dev/null)

    if [ "$response_code" == "404" ]; then
        log_success "404 error handling working"
    else
        log_fail "404 error handling failed (got $response_code)"
    fi
}

test_performance() {
    log_info "Testing API response time..."

    start_time=$(date +%s%N)
    curl -s --max-time "$TIMEOUT" "$API_URL/health" > /dev/null 2>&1
    end_time=$(date +%s%N)

    duration=$((($end_time - $start_time) / 1000000))

    if [ "$duration" -lt 1000 ]; then
        log_success "API response time acceptable (${duration}ms)"
    else
        log_warn "API response time slow (${duration}ms)"
    fi
}

# Main execution
main() {
    echo ""
    echo "========================================="
    echo "  Telos Idea Matrix - Smoke Test Suite  "
    echo "========================================="
    echo ""
    echo "API URL: $API_URL"
    echo "Web URL: $WEB_URL"
    echo ""

    # Wait for services to be ready
    log_info "Waiting for services to be ready..."
    sleep 5

    # Run all tests
    test_api_health
    test_api_version
    test_frontend_health
    test_api_ideas_list
    test_api_create_idea
    test_api_telos_config
    test_database_persistence
    test_error_handling
    test_performance
    test_monitoring_endpoints

    # Summary
    echo ""
    echo "========================================="
    echo "  Test Summary"
    echo "========================================="
    echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
    echo ""

    if [ "$FAILED_TESTS" -eq 0 ]; then
        echo -e "${GREEN}✓ All critical tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some tests failed. Please investigate.${NC}"
        exit 1
    fi
}

# Show usage if help requested
if [[ "$1" == "--help" ]] || [[ "$1" == "-h" ]]; then
    cat <<EOF
Usage: $0 [OPTIONS]

Smoke test script for Telos Idea Matrix deployment.

Options:
    --api-url URL     API base URL (default: http://localhost:8080)
    --web-url URL     Frontend URL (default: http://localhost:3000)
    --timeout SEC     Request timeout in seconds (default: 10)
    -h, --help        Show this help message

Environment Variables:
    API_URL          Override default API URL
    WEB_URL          Override default Web URL

Examples:
    $0
    $0 --api-url http://staging.example.com:8080
    API_URL=https://api.production.com $0
EOF
    exit 0
fi

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --api-url)
            API_URL="$2"
            shift 2
            ;;
        --web-url)
            WEB_URL="$2"
            shift 2
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run main
main
