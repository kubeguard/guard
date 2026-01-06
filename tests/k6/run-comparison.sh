#!/bin/bash
#
# Guard Cache Performance Comparison Script
#
# This script runs k6 load tests against Guard with different cache configurations
# to compare master branch vs improved branch performance.
#
# Usage:
#   ./run-comparison.sh [options]
#
# Options:
#   --skip-build     Skip building Guard binary
#   --skip-certs     Skip certificate generation
#   --config NAME    Run only specific config (master, improved_default, etc.)
#   --scenario NAME  Run only specific scenario (cache-warmup, sustained-load, etc.)
#   --mock-only      Only start mock server (for debugging)
#   -h, --help       Show this help message
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
GUARD_BINARY="${GUARD_BINARY:-$PROJECT_ROOT/bin/guard-darwin-arm64}"
GUARD_DATA_DIR="${GUARD_DATA_DIR:-/tmp/guard-test}"
RESULTS_DIR="${RESULTS_DIR:-$SCRIPT_DIR/results}"
K6_SCENARIOS_DIR="$SCRIPT_DIR/scenarios"

MOCK_SERVER_PORT=8080
GUARD_PORT=8443

# Cache configurations to test
declare -A CACHE_CONFIGS
CACHE_CONFIGS["master"]="5:3"       # 5MB, 3min TTL (original)
CACHE_CONFIGS["improved_default"]="50:10"   # 50MB, 10min TTL
CACHE_CONFIGS["improved_large"]="100:10"    # 100MB, 10min TTL
CACHE_CONFIGS["improved_long_ttl"]="50:30"  # 50MB, 30min TTL

# Process options
SKIP_BUILD=false
SKIP_CERTS=false
CONFIG_FILTER=""
SCENARIO_FILTER=""
MOCK_ONLY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --skip-certs)
            SKIP_CERTS=true
            shift
            ;;
        --config)
            CONFIG_FILTER="$2"
            shift 2
            ;;
        --scenario)
            SCENARIO_FILTER="$2"
            shift 2
            ;;
        --mock-only)
            MOCK_ONLY=true
            shift
            ;;
        -h|--help)
            head -25 "$0" | tail -20
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Utility functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check for k6
    if ! command -v k6 &> /dev/null; then
        log_error "k6 is not installed. Install it with: brew install k6"
        exit 1
    fi

    # Check for Go
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi

    # Check for Guard binary
    if [[ ! -f "$GUARD_BINARY" ]] && [[ "$SKIP_BUILD" != "true" ]]; then
        log_warn "Guard binary not found at $GUARD_BINARY, will build"
    fi

    log_success "Prerequisites check passed"
}

# Build Guard binary
build_guard() {
    if [[ "$SKIP_BUILD" == "true" ]]; then
        log_info "Skipping Guard build"
        return
    fi

    log_info "Building Guard binary..."
    cd "$PROJECT_ROOT"

    # Detect OS and architecture
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    if [[ "$ARCH" == "x86_64" ]]; then
        ARCH="amd64"
    elif [[ "$ARCH" == "arm64" ]]; then
        ARCH="arm64"
    fi

    GUARD_BINARY="$PROJECT_ROOT/bin/guard-${OS}-${ARCH}"

    make build-${OS}_${ARCH} || {
        log_error "Failed to build Guard"
        exit 1
    }

    if [[ ! -f "$GUARD_BINARY" ]]; then
        log_error "Guard binary not found after build: $GUARD_BINARY"
        exit 1
    fi

    log_success "Guard binary built: $GUARD_BINARY"
}

