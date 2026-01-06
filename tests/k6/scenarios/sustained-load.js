/*
 * Sustained Load Scenario
 *
 * Measures steady-state performance with a warm cache.
 * Sends requests at a constant rate with configurable cache hit ratio.
 *
 * Key Metrics:
 *   - Cache hit rate (inferred from latency)
 *   - P50/P95/P99 latencies
 *   - Throughput (RPS)
 *   - Error rate
 *
 * Usage:
 *   k6 run -e GUARD_URL=https://localhost:8443 -e CACHE_HIT_RATIO=0.8 scenarios/sustained-load.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import exec from 'k6/execution';
import { generateUserPool, generateMixedPayload } from '../lib/payloads.js';
import {
  cacheHitRate,
  cacheHitsInferred,
  cacheMissesInferred,
  authzLatency,
  authzAllowed,
  authzDenied,
  authzErrors,
  authzSuccessTotal,
  cacheHitLatency,
  cacheMissLatency,
  throttledRequests,
} from '../lib/metrics.js';

// Configuration from environment
const BASE_URL = __ENV.GUARD_URL || 'https://localhost:8443';
const CACHE_HIT_RATIO = parseFloat(__ENV.CACHE_HIT_RATIO || '0.8');
const TARGET_RPS = parseInt(__ENV.TARGET_RPS || '100');
const DURATION = __ENV.DURATION || '5m';
const USER_POOL_SIZE = parseInt(__ENV.USER_POOL_SIZE || '100');

// Cache hit threshold in milliseconds
const CACHE_HIT_THRESHOLD_MS = 50;

// Pre-generate user pool (shared across VUs)
const users = new SharedArray('users', function() {
  return generateUserPool(USER_POOL_SIZE);
});

export const options = {
  scenarios: {
    sustained_load: {
      executor: 'constant-arrival-rate',
      rate: TARGET_RPS,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: Math.ceil(TARGET_RPS / 2),
      maxVUs: TARGET_RPS * 2,
    },
  },
  insecureSkipTLSVerify: true,
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(50)<100', 'p(95)<500', 'p(99)<1000'],
    cache_hit_rate: [`rate>${CACHE_HIT_RATIO * 0.8}`], // Allow 20% degradation from target
  },
};

export function setup() {
  console.log(`Sustained load test starting`);
  console.log(`Guard URL: ${BASE_URL}`);
  console.log(`Target RPS: ${TARGET_RPS}`);
  console.log(`Duration: ${DURATION}`);
  console.log(`Cache hit ratio target: ${(CACHE_HIT_RATIO * 100).toFixed(0)}%`);

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
    config: {
      targetRps: TARGET_RPS,
      duration: DURATION,
      cacheHitRatio: CACHE_HIT_RATIO,
    },
  };
}

export default function() {
  const vuId = exec.vu.idInTest;
  const iteration = exec.scenario.iterationInTest;

  // Generate payload with configured cache hit ratio
  const payload = generateMixedPayload(users, vuId, iteration, CACHE_HIT_RATIO);

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

  // Record latency
  authzLatency.add(duration);

  // Infer cache hit from latency
  const likelyCacheHit = duration < CACHE_HIT_THRESHOLD_MS;
  cacheHitRate.add(likelyCacheHit ? 1 : 0);

  if (likelyCacheHit) {
    cacheHitsInferred.add(1);
    cacheHitLatency.add(duration);
  } else {
    cacheMissesInferred.add(1);
    cacheMissLatency.add(duration);
  }

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
  console.log(`Sustained load test completed in ${durationSec.toFixed(2)}s`);
}

export function handleSummary(data) {
  const metrics = data.metrics || {};

  const hits = metrics.cache_hits_inferred ? metrics.cache_hits_inferred.values.count : 0;
  const misses = metrics.cache_misses_inferred ? metrics.cache_misses_inferred.values.count : 0;
  const totalRequests = hits + misses;
  const actualCacheHitRate = totalRequests > 0 ? hits / totalRequests : 0;

  const summary = {
    scenario: 'sustained-load',
    timestamp: new Date().toISOString(),
    config: {
      guard_url: BASE_URL,
      target_rps: TARGET_RPS,
      duration: DURATION,
      cache_hit_ratio_target: CACHE_HIT_RATIO,
      user_pool_size: USER_POOL_SIZE,
    },
    metrics: {
      total_requests: totalRequests,
      actual_rps: metrics.http_reqs ? metrics.http_reqs.values.rate : 0,
      cache: {
        hits_inferred: hits,
        misses_inferred: misses,
        hit_rate: actualCacheHitRate,
        hit_latency_avg: metrics.cache_hit_latency_ms ? metrics.cache_hit_latency_ms.values.avg : 0,
        miss_latency_avg: metrics.cache_miss_latency_ms ? metrics.cache_miss_latency_ms.values.avg : 0,
      },
      latency: {
        avg: metrics.http_req_duration ? metrics.http_req_duration.values.avg : 0,
        p50: metrics.http_req_duration ? metrics.http_req_duration.values['p(50)'] : 0,
        p95: metrics.http_req_duration ? metrics.http_req_duration.values['p(95)'] : 0,
        p99: metrics.http_req_duration ? metrics.http_req_duration.values['p(99)'] : 0,
        min: metrics.http_req_duration ? metrics.http_req_duration.values.min : 0,
        max: metrics.http_req_duration ? metrics.http_req_duration.values.max : 0,
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
    'sustained-load-summary.json': JSON.stringify(summary, null, 2),
    stdout: generateTextSummary(summary),
  };
}

function generateTextSummary(summary) {
  const m = summary.metrics;
  return `
================================================================================
Sustained Load Test Summary
================================================================================
Scenario: ${summary.scenario}
Timestamp: ${summary.timestamp}
Guard URL: ${summary.config.guard_url}

Configuration:
  Target RPS:       ${summary.config.target_rps}
  Duration:         ${summary.config.duration}
  Cache Hit Target: ${(summary.config.cache_hit_ratio_target * 100).toFixed(0)}%

Results:
  Total Requests:   ${m.total_requests}
  Actual RPS:       ${m.actual_rps.toFixed(2)}

Cache Performance:
  Hits (inferred):  ${m.cache.hits_inferred}
  Misses (inferred): ${m.cache.misses_inferred}
  Hit Rate:         ${(m.cache.hit_rate * 100).toFixed(2)}%
  Hit Latency Avg:  ${m.cache.hit_latency_avg.toFixed(2)}ms
  Miss Latency Avg: ${m.cache.miss_latency_avg.toFixed(2)}ms

Overall Latency:
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
