/*
 * Simple Cache Performance Test
 *
 * This test properly configures mTLS for Guard and measures cache performance.
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';

// Configuration
const BASE_URL = __ENV.GUARD_URL || 'https://localhost:8443';
const TARGET_RPS = parseInt(__ENV.TARGET_RPS || '50');
const DURATION = __ENV.DURATION || '2m';
const CACHE_HIT_RATIO = parseFloat(__ENV.CACHE_HIT_RATIO || '0.8');

// Cache hit threshold in milliseconds
const CACHE_HIT_THRESHOLD_MS = 50;

// Custom metrics
const cacheHitRate = new Rate('cache_hit_rate');
const authzLatency = new Trend('authz_latency_ms');
const authzAllowed = new Counter('authz_allowed');
const authzDenied = new Counter('authz_denied');
const authzErrors = new Counter('authz_errors');

// TLS configuration - read certs
const clientCert = open('../certs/client.crt');
const clientKey = open('../certs/client.key');

export const options = {
  scenarios: {
    load_test: {
      executor: 'constant-arrival-rate',
      rate: TARGET_RPS,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: Math.ceil(TARGET_RPS / 2),
      maxVUs: TARGET_RPS * 2,
    },
  },
  tlsAuth: [
    {
      cert: clientCert,
      key: clientKey,
      domains: ['localhost'],
    },
  ],
  insecureSkipTLSVerify: true,
  thresholds: {
    http_req_failed: ['rate<0.1'],
    http_req_duration: ['p(95)<1000'],
  },
};

// User pool for cache key generation
const USER_POOL_SIZE = 50;
const NAMESPACES = ['default', 'kube-system', 'dev', 'staging', 'production'];
const RESOURCES = ['pods', 'deployments', 'services', 'configmaps', 'secrets'];
const VERBS = ['get', 'list', 'watch', 'create', 'update', 'delete'];

function generateUUID(seed) {
  const hex = seed.toString(16).padStart(8, '0');
  return `${hex.slice(0, 8)}-0000-4000-8000-${hex.padStart(12, '0').slice(0, 12)}`;
}

function generatePayload(iteration, useCacheHit) {
  let userIdx, nsIdx, resIdx, verbIdx;

  if (useCacheHit) {
    // Deterministic for cache hits
    userIdx = iteration % USER_POOL_SIZE;
    nsIdx = iteration % NAMESPACES.length;
    resIdx = iteration % RESOURCES.length;
    verbIdx = iteration % VERBS.length;
  } else {
    // Random for cache misses
    userIdx = Math.floor(Math.random() * 1000) + USER_POOL_SIZE;
    nsIdx = Math.floor(Math.random() * NAMESPACES.length);
    resIdx = Math.floor(Math.random() * RESOURCES.length);
    verbIdx = Math.floor(Math.random() * VERBS.length);
  }

  return {
    apiVersion: 'authorization.k8s.io/v1',
    kind: 'SubjectAccessReview',
    spec: {
      user: `testuser${userIdx}@contoso.com`,
      groups: ['system:authenticated'],
      resourceAttributes: {
        namespace: NAMESPACES[nsIdx],
        verb: VERBS[verbIdx],
        resource: RESOURCES[resIdx],
        version: 'v1',
      },
      extra: {
        oid: [generateUUID(userIdx + 1000)],
      },
    },
  };
}

export function setup() {
  console.log(`Load test starting`);
  console.log(`Guard URL: ${BASE_URL}`);
  console.log(`Target RPS: ${TARGET_RPS}`);
  console.log(`Duration: ${DURATION}`);
  console.log(`Cache hit ratio: ${CACHE_HIT_RATIO}`);

  return { startTime: Date.now() };
}

export default function() {
  // Decide if this should be a cache hit or miss
  const useCacheHit = Math.random() < CACHE_HIT_RATIO;
  const payload = generatePayload(__ITER, useCacheHit);

  const startTime = Date.now();
  const response = http.post(
    `${BASE_URL}/subjectaccessreviews`,
    JSON.stringify(payload),
    {
      headers: { 'Content-Type': 'application/json' },
      timeout: '30s',
    }
  );
  const duration = Date.now() - startTime;

  authzLatency.add(duration);

  // Infer cache hit from latency
  const likelyCacheHit = duration < CACHE_HIT_THRESHOLD_MS;
  cacheHitRate.add(likelyCacheHit ? 1 : 0);

  const success = check(response, {
    'status is 200': (r) => r.status === 200,
  });

  if (success) {
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
  console.log(`Load test completed in ${durationSec.toFixed(2)}s`);
}

export function handleSummary(data) {
  const metrics = data.metrics || {};

  const hits = metrics.cache_hit_rate ? (metrics.cache_hit_rate.values.passes || 0) : 0;
  const total = metrics.http_reqs ? metrics.http_reqs.values.count : 0;
  const hitRate = metrics.cache_hit_rate ? metrics.cache_hit_rate.values.rate : 0;

  const summary = {
    scenario: 'simple-test',
    timestamp: new Date().toISOString(),
    config: {
      guard_url: BASE_URL,
      target_rps: TARGET_RPS,
      duration: DURATION,
      cache_hit_ratio_target: CACHE_HIT_RATIO,
    },
    results: {
      total_requests: total,
      actual_rps: metrics.http_reqs ? metrics.http_reqs.values.rate : 0,
      cache_hit_rate: hitRate,
      latency: {
        avg: metrics.http_req_duration ? metrics.http_req_duration.values.avg : 0,
        p50: metrics.http_req_duration ? metrics.http_req_duration.values['p(50)'] : 0,
        p95: metrics.http_req_duration ? metrics.http_req_duration.values['p(95)'] : 0,
        p99: metrics.http_req_duration ? metrics.http_req_duration.values['p(99)'] : 0,
      },
      authz: {
        allowed: metrics.authz_allowed ? metrics.authz_allowed.values.count : 0,
        denied: metrics.authz_denied ? metrics.authz_denied.values.count : 0,
        errors: metrics.authz_errors ? metrics.authz_errors.values.count : 0,
      },
      error_rate: metrics.http_req_failed ? metrics.http_req_failed.values.rate : 0,
    },
  };

  console.log('\n' + '='.repeat(70));
  console.log('RESULTS SUMMARY');
  console.log('='.repeat(70));
  console.log(`Total Requests:  ${summary.results.total_requests}`);
  console.log(`Actual RPS:      ${summary.results.actual_rps.toFixed(2)}`);
  console.log(`Cache Hit Rate:  ${(summary.results.cache_hit_rate * 100).toFixed(2)}%`);
  console.log(`Latency P50:     ${summary.results.latency.p50.toFixed(2)}ms`);
  console.log(`Latency P95:     ${summary.results.latency.p95.toFixed(2)}ms`);
  console.log(`Latency P99:     ${summary.results.latency.p99.toFixed(2)}ms`);
  console.log(`Allowed:         ${summary.results.authz.allowed}`);
  console.log(`Denied:          ${summary.results.authz.denied}`);
  console.log(`Errors:          ${summary.results.authz.errors}`);
  console.log(`Error Rate:      ${(summary.results.error_rate * 100).toFixed(2)}%`);
  console.log('='.repeat(70));

  return {
    'summary.json': JSON.stringify(summary, null, 2),
  };
}
