/*
 * SubjectAccessReview Payload Generators for k6 Load Testing
 *
 * Generates realistic Kubernetes SubjectAccessReview payloads
 * that match Guard's expected format for Azure RBAC authorization.
 */

import { SharedArray } from 'k6/data';

// Kubernetes resource types for realistic workload simulation
const NAMESPACES = ['default', 'kube-system', 'kube-public', 'dev', 'staging', 'production', 'monitoring', 'logging'];
const RESOURCES = ['pods', 'deployments', 'services', 'configmaps', 'secrets', 'namespaces', 'nodes', 'persistentvolumeclaims'];
const VERBS = ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete'];
const GROUPS = ['', 'apps', 'extensions', 'rbac.authorization.k8s.io', 'networking.k8s.io', 'batch'];
const SUBRESOURCES = ['', 'status', 'logs', 'exec', 'scale'];

// Generate deterministic UUID from seed for cache key consistency
function generateDeterministicUUID(seed) {
  const hex = seed.toString(16).padStart(8, '0');
  return `${hex.slice(0, 8)}-${hex.slice(0, 4)}-4${hex.slice(1, 4)}-8${hex.slice(1, 4)}-${hex.padStart(12, '0').slice(0, 12)}`;
}

// Generate random UUID for unique requests
function generateRandomUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

/**
 * Generate a pool of test users with deterministic UUIDs.
 * Using a fixed pool ensures consistent cache keys across runs.
 *
 * @param {number} size - Number of users to generate
 * @returns {Array} Array of user objects with email and oid
 */
export function generateUserPool(size) {
  const users = [];
  for (let i = 0; i < size; i++) {
    users.push({
      email: `testuser${i}@contoso.com`,
      oid: generateDeterministicUUID(i + 1000), // Start from 1000 for cleaner UUIDs
      groups: ['system:authenticated', `team-${i % 5}`],
    });
  }
  return users;
}

/**
 * Generate a SubjectAccessReview payload optimized for cache HITS.
 * Uses deterministic combinations to maximize cache reuse.
 *
 * @param {Array} users - User pool from generateUserPool()
 * @param {number} index - Request index for deterministic selection
 * @returns {Object} SubjectAccessReview payload
 */
export function generateCacheHitPayload(users, index) {
  const user = users[index % users.length];
  const nsIndex = index % NAMESPACES.length;
  const resIndex = index % RESOURCES.length;
  const verbIndex = index % VERBS.length;
  const groupIndex = resIndex % GROUPS.length;

  return {
    apiVersion: 'authorization.k8s.io/v1',
    kind: 'SubjectAccessReview',
    spec: {
      user: user.email,
      groups: user.groups,
      resourceAttributes: {
        namespace: NAMESPACES[nsIndex],
        verb: VERBS[verbIndex],
        group: GROUPS[groupIndex],
        resource: RESOURCES[resIndex],
        version: 'v1',
        name: `test-resource-${resIndex}`,
      },
      extra: {
        oid: [user.oid],
      },
    },
  };
}

/**
 * Generate a SubjectAccessReview payload that will result in cache MISS.
 * Uses unique combinations that won't match existing cache entries.
 *
 * @param {number} vuId - Virtual user ID
 * @param {number} iteration - Current iteration number
 * @returns {Object} SubjectAccessReview payload
 */
export function generateCacheMissPayload(vuId, iteration) {
  const uniqueId = `${vuId}-${iteration}-${Date.now()}`;
  const oid = generateRandomUUID();

  return {
    apiVersion: 'authorization.k8s.io/v1',
    kind: 'SubjectAccessReview',
    spec: {
      user: `unique-user-${uniqueId}@contoso.com`,
      groups: ['system:authenticated'],
      resourceAttributes: {
        namespace: `ns-${uniqueId.slice(0, 8)}`,
        verb: VERBS[Math.floor(Math.random() * VERBS.length)],
        group: GROUPS[Math.floor(Math.random() * GROUPS.length)],
        resource: RESOURCES[Math.floor(Math.random() * RESOURCES.length)],
        version: 'v1',
        name: `resource-${uniqueId.slice(0, 8)}`,
      },
      extra: {
        oid: [oid],
      },
    },
  };
}

