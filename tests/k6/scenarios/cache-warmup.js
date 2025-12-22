/*
 * Cache Warm-up Scenario
 *
 * Populates Guard's authorization cache with deterministic entries.
 * This scenario should be run before sustained-load tests to ensure
 * the cache is pre-populated and subsequent tests measure steady-state behavior.
 *
 * Usage:
 *   k6 run -e GUARD_URL=https://localhost:8443 scenarios/cache-warmup.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import { generateUserPool, generateCacheHitPayload } from '../lib/payloads.js';
import { authzLatency, authzAllowed, authzDenied, authzErrors, authzSuccessTotal } from '../lib/metrics.js';

// Configuration from environment
const BASE_URL = __ENV.GUARD_URL || 'https://localhost:8443';
const CACHE_ENTRIES_TARGET = parseInt(__ENV.CACHE_ENTRIES || '1000');
const USER_POOL_SIZE = parseInt(__ENV.USER_POOL_SIZE || '100');

// Pre-generate user pool (shared across VUs for consistency)
const users = new SharedArray('users', function() {
  return generateUserPool(USER_POOL_SIZE);
});

export const options = {
  scenarios: {
    cache_warmup: {
      executor: 'shared-iterations',
      vus: 10,
      iterations: CACHE_ENTRIES_TARGET,
      maxDuration: '5m',
    },
  },
  // TLS configuration will be loaded from environment or defaults
  insecureSkipTLSVerify: true,
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<5000'],
  },
};

export function setup() {
  console.log(`Cache warm-up starting: target=${CACHE_ENTRIES_TARGET} entries, users=${USER_POOL_SIZE}`);
  console.log(`Guard URL: ${BASE_URL}`);

  // Verify Guard is accessible
  const healthCheck = http.get(`${BASE_URL}/healthz`, {
    insecureSkipTLSVerify: true,
    timeout: '10s',
  });

  if (healthCheck.status !== 200) {
    console.error(`Guard health check failed: status=${healthCheck.status}`);
  } else {
    console.log('Guard health check passed');
  }

  return { startTime: Date.now() };
}

export default function() {
  const payload = generateCacheHitPayload(users, __ITER);

  const startTime = Date.now();
  const response = http.post(
    `${BASE_URL}/subjectaccessreviews`,
    JSON.stringify(payload),
    {
      headers: { 'Content-Type': 'application/json' },
      timeout: '30s',
      insecureSkipTLSVerify: true,
    }
  );
  const duration = Date.now() - startTime;

  authzLatency.add(duration);

  const success = check(response, {
    'status is 200': (r) => r.status === 200,
    'has valid response': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.status !== undefined;
      } catch {
        return false;
      }
    },
  });

  if (success) {
    authzSuccessTotal.add(1);
    try {
      const body = JSON.parse(response.body);
      if (body.status && body.status.allowed) {
        authzAllowed.add(1);
      } else {
        authzDenied.add(1);
      }
    } catch (e) {
      authzErrors.add(1);
    }
  } else {
    authzErrors.add(1);
    if (__ITER < 5) {
      console.error(`Request ${__ITER} failed: status=${response.status}, body=${response.body}`);
    }
  }

  // Small delay to avoid overwhelming during warmup
  sleep(0.05);
}

export function teardown(data) {
  const durationSec = (Date.now() - data.startTime) / 1000;
  console.log(`Cache warm-up completed in ${durationSec.toFixed(2)}s`);
  console.log(`Target entries: ${CACHE_ENTRIES_TARGET}`);
}

export function handleSummary(data) {
  const summary = {
    scenario: 'cache-warmup',
    timestamp: new Date().toISOString(),
    config: {
      guard_url: BASE_URL,
      cache_entries_target: CACHE_ENTRIES_TARGET,
      user_pool_size: USER_POOL_SIZE,
    },
    metrics: {
      iterations: data.metrics.iterations ? data.metrics.iterations.values.count : 0,
      http_req_duration_avg: data.metrics.http_req_duration ? data.metrics.http_req_duration.values.avg : 0,
      http_req_duration_p95: data.metrics.http_req_duration ? data.metrics.http_req_duration.values['p(95)'] : 0,
      http_req_failed_rate: data.metrics.http_req_failed ? data.metrics.http_req_failed.values.rate : 0,
      authz_allowed: data.metrics.authz_allowed_total ? data.metrics.authz_allowed_total.values.count : 0,
      authz_denied: data.metrics.authz_denied_total ? data.metrics.authz_denied_total.values.count : 0,
      authz_errors: data.metrics.authz_errors_total ? data.metrics.authz_errors_total.values.count : 0,
    },
  };

  return {
    'warmup-summary.json': JSON.stringify(summary, null, 2),
    stdout: generateTextSummary(summary),
  };
}

function generateTextSummary(summary) {
  return `
================================================================================
Cache Warm-up Summary
================================================================================
Scenario: ${summary.scenario}
Timestamp: ${summary.timestamp}
Guard URL: ${summary.config.guard_url}

Target Entries: ${summary.config.cache_entries_target}
Iterations:     ${summary.metrics.iterations}

Latency:
  Average: ${summary.metrics.http_req_duration_avg.toFixed(2)}ms
  P95:     ${summary.metrics.http_req_duration_p95.toFixed(2)}ms

Results:
  Allowed: ${summary.metrics.authz_allowed}
  Denied:  ${summary.metrics.authz_denied}
  Errors:  ${summary.metrics.authz_errors}
  Failed:  ${(summary.metrics.http_req_failed_rate * 100).toFixed(2)}%
================================================================================
`;
}
