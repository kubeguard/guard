/*
 * Custom k6 Metrics for Guard Cache Performance Testing
 *
 * Defines metrics for tracking cache behavior, authorization decisions,
 * and performance characteristics during load tests.
 */

import { Counter, Trend, Rate, Gauge } from 'k6/metrics';

// ============================================================================
// Cache Performance Metrics
// ============================================================================

/**
 * Estimated cache hit rate based on response latency.
 * Cache hits typically have latency < 50ms, while misses involve Azure API calls.
 */
export const cacheHitRate = new Rate('cache_hit_rate');

/**
 * Inferred cache hit count based on low latency responses.
 */
export const cacheHitsInferred = new Counter('cache_hits_inferred');

/**
 * Inferred cache miss count based on high latency responses.
 */
export const cacheMissesInferred = new Counter('cache_misses_inferred');

// ============================================================================
// Authorization Metrics
// ============================================================================

/**
 * Count of authorization decisions that returned "allowed".
 */
export const authzAllowed = new Counter('authz_allowed_total');

/**
 * Count of authorization decisions that returned "denied".
 */
export const authzDenied = new Counter('authz_denied_total');

/**
 * Count of authorization errors (non-200 responses, parse errors).
 */
export const authzErrors = new Counter('authz_errors_total');

// ============================================================================
// Latency Metrics
// ============================================================================

/**
 * Authorization request latency in milliseconds.
 */
export const authzLatency = new Trend('authz_latency_ms');

/**
 * Latency for requests identified as cache hits.
 */
export const cacheHitLatency = new Trend('cache_hit_latency_ms');

/**
 * Latency for requests identified as cache misses.
 */
export const cacheMissLatency = new Trend('cache_miss_latency_ms');

// ============================================================================
// Throughput Metrics
// ============================================================================

/**
 * Total successful authorization requests.
 */
export const authzSuccessTotal = new Counter('authz_success_total');

/**
 * Current requests in flight (for concurrency tracking).
 */
export const requestsInFlight = new Gauge('requests_in_flight');

// ============================================================================
// Error Metrics
// ============================================================================

/**
 * Count of HTTP 429 (Too Many Requests) responses.
 */
export const throttledRequests = new Counter('throttled_requests_total');

/**
 * Count of timeout errors.
 */
export const timeoutErrors = new Counter('timeout_errors_total');

/**
 * Count of connection errors.
 */
export const connectionErrors = new Counter('connection_errors_total');

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Cache hit latency threshold in milliseconds.
 * Responses faster than this are considered cache hits.
 */
const CACHE_HIT_THRESHOLD_MS = 50;

/**
 * Record metrics for an authorization response.
 *
 * @param {Object} response - k6 HTTP response object
 * @param {number} duration - Request duration in milliseconds
 */
export function recordAuthzMetrics(response, duration) {
  authzLatency.add(duration);

  // Infer cache hit/miss from latency
  const likelyCacheHit = duration < CACHE_HIT_THRESHOLD_MS;
  cacheHitRate.add(likelyCacheHit ? 1 : 0);

  if (likelyCacheHit) {
    cacheHitsInferred.add(1);
    cacheHitLatency.add(duration);
  } else {
    cacheMissesInferred.add(1);
    cacheMissLatency.add(duration);
  }

  // Record authorization decision
  if (response.status === 200) {
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
  } else if (response.status === 429) {
    throttledRequests.add(1);
    authzErrors.add(1);
  } else {
    authzErrors.add(1);
  }
}

/**
 * Record error metrics based on error type.
 *
 * @param {Error} error - The error that occurred
 */
export function recordErrorMetrics(error) {
  authzErrors.add(1);

  if (error && error.message) {
    if (error.message.includes('timeout')) {
      timeoutErrors.add(1);
    } else if (error.message.includes('connection')) {
      connectionErrors.add(1);
    }
  }
}

/**
 * Generate a summary of custom metrics for reporting.
 *
 * @param {Object} data - k6 summary data
 * @returns {Object} Processed metrics summary
 */
export function generateMetricsSummary(data) {
  const metrics = data.metrics || {};

  const summary = {
    cache: {
      hit_rate: getMetricValue(metrics, 'cache_hit_rate', 'rate'),
      hits_inferred: getMetricValue(metrics, 'cache_hits_inferred', 'count'),
      misses_inferred: getMetricValue(metrics, 'cache_misses_inferred', 'count'),
      hit_latency_avg: getMetricValue(metrics, 'cache_hit_latency_ms', 'avg'),
      miss_latency_avg: getMetricValue(metrics, 'cache_miss_latency_ms', 'avg'),
    },
    authz: {
      allowed: getMetricValue(metrics, 'authz_allowed_total', 'count'),
      denied: getMetricValue(metrics, 'authz_denied_total', 'count'),
      errors: getMetricValue(metrics, 'authz_errors_total', 'count'),
      success_total: getMetricValue(metrics, 'authz_success_total', 'count'),
    },
    latency: {
      avg: getMetricValue(metrics, 'authz_latency_ms', 'avg'),
      p50: getMetricValue(metrics, 'authz_latency_ms', 'p(50)'),
      p95: getMetricValue(metrics, 'authz_latency_ms', 'p(95)'),
      p99: getMetricValue(metrics, 'authz_latency_ms', 'p(99)'),
    },
    errors: {
      throttled: getMetricValue(metrics, 'throttled_requests_total', 'count'),
      timeouts: getMetricValue(metrics, 'timeout_errors_total', 'count'),
      connections: getMetricValue(metrics, 'connection_errors_total', 'count'),
    },
  };

  return summary;
}

/**
 * Helper to safely extract metric values.
 */
function getMetricValue(metrics, name, key) {
  if (metrics[name] && metrics[name].values && metrics[name].values[key] !== undefined) {
    return metrics[name].values[key];
  }
  return 0;
}

// ============================================================================
// Threshold Helpers
// ============================================================================

/**
 * Get thresholds configuration for cache-focused testing.
 */
export function getCacheThresholds() {
  return {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(50)<100', 'p(95)<500', 'p(99)<1000'],
    cache_hit_rate: ['rate>0.7'],
    authz_errors_total: ['count<10'],
    throttled_requests_total: ['count<5'],
  };
}

/**
 * Get thresholds configuration for stress testing.
 */
export function getStressThresholds() {
  return {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<2000'],
    throttled_requests_total: ['count<50'],
  };
}
