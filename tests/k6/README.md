# Guard Cache Performance Testing Framework

This directory contains k6 load testing scripts for evaluating Guard's Azure authorization cache improvements.

## Prerequisites

1. **k6** - Load testing tool
   ```bash
   brew install k6
   ```

2. **Go** - For building the mock Azure server
   ```bash
   # Already installed if you're developing Guard
   ```

3. **Guard binary** - Built from this repository
   ```bash
   make build
   ```

## Quick Start

Run the full comparison test suite:

```bash
./run-comparison.sh
```

This will:
1. Build the Guard binary
2. Generate TLS certificates
3. Start the mock Azure server
4. Run tests for each cache configuration
5. Generate a comparison report

## Generating TLS Certificates

The test framework requires mTLS certificates. Generate them using Guard's built-in commands:

```bash
# Create certs directory
mkdir -p tests/k6/certs
cd tests/k6/certs

# Generate CA certificate
guard init ca

# Generate server certificate for Guard
guard init server --domains=localhost --ips=127.0.0.1

# Generate client certificate for k6 (must use "azure" organization for authz)
guard init client azure -o azure

# Verify certificates
ls -la
# Expected files: ca.crt, ca.key, server.crt, server.key, client.crt, client.key
```

**Note**: Private keys (`*.key`) are excluded from git via `.gitignore`. You must generate them locally before running tests.

## Directory Structure

```
tests/k6/
├── certs/                  # TLS certificates (generated)
├── lib/
│   ├── config.js           # k6 configuration
│   ├── payloads.js         # SubjectAccessReview generators
│   └── metrics.js          # Custom k6 metrics
├── scenarios/
│   ├── cache-warmup.js     # Populate cache
│   ├── sustained-load.js   # Steady 100 RPS
│   ├── burst-load.js       # Traffic spikes
│   └── cache-eviction.js   # TTL testing
├── results/                # Test results (generated)
├── run-comparison.sh       # Main orchestration script
├── compare-results.py      # Results analysis
└── README.md               # This file

tests/mock-server/
└── main.go                 # Mock Azure RBAC server
```

## Test Scenarios

### 1. Cache Warm-up
Populates the cache with deterministic entries before main tests.

```bash
k6 run -e GUARD_URL=https://localhost:8443 scenarios/cache-warmup.js
```

### 2. Sustained Load
Measures steady-state performance at 100 RPS with 80% cache hit ratio.

```bash
k6 run -e GUARD_URL=https://localhost:8443 -e CACHE_HIT_RATIO=0.8 scenarios/sustained-load.js
```

### 3. Burst Load
Tests behavior under traffic spikes (10→300→500 RPS).

```bash
k6 run -e GUARD_URL=https://localhost:8443 scenarios/burst-load.js
```

### 4. Cache Eviction
Tests TTL behavior by measuring performance before and after cache expiration.

```bash
k6 run -e GUARD_URL=https://localhost:8443 -e WAIT_DURATION=4m scenarios/cache-eviction.js
```

## Cache Configurations Tested

| Config | Cache Size | TTL | Description |
|--------|------------|-----|-------------|
| master | 5 MB | 3 min | Original configuration |
| improved_default | 50 MB | 10 min | New defaults |
| improved_large | 100 MB | 10 min | Large cache variant |
| improved_long_ttl | 50 MB | 30 min | Long TTL variant |

## Mock Server

The mock Azure server simulates realistic Azure RBAC CheckAccess API behavior:

- **Latency**: 50-200ms (configurable)
- **Allow Rate**: 90% (configurable)
- **Throttling**: 1% chance of HTTP 429 (configurable)

Start manually for debugging:

```bash
cd tests/mock-server
go build -o mock-server main.go
./mock-server -port 8080 -min-latency 50 -max-latency 200 -allow-rate 0.9 -throttle-rate 0.01 -verbose
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| GUARD_URL | https://localhost:8443 | Guard server URL |
| MOCK_SERVER_URL | http://localhost:8080 | Mock server URL |
| CACHE_HIT_RATIO | 0.8 | Target cache hit ratio |
| TARGET_RPS | 100 | Requests per second |
| DURATION | 5m | Test duration |
| USER_POOL_SIZE | 100 | Number of test users |

## Results Analysis

After running tests, analyze results:

```bash
python3 compare-results.py ./results
```

Sample output:
```
================================================================================
Scenario: sustained-load
================================================================================
Config               P50(ms)    P95(ms)    P99(ms)    Avg(ms)        RPS  Cache Hit%
-------------------------------------------------------------------------------------
master                 150.00     480.00     950.00     180.00     100.00       65.0%
improved_default        45.00     180.00     350.00      60.00     150.00       82.5%

Improvements vs master:
  improved_default: p95 latency -62.5%, cache hit rate +26.9%
```

## Key Metrics

### From k6
- `http_req_duration` - Request latency percentiles
- `http_reqs` - Request rate (RPS)
- `http_req_failed` - Error rate

### From Prometheus (/metrics)
- `guard_azure_authz_cache_hits_total` - Cache hits
- `guard_azure_authz_cache_misses_total` - Cache misses
- `guard_azure_authz_cache_entries` - Current cache size

### Custom Metrics
- `cache_hit_rate` - Inferred from latency
- `authz_latency_ms` - Authorization latency
- `throttled_requests_total` - HTTP 429 responses

## Troubleshooting

### Guard fails to start
- Check TLS certificates exist in `certs/`
- Verify mock server is running: `curl http://localhost:8080/health`
- Check Guard logs for errors

### Low cache hit rate
- Ensure cache-warmup ran before other tests
- Verify USER_POOL_SIZE matches between warmup and tests
- Check TTL hasn't expired (especially for master config with 3min TTL)

### k6 connection errors
- Verify Guard is running: `curl -k https://localhost:8443/healthz`
- Check TLS certificates are correct
- Ensure client cert organization is "azure"
