# Guard

Kubernetes Webhook Authentication/Authorization server.

## Build

```bash
make build              # build for current platform
make build-linux_amd64  # cross-compile
make fmt                # format
make lint               # lint
make ci                 # full CI (verify + lint + build + test)
go test ./authz/providers/azure/rbac/... -v  # unit tests
```

## Known Gotchas

- **Guard CLI flags**: `--tls-ca-file`, `--tls-cert-file`, `--tls-private-key-file` (NOT `--ca-cert-file`)
- **Webhook endpoint**: `/subjectaccessreviews` (NOT the full K8s API path)
- **Client cert**: `-o Azure` org required for mTLS
- **Base image for testing**: Use `alpine:3.20` not `distroless/static` - needs CA certificates
- **Azure SDK**: `BearerTokenPolicy` requires HTTPS for auth tokens

## Commands

- `/guard-staging-test` - deploy Guard + mock PDP to staging and run e2e validation