/**
 * Generate a mixed payload with configurable cache hit ratio.
 * Simulates realistic traffic patterns.
 *
 * @param {Array} users - User pool from generateUserPool()
 * @param {number} vuId - Virtual user ID
 * @param {number} iteration - Current iteration number
 * @param {number} cacheHitRatio - Target cache hit ratio (0.0-1.0)
 * @returns {Object} SubjectAccessReview payload
 */
export function generateMixedPayload(users, vuId, iteration, cacheHitRatio = 0.8) {
  if (Math.random() < cacheHitRatio) {
    return generateCacheHitPayload(users, iteration);
  }
  return generateCacheMissPayload(vuId, iteration);
}

/**
 * Generate a batch of payloads for warm-up scenarios.
 * Creates diverse but deterministic entries to populate the cache.
 *
 * @param {Array} users - User pool from generateUserPool()
 * @param {number} count - Number of payloads to generate
 * @returns {Array} Array of SubjectAccessReview payloads
 */
export function generateWarmupBatch(users, count) {
  const payloads = [];
  for (let i = 0; i < count; i++) {
    // Create diverse combinations
    const userIdx = i % users.length;
    const nsIdx = Math.floor(i / users.length) % NAMESPACES.length;
    const resIdx = Math.floor(i / (users.length * NAMESPACES.length)) % RESOURCES.length;
    const verbIdx = Math.floor(i / (users.length * NAMESPACES.length * RESOURCES.length)) % VERBS.length;

    const user = users[userIdx];

    payloads.push({
      apiVersion: 'authorization.k8s.io/v1',
      kind: 'SubjectAccessReview',
      spec: {
        user: user.email,
        groups: user.groups,
        resourceAttributes: {
          namespace: NAMESPACES[nsIdx],
          verb: VERBS[verbIdx],
          group: GROUPS[resIdx % GROUPS.length],
          resource: RESOURCES[resIdx],
          version: 'v1',
          name: `warmup-resource-${i}`,
        },
        extra: {
          oid: [user.oid],
        },
      },
    });
  }
  return payloads;
}

/**
 * Generate a high-cardinality payload for stress testing.
 * Creates many unique cache keys to test eviction behavior.
 *
 * @param {number} vuId - Virtual user ID
 * @param {number} iteration - Current iteration number
 * @returns {Object} SubjectAccessReview payload
 */
export function generateHighCardinalityPayload(vuId, iteration) {
  const timestamp = Date.now();
  const oid = generateRandomUUID();

  return {
    apiVersion: 'authorization.k8s.io/v1',
    kind: 'SubjectAccessReview',
    spec: {
      user: `stress-user-${vuId}-${iteration}@contoso.com`,
      groups: ['system:authenticated', `stress-group-${vuId}`],
      resourceAttributes: {
        namespace: `stress-ns-${iteration % 100}`,
        verb: VERBS[iteration % VERBS.length],
        group: GROUPS[iteration % GROUPS.length],
        resource: RESOURCES[iteration % RESOURCES.length],
        version: 'v1',
        name: `stress-resource-${timestamp}-${iteration}`,
        subresource: SUBRESOURCES[iteration % SUBRESOURCES.length],
      },
      extra: {
        oid: [oid],
      },
    },
  };
}

/**
 * Calculate expected cache key from payload.
 * Useful for debugging and cache analysis.
 *
 * Guard cache key format: {user}/{namespace}/{group}/{resource}/{action}[/{subresource}]
 *
 * @param {Object} payload - SubjectAccessReview payload
 * @returns {string} Expected cache key
 */
export function calculateCacheKey(payload) {
  const spec = payload.spec;
  const ra = spec.resourceAttributes;
  const oid = spec.extra && spec.extra.oid && spec.extra.oid[0] ? spec.extra.oid[0] : spec.user;

  let key = `${oid}/${ra.namespace || '-'}/${ra.group || '-'}/${ra.resource}/${ra.verb}`;
  if (ra.subresource) {
    key += `/${ra.subresource}`;
  }
  return key;
}

// Export constants for test customization
export const constants = {
  NAMESPACES,
  RESOURCES,
  VERBS,
  GROUPS,
  SUBRESOURCES,
};
