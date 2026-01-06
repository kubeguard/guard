# Azure Authorization Caching Architecture

Guard implements a two-layer caching architecture to minimize latency and reduce load on the Azure CheckAccess API.

## Layer 1: Guard's Internal Cache (BigCache)

Guard maintains an in-memory cache using BigCache. This is the first layer checked for every authorization request.

| Flag | Default | Description |
|------|---------|-------------|
| `--azure.cache-size-mb` | 50 | Maximum cache size in megabytes |
| `--azure.cache-ttl-minutes` | 3 | Time-to-live for cached entries |

## Layer 2: Kubernetes API Server Webhook Cache

The Kubernetes API server has built-in caching for authorization webhook responses. This provides a second layer of caching before requests even reach Guard.

| Flag | Default | Description |
|------|---------|-------------|
| `--authorization-webhook-cache-authorized-ttl` | 5m | TTL for allowed responses |
| `--authorization-webhook-cache-unauthorized-ttl` | 5m | TTL for denied responses |

> **Note**: The API server only caches HTTP 200 responses. Guard converts deterministic errors (4xx from Azure) into denied responses (HTTP 200 with `Denied: true`) to enable caching at both layers.

## Request Flow

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Kubernetes     │     │     Guard       │     │  Azure RBAC     │
│  API Server     │     │   (Webhook)     │     │  CheckAccess    │
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                       │
         │  1. Check API server  │                       │
         │     webhook cache     │                       │
         │     (Layer 2)         │                       │
         │                       │                       │
         │  [Cache Miss]         │                       │
         │──────────────────────>│                       │
         │                       │                       │
         │                       │  2. Check BigCache    │
         │                       │     (Layer 1)         │
         │                       │                       │
         │                       │  [Cache Miss]         │
         │                       │──────────────────────>│
         │                       │                       │
         │                       │  3. Azure response    │
         │                       │<──────────────────────│
         │                       │                       │
         │                       │  4. Cache in BigCache │
         │                       │                       │
         │  5. Return response   │                       │
         │<──────────────────────│                       │
         │                       │                       │
         │  6. Cache in API      │                       │
         │     server cache      │                       │
```

## Error Handling

| Azure Response | Guard Behavior | Cached? |
|----------------|----------------|---------|
| Success (allowed) | Return `{Allowed: true}` | Yes, both layers |
| Success (denied) | Return `{Allowed: false, Denied: true}` | Yes, both layers |
| 4xx Error | Return `{Allowed: false, Denied: true}` | Yes, both layers |
| 5xx Error | Return HTTP 5xx error | No (transient) |

Deterministic errors (4xx) are cached because they indicate a permanent failure (e.g., invalid subscription, missing permissions). Transient errors (5xx) are not cached because they may resolve on retry.

## Monitoring

Cache behavior can be monitored via Prometheus metrics:

| Metric | Description |
|--------|-------------|
| `guard_azure_authz_cache_hits_total` | Total cache hits |
| `guard_azure_authz_cache_misses_total` | Total cache misses |
| `guard_azure_authz_cache_entries` | Current number of cached entries |
| `guard_azure_authz_cache_errors_cached_total` | 4xx errors cached as denied |

## Recommended Configuration

For production environments with Azure RBAC:

```yaml
# Guard flags
--azure.cache-size-mb=50
--azure.cache-ttl-minutes=3

# Kubernetes API server flags
--authorization-webhook-cache-authorized-ttl=5m
--authorization-webhook-cache-unauthorized-ttl=5m
```

This configuration provides:
- Up to 50MB of cached authorization decisions in Guard
- 3-minute TTL in Guard's cache
- 5-minute TTL in the API server's webhook cache
- Effective caching of both successful and failed authorization checks

## References

- [Kubernetes Authorization Webhook](https://kubernetes.io/docs/reference/access-authn-authz/webhook/)
- [Azure RBAC Overview](https://docs.microsoft.com/en-us/azure/role-based-access-control/overview)
- [Azure RBAC Best Practices](https://docs.microsoft.com/en-us/azure/role-based-access-control/best-practices)