# Generate certificates
setup_certificates() {
    if [[ "$SKIP_CERTS" == "true" ]]; then
        log_info "Skipping certificate generation"
        return
    fi

    log_info "Generating test certificates..."
    mkdir -p "$GUARD_DATA_DIR/pki"

    # Check if certs already exist
    if [[ -f "$GUARD_DATA_DIR/pki/ca.crt" ]] && \
       [[ -f "$GUARD_DATA_DIR/pki/server.crt" ]] && \
       [[ -f "$GUARD_DATA_DIR/pki/azure@azure.crt" ]]; then
        log_info "Certificates already exist, skipping generation"
    else
        cd "$PROJECT_ROOT"
        export GUARD_DATA_DIR="$GUARD_DATA_DIR"

        "$GUARD_BINARY" init ca --pki-dir="$GUARD_DATA_DIR/pki" 2>/dev/null || true
        "$GUARD_BINARY" init server --pki-dir="$GUARD_DATA_DIR/pki" --hosts=localhost,127.0.0.1 2>/dev/null || true
        "$GUARD_BINARY" init client -o azure --pki-dir="$GUARD_DATA_DIR/pki" 2>/dev/null || true
    fi

    # Copy certs to k6 directory
    mkdir -p "$SCRIPT_DIR/certs"
    cp "$GUARD_DATA_DIR/pki/ca.crt" "$SCRIPT_DIR/certs/" 2>/dev/null || true
    cp "$GUARD_DATA_DIR/pki/azure@azure.crt" "$SCRIPT_DIR/certs/client.crt" 2>/dev/null || true
    cp "$GUARD_DATA_DIR/pki/azure@azure.key" "$SCRIPT_DIR/certs/client.key" 2>/dev/null || true

    log_success "Certificates ready"
}

# Start mock Azure server
start_mock_server() {
    log_info "Starting mock Azure server on port $MOCK_SERVER_PORT..."

    # Build mock server if needed
    cd "$PROJECT_ROOT/tests/mock-server"
    go build -o mock-server main.go

    # Start mock server in background
    ./mock-server \
        -port "$MOCK_SERVER_PORT" \
        -min-latency 50 \
        -max-latency 200 \
        -allow-rate 0.9 \
        -throttle-rate 0.01 \
        -verbose &
    MOCK_PID=$!
    echo "$MOCK_PID" > /tmp/guard-mock-server.pid

    sleep 2

    # Verify mock server is running
    if curl -s "http://localhost:$MOCK_SERVER_PORT/health" > /dev/null; then
        log_success "Mock server started (PID: $MOCK_PID)"
    else
        log_error "Failed to start mock server"
        exit 1
    fi
}

# Stop mock Azure server
stop_mock_server() {
    if [[ -f /tmp/guard-mock-server.pid ]]; then
        MOCK_PID=$(cat /tmp/guard-mock-server.pid)
        if kill -0 "$MOCK_PID" 2>/dev/null; then
            log_info "Stopping mock server (PID: $MOCK_PID)"
            kill "$MOCK_PID" 2>/dev/null || true
        fi
        rm -f /tmp/guard-mock-server.pid
    fi
}

# Start Guard server with specific configuration
start_guard() {
    local config_name=$1
    local cache_config=${CACHE_CONFIGS[$config_name]}
    local cache_size_mb=$(echo "$cache_config" | cut -d: -f1)
    local cache_ttl_min=$(echo "$cache_config" | cut -d: -f2)

    log_info "Starting Guard with config: $config_name (size=${cache_size_mb}MB, ttl=${cache_ttl_min}min)"

    # Azure resource ID (dummy for testing)
    local resource_id="/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ContainerService/managedClusters/test-cluster"

    "$GUARD_BINARY" run \
        --ca-cert-file="$GUARD_DATA_DIR/pki/ca.crt" \
        --cert-file="$GUARD_DATA_DIR/pki/server.crt" \
        --key-file="$GUARD_DATA_DIR/pki/server.key" \
        --secure-addr=":$GUARD_PORT" \
        --authz-provider=azure \
        --azure.authz-mode=aks \
        --azure.resource-id="$resource_id" \
        --azure.aks-authz-token-url="http://localhost:$MOCK_SERVER_PORT/authz/token" \
        --azure.cache-size-mb="$cache_size_mb" \
        --azure.cache-ttl-minutes="$cache_ttl_min" \
        --v=2 &
    GUARD_PID=$!
    echo "$GUARD_PID" > /tmp/guard-server.pid

    sleep 3

    # Verify Guard is running
    if curl -sk "https://localhost:$GUARD_PORT/healthz" > /dev/null 2>&1; then
        log_success "Guard started (PID: $GUARD_PID)"
    else
        log_error "Failed to start Guard"
        kill "$GUARD_PID" 2>/dev/null || true
        exit 1
    fi
}

