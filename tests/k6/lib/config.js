/*
 * k6 Configuration for Guard Load Testing
 *
 * This module provides TLS configuration and common settings
 * for testing Guard's SubjectAccessReview endpoint.
 */

// Base URL for Guard server
export const baseUrl = __ENV.GUARD_URL || 'https://localhost:8443';

// Mock server URL (for AKS token endpoint simulation)
export const mockServerUrl = __ENV.MOCK_SERVER_URL || 'http://localhost:8080';

// TLS configuration for mTLS with Guard
// Guard requires client certificates with organization field set to "azure"
export function getTLSConfig() {
  // Try to load certificates from environment or default paths
  const certPath = __ENV.CLIENT_CERT || '../certs/client.crt';
  const keyPath = __ENV.CLIENT_KEY || '../certs/client.key';

  return {
    tlsAuth: [
      {
        domains: ['localhost', '127.0.0.1'],
        cert: open(certPath),
        key: open(keyPath),
      },
    ],
    tlsVersion: {
      min: 'tls1.2',
      max: 'tls1.3',
    },
    insecureSkipTLSVerify: true, // For self-signed certs in testing
  };
}

// Common HTTP parameters for SubjectAccessReview requests
export const httpParams = {
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: '30s',
};

// Test configuration presets
export const testPresets = {
  // Quick smoke test
  smoke: {
    vus: 5,
    duration: '30s',
    rps: 10,
  },
  // Standard load test
  standard: {
    vus: 50,
    duration: '5m',
    rps: 100,
  },
  // Stress test
  stress: {
    vus: 200,
    duration: '10m',
    rps: 300,
  },
  // Spike test
  spike: {
    maxVus: 500,
    duration: '5m',
  },
};

// Thresholds for different test types
export const thresholds = {
  // Standard thresholds for production-like testing
  standard: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(50)<100', 'p(95)<500', 'p(99)<1000'],
  },
  // Relaxed thresholds for stress testing
  stress: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<2000'],
  },
  // Cache-focused thresholds
  cache: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(50)<50', 'p(95)<200'],
    cache_hit_rate: ['rate>0.7'],
  },
};

// Cache configuration being tested
export const cacheConfigs = {
  master: {
    name: 'master',
    sizeMB: 5,
    ttlMinutes: 3,
    description: 'Original configuration (5MB cache, 3min TTL)',
  },
  improved_default: {
    name: 'improved_default',
    sizeMB: 50,
    ttlMinutes: 10,
    description: 'Improved default (50MB cache, 10min TTL)',
  },
  improved_large: {
    name: 'improved_large',
    sizeMB: 100,
    ttlMinutes: 10,
    description: 'Large cache (100MB cache, 10min TTL)',
  },
  improved_long_ttl: {
    name: 'improved_long_ttl',
    sizeMB: 50,
    ttlMinutes: 30,
    description: 'Long TTL (50MB cache, 30min TTL)',
  },
};

// Logging utility
export function log(message, level = 'INFO') {
  const timestamp = new Date().toISOString();
  console.log(`[${timestamp}] [${level}] ${message}`);
}
