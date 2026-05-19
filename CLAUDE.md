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

## Pre-commit Verification

**ALWAYS** run before committing. Docker is often unavailable locally so `make ci`
won't work. Run these checks manually instead:

```bash
go build ./...                                                          # build
golangci-lint run ./...                                                 # lint
gofmt -l .                                                              # format (expect no output)
go test ./authz/providers/azure/rbac/... ./auth/providers/azure/graph/... -count=1  # unit tests
```

License headers: all non-vendor `.go`, `.sh`, and `Dockerfile` files need Apache 2.0 headers.
Shell scripts need a blank line between shebang and header (see `hack/license/bash.txt`).
Do NOT commit if any check fails.

## Known Gotchas

- **Guard CLI flags**: `--tls-ca-file`, `--tls-cert-file`, `--tls-private-key-file` (NOT `--ca-cert-file`)
- **Webhook endpoint**: `/subjectaccessreviews` (NOT the full K8s API path)
- **Client cert**: `-o Azure` org required for mTLS
- **Base image for testing**: Use `alpine:3.20` not `distroless/static` - needs CA certificates
- **Azure SDK**: `BearerTokenPolicy` requires HTTPS for auth tokens

## Commands

- `/guard-staging-test` - deploy Guard + mock PDP to staging and run e2e validation
