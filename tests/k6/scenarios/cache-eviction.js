/*
 * Cache Eviction Scenario
 *
 * Tests Guard's cache TTL behavior and measures performance
 * degradation when cache entries expire.
 *
 * Test Phases:
 *   1. Warm-up: Populate cache with deterministic entries (2 min)
 *   2. Pre-eviction: Access cached entries, expect high hit rate (2 min)
 *   3. Wait: Let TTL expire (configurable, default 3 min for master config)
 *   4. Post-eviction: Access same entries, measure cache miss penalty (2 min)
 *
 * Note: For testing TTL=10min (improved config), adjust WAIT_DURATION accordingly.
 *
 * Usage:
 *   k6 run -e GUARD_URL=https://localhost:8443 -e WAIT_DURATION=3m scenarios/cache-eviction.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import { Trend, Counter, Rate } from 'k6/metrics';
import exec from 'k6/execution';
import { generateUserPool, generateCacheHitPayload } from '../lib/payloads.js';

// Configuration from environment
const BASE_URL = __ENV.GUARD_URL || 'https://localhost:8443';
const USER_POOL_SIZE = parseInt(__ENV.USER_POOL_SIZE || '50');
const WARMUP_ITERATIONS = parseInt(__ENV.WARMUP_ITERATIONS || '200');
// Wait duration should be TTL + buffer (e.g., for 3min TTL, use 4m; for 10min TTL, use 11m)
const WAIT_DURATION = __ENV.WAIT_DURATION || '4m';

// Pre-generate user pool (shared across VUs)
const users = new SharedArray('users', function() {
  return generateUserPool(USER_POOL_SIZE);
});

// Phase-specific metrics
const warmupLatency = new Trend('warmup_latency_ms');
const preEvictionLatency = new Trend('pre_eviction_latency_ms');
const postEvictionLatency = new Trend('post_eviction_latency_ms');
const evictionDetected = new Counter('eviction_detected');
const cacheHitsPre = new Counter('cache_hits_pre');
const cacheMissesPre = new Counter('cache_misses_pre');
const cacheHitsPost = new Counter('cache_hits_post');
const cacheMissesPost = new Counter('cache_misses_post');

// Cache hit threshold in milliseconds
const CACHE_HIT_THRESHOLD_MS = 50;

// Parse wait duration to calculate start times
function parseDuration(duration) {
  const match = duration.match(/^(\d+)(s|m|h)$/);
  if (!match) return 240; // Default 4 minutes in seconds
  const value = parseInt(match[1]);
  const unit = match[2];
  if (unit === 's') return value;
  if (unit === 'm') return value * 60;
  if (unit === 'h') return value * 3600;
  return 240;
}

const waitDurationSec = parseDuration(WAIT_DURATION);
const warmupEndTime = '2m';
const preEvictionStart = '2m';
const preEvictionEnd = '4m';
const waitStart = '4m';
const waitEnd = `${4 * 60 + waitDurationSec}s`;
const postEvictionStart = waitEnd;

export const options = {
  scenarios: {
    // Phase 1: Warm up cache with deterministic entries
    warmup: {
      executor: 'shared-iterations',
      vus: 5,
      iterations: WARMUP_ITERATIONS,
      maxDuration: '2m',
      exec: 'warmupPhase',
      gracefulStop: '10s',
    },
    // Phase 2: Sustained access (should be mostly cache hits)
    pre_eviction: {
      executor: 'constant-arrival-rate',
      rate: 10,
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 10,
      startTime: '2m',
      exec: 'preEvictionPhase',
      gracefulStop: '10s',
    },
    // Phase 3: Wait for TTL expiration
    wait_for_eviction: {
      executor: 'constant-vus',
      vus: 1,
      duration: WAIT_DURATION,
      startTime: '4m',
      exec: 'waitPhase',
      gracefulStop: '10s',
    },
    // Phase 4: Post-eviction access (should be cache misses)
    post_eviction: {
      executor: 'constant-arrival-rate',
      rate: 10,
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 10,
      startTime: postEvictionStart,
      exec: 'postEvictionPhase',
      gracefulStop: '10s',
    },
  },
  insecureSkipTLSVerify: true,
  thresholds: {
    http_req_failed: ['rate<0.05'],
  },
};

export function setup() {
  console.log(`Cache eviction test starting`);
  console.log(`Guard URL: ${BASE_URL}`);
  console.log(`Wait duration for TTL: ${WAIT_DURATION}`);
  console.log(`User pool size: ${USER_POOL_SIZE}`);

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
    waitDuration: WAIT_DURATION,
  };
}

export function warmupPhase() {
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

  warmupLatency.add(duration);

  check(response, {
    'warmup: status is 200': (r) => r.status === 200,
  });

  sleep(0.05);
}

export function preEvictionPhase() {
  const iteration = exec.scenario.iterationInTest;
  // Use limited index to ensure we hit cached entries
  const payload = generateCacheHitPayload(users, iteration % WARMUP_ITERATIONS);

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

  preEvictionLatency.add(duration);

  // Track cache hits/misses
  if (duration < CACHE_HIT_THRESHOLD_MS) {
    cacheHitsPre.add(1);
  } else {
    cacheMissesPre.add(1);
  }

  check(response, {
    'pre-eviction: status is 200': (r) => r.status === 200,
  });
}

export function waitPhase() {
  // Just wait - this phase exists to let TTL expire
  // Log progress every 30 seconds
  if (__ITER % 30 === 0) {
    console.log(`Waiting for TTL expiration... ${__ITER}s elapsed`);
  }
  sleep(1);
}

export function postEvictionPhase() {
  const iteration = exec.scenario.iterationInTest;
  // Use same limited index as pre-eviction to access same cache keys
  const payload = generateCacheHitPayload(users, iteration % WARMUP_ITERATIONS);

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

  postEvictionLatency.add(duration);

  // Track cache hits/misses
  if (duration < CACHE_HIT_THRESHOLD_MS) {
    cacheHitsPost.add(1);
  } else {
    cacheMissesPost.add(1);
    evictionDetected.add(1);
  }

  check(response, {
    'post-eviction: status is 200': (r) => r.status === 200,
  });
}

export function teardown(data) {
  const durationSec = (Date.now() - data.startTime) / 1000;
  console.log(`Cache eviction test completed in ${durationSec.toFixed(2)}s`);
}

export function handleSummary(data) {
  const metrics = data.metrics || {};

  // Calculate pre-eviction cache hit rate
  const preHits = metrics.cache_hits_pre ? metrics.cache_hits_pre.values.count : 0;
  const preMisses = metrics.cache_misses_pre ? metrics.cache_misses_pre.values.count : 0;
  const preTotal = preHits + preMisses;
  const preHitRate = preTotal > 0 ? preHits / preTotal : 0;

  // Calculate post-eviction cache hit rate
  const postHits = metrics.cache_hits_post ? metrics.cache_hits_post.values.count : 0;
  const postMisses = metrics.cache_misses_post ? metrics.cache_misses_post.values.count : 0;
  const postTotal = postHits + postMisses;
  const postHitRate = postTotal > 0 ? postHits / postTotal : 0;

  const summary = {
    scenario: 'cache-eviction',
    timestamp: new Date().toISOString(),
    config: {
      guard_url: BASE_URL,
      wait_duration: WAIT_DURATION,
      warmup_iterations: WARMUP_ITERATIONS,
      user_pool_size: USER_POOL_SIZE,
    },
    phases: {
      warmup: {
        latency_avg: metrics.warmup_latency_ms ? metrics.warmup_latency_ms.values.avg : 0,
        latency_p95: metrics.warmup_latency_ms ? metrics.warmup_latency_ms.values['p(95)'] : 0,
      },
      pre_eviction: {
        requests: preTotal,
        cache_hits: preHits,
        cache_misses: preMisses,
        hit_rate: preHitRate,
        latency_avg: metrics.pre_eviction_latency_ms ? metrics.pre_eviction_latency_ms.values.avg : 0,
        latency_p95: metrics.pre_eviction_latency_ms ? metrics.pre_eviction_latency_ms.values['p(95)'] : 0,
      },
      post_eviction: {
        requests: postTotal,
        cache_hits: postHits,
        cache_misses: postMisses,
        hit_rate: postHitRate,
        latency_avg: metrics.post_eviction_latency_ms ? metrics.post_eviction_latency_ms.values.avg : 0,
        latency_p95: metrics.post_eviction_latency_ms ? metrics.post_eviction_latency_ms.values['p(95)'] : 0,
        evictions_detected: metrics.eviction_detected ? metrics.eviction_detected.values.count : 0,
      },
    },
    analysis: {
      hit_rate_change: postHitRate - preHitRate,
      latency_increase_ms: (metrics.post_eviction_latency_ms ? metrics.post_eviction_latency_ms.values.avg : 0) -
                           (metrics.pre_eviction_latency_ms ? metrics.pre_eviction_latency_ms.values.avg : 0),
      ttl_effective: postHitRate < preHitRate * 0.5, // TTL is effective if hit rate drops significantly
    },
  };

  return {
    'cache-eviction-summary.json': JSON.stringify(summary, null, 2),
    stdout: generateTextSummary(summary),
  };
}

function generateTextSummary(summary) {
  const p = summary.phases;
  const a = summary.analysis;
  return `
================================================================================
Cache Eviction Test Summary
================================================================================
Scenario: ${summary.scenario}
Timestamp: ${summary.timestamp}
Guard URL: ${summary.config.guard_url}
Wait Duration (TTL): ${summary.config.wait_duration}

Phase: Warm-up
  Latency Avg: ${p.warmup.latency_avg.toFixed(2)}ms
  Latency P95: ${p.warmup.latency_p95.toFixed(2)}ms

Phase: Pre-Eviction (cached)
  Requests:    ${p.pre_eviction.requests}
  Cache Hits:  ${p.pre_eviction.cache_hits}
  Cache Miss:  ${p.pre_eviction.cache_misses}
  Hit Rate:    ${(p.pre_eviction.hit_rate * 100).toFixed(2)}%
  Latency Avg: ${p.pre_eviction.latency_avg.toFixed(2)}ms
  Latency P95: ${p.pre_eviction.latency_p95.toFixed(2)}ms

Phase: Post-Eviction (TTL expired)
  Requests:    ${p.post_eviction.requests}
  Cache Hits:  ${p.post_eviction.cache_hits}
  Cache Miss:  ${p.post_eviction.cache_misses}
  Hit Rate:    ${(p.post_eviction.hit_rate * 100).toFixed(2)}%
  Latency Avg: ${p.post_eviction.latency_avg.toFixed(2)}ms
  Latency P95: ${p.post_eviction.latency_p95.toFixed(2)}ms
  Evictions:   ${p.post_eviction.evictions_detected}

Analysis:
  Hit Rate Change:   ${(a.hit_rate_change * 100).toFixed(2)}%
  Latency Increase:  ${a.latency_increase_ms.toFixed(2)}ms
  TTL Effective:     ${a.ttl_effective ? 'YES' : 'NO'}
================================================================================
`;
}