# Stop Guard server
stop_guard() {
    if [[ -f /tmp/guard-server.pid ]]; then
        GUARD_PID=$(cat /tmp/guard-server.pid)
        if kill -0 "$GUARD_PID" 2>/dev/null; then
            log_info "Stopping Guard (PID: $GUARD_PID)"
            kill "$GUARD_PID" 2>/dev/null || true
            wait "$GUARD_PID" 2>/dev/null || true
        fi
        rm -f /tmp/guard-server.pid
    fi
}

# Scrape Prometheus metrics
scrape_metrics() {
    local output_file=$1
    curl -sk "https://localhost:$GUARD_PORT/metrics" > "$output_file" 2>/dev/null || echo "# Failed to scrape metrics" > "$output_file"
}

# Run k6 test scenario
run_k6_test() {
    local scenario=$1
    local config_name=$2
    local output_dir="$RESULTS_DIR/$config_name"

    log_info "Running $scenario for $config_name..."

    mkdir -p "$output_dir"

    # Scrape metrics before test
    scrape_metrics "$output_dir/${scenario}_metrics_before.txt"

    # Run k6 test
    cd "$SCRIPT_DIR"
    k6 run \
        --out json="$output_dir/${scenario}_results.json" \
        --summary-export="$output_dir/${scenario}_summary.json" \
        -e GUARD_URL="https://localhost:$GUARD_PORT" \
        -e MOCK_SERVER_URL="http://localhost:$MOCK_SERVER_PORT" \
        "$K6_SCENARIOS_DIR/${scenario}.js" 2>&1 | tee "$output_dir/${scenario}_output.log"

    # Scrape metrics after test
    scrape_metrics "$output_dir/${scenario}_metrics_after.txt"

    log_success "Completed $scenario for $config_name"
}

# Run all scenarios for a configuration
run_all_scenarios() {
    local config_name=$1

    if [[ -n "$SCENARIO_FILTER" ]]; then
        run_k6_test "$SCENARIO_FILTER" "$config_name"
    else
        run_k6_test "cache-warmup" "$config_name"
        run_k6_test "sustained-load" "$config_name"
        run_k6_test "burst-load" "$config_name"
        # Skip cache-eviction by default as it takes longer
        # run_k6_test "cache-eviction" "$config_name"
    fi
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."
    stop_guard
    stop_mock_server
}

# Main execution
main() {
    echo "=========================================="
    echo "Guard Cache Performance Comparison"
    echo "=========================================="

    trap cleanup EXIT

    check_prerequisites
    build_guard
    setup_certificates

    # Create results directory
    mkdir -p "$RESULTS_DIR"

    # Start mock server
    start_mock_server

    if [[ "$MOCK_ONLY" == "true" ]]; then
        log_info "Mock server running. Press Ctrl+C to stop."
        wait
        exit 0
    fi

    # Run tests for each configuration
    for config_name in "${!CACHE_CONFIGS[@]}"; do
        # Apply config filter if specified
        if [[ -n "$CONFIG_FILTER" ]] && [[ "$CONFIG_FILTER" != "$config_name" ]]; then
            continue
        fi

        echo ""
        echo "============================================"
        echo "Testing configuration: $config_name"
        echo "============================================"

        start_guard "$config_name"
        run_all_scenarios "$config_name"
        stop_guard

        sleep 2
    done

    echo ""
    echo "============================================"
    echo "All tests completed!"
    echo "Results saved to: $RESULTS_DIR"
    echo "============================================"

    # Generate comparison report
    if command -v python3 &> /dev/null && [[ -f "$SCRIPT_DIR/compare-results.py" ]]; then
        log_info "Generating comparison report..."
        python3 "$SCRIPT_DIR/compare-results.py" "$RESULTS_DIR"
    fi
}

main "$@"
