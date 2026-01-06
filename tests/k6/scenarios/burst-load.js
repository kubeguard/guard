/*
 * Burst Load Scenario
 *
 * Tests Guard's behavior under traffic spikes.
 * Simulates realistic traffic patterns with sudden increases and decreases.
 *
 * Traffic Pattern:
 *   1. Ramp up from 10 to 50 RPS (30s)
 *   2. Spike to 300 RPS (1m)
 *   3. Drop to 50 RPS (30s)
 *   4. Sustain at 100 RPS (1m)
 *   5. Major spike to 500 RPS (30s)
 *   6. Recovery to 50 RPS (30s)
 *
 * Usage:
 *   k6 run -e GUARD_URL=https://localhost:8443 scenarios/burst-load.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import exec from 'k6/execution';
import { generateUserPool, generateCacheHitPayload } from '../lib/payloads.js';
import {
  authzLatency,
  authzAllowed,
  authzDenied,
  authzErrors,
  authzSuccessTotal,
  throttledRequests,
  cacheHitRate,
} from '../lib/metrics.js';

// Configuration from environment
const BASE_URL = __ENV.GUARD_URL || 'https://localhost:8443';
const USER_POOL_SIZE = parseInt(__ENV.USER_POOL_SIZE || '100');
const MAX_VUS = parseInt(__ENV.MAX_VUS || '500');

// Cache hit threshold in milliseconds
const CACHE_HIT_THRESHOLD_MS = 50;

// Pre-generate user pool (shared across VUs)
const users = new SharedArray('users', function() {
  return generateUserPool(USER_POOL_SIZE);
});

export const options = {
  scenarios: {
    burst_traffic: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: MAX_VUS,
      stages: [
        { duration: '30s', target: 50 },    // Warm up
        { duration: '1m', target: 300 },    // First spike
        { duration: '30s', target: 50 },    // Recovery
        { duration: '1m', target: 100 },    // Sustained
        { duration: '30s', target: 500 },   // Major spike
        { duration: '30s', target: 50 },    // Final recovery
      ],
    },
  },
  insecureSkipTLSVerify: true,
  thresholds: {
    http_req_failed: ['rate<0.05'],  // Allow 5% errors during bursts
    http_req_duration: ['p(95)<2000'],
  },
};

export function setup() {
  console.log(`Burst load test starting`);
  console.log(`Guard URL: ${BASE_URL}`);
  console.log(`Max VUs: ${MAX_VUS}`);
  console.log(`Traffic pattern: 10→50→300→50→100→500→50 RPS`);

  // Verify Guard is accessible
  const healthCheck = http.get(`${BASE_URL}/healthz`, {
    insecureSkipTLSVerify: true,
    timeout: '10s',
  });

  if (healthCheck.status !== 200) {
    console.error(`Guard health check failed: status=${healthCheck.status}`);
  }

  return {
    startTime: Date.now(),
  };
}

export default function() {
  const iteration = exec.scenario.iterationInTest;

  // Use cache-hit optimized payloads to focus on throughput testing
  const payload = generateCacheHitPayload(users, iteration);

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

  // Record metrics
  authzLatency.add(duration);

  // Infer cache hit from latency
  const likelyCacheHit = duration < CACHE_HIT_THRESHOLD_MS;
  cacheHitRate.add(likelyCacheHit ? 1 : 0);

  // Check response
  const success = check(response, {
    'status is 200': (r) => r.status === 200,
  });

  if (response.status === 429) {
    throttledRequests.add(1);
  }

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
  }
}

export function teardown(data) {
  const durationSec = (Date.now() - data.startTime) / 1000;
  console.log(`Burst load test completed in ${durationSec.toFixed(2)}s`);
}

export function handleSummary(data) {
  const metrics = data.metrics || {};

  const summary = {
    scenario: 'burst-load',
    timestamp: new Date().toISOString(),
    config: {
      guard_url: BASE_URL,
      max_vus: MAX_VUS,
      traffic_pattern: '10→50→300→50→100→500→50 RPS',
      user_pool_size: USER_POOL_SIZE,
    },
    metrics: {
      total_requests: metrics.http_reqs ? metrics.http_reqs.values.count : 0,
      peak_rps: metrics.http_reqs ? metrics.http_reqs.values.rate : 0,
      latency: {
        avg: metrics.http_req_duration ? metrics.http_req_duration.values.avg : 0,
        p50: metrics.http_req_duration ? metrics.http_req_duration.values['p(50)'] : 0,
        p95: metrics.http_req_duration ? metrics.http_req_duration.values['p(95)'] : 0,
        p99: metrics.http_req_duration ? metrics.http_req_duration.values['p(99)'] : 0,
        min: metrics.http_req_duration ? metrics.http_req_duration.values.min : 0,
        max: metrics.http_req_duration ? metrics.http_req_duration.values.max : 0,
      },
      cache: {
        hit_rate: metrics.cache_hit_rate ? metrics.cache_hit_rate.values.rate : 0,
      },
      authz: {
        allowed: metrics.authz_allowed_total ? metrics.authz_allowed_total.values.count : 0,
        denied: metrics.authz_denied_total ? metrics.authz_denied_total.values.count : 0,
        errors: metrics.authz_errors_total ? metrics.authz_errors_total.values.count : 0,
      },
      errors: {
        failed_rate: metrics.http_req_failed ? metrics.http_req_failed.values.rate : 0,
        throttled: metrics.throttled_requests_total ? metrics.throttled_requests_total.values.count : 0,
      },
    },
  };

  return {
    'burst-load-summary.json': JSON.stringify(summary, null, 2),
    stdout: generateTextSummary(summary),
  };
}

function generateTextSummary(summary) {
  const m = summary.metrics;
  return `
================================================================================
Burst Load Test Summary
================================================================================
Scenario: ${summary.scenario}
Timestamp: ${summary.timestamp}
Guard URL: ${summary.config.guard_url}

Configuration:
  Max VUs:          ${summary.config.max_vus}
  Traffic Pattern:  ${summary.config.traffic_pattern}

Results:
  Total Requests:   ${m.total_requests}
  Peak RPS:         ${m.peak_rps.toFixed(2)}

Cache Performance:
  Hit Rate:         ${(m.cache.hit_rate * 100).toFixed(2)}%

Latency:
  Average: ${m.latency.avg.toFixed(2)}ms
  P50:     ${m.latency.p50.toFixed(2)}ms
  P95:     ${m.latency.p95.toFixed(2)}ms
  P99:     ${m.latency.p99.toFixed(2)}ms
  Min:     ${m.latency.min.toFixed(2)}ms
  Max:     ${m.latency.max.toFixed(2)}ms

Authorization:
  Allowed: ${m.authz.allowed}
  Denied:  ${m.authz.denied}
  Errors:  ${m.authz.errors}

Errors:
  Failed Rate: ${(m.errors.failed_rate * 100).toFixed(2)}%
  Throttled:   ${m.errors.throttled}
================================================================================
`;
}
