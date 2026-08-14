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
- **DataAction strings**: Guard does not own the action namespace. Every action it emits
  must already be published by the resource provider, or it is deniable but not
  grantable (see below)
- **`system:` users skip Azure RBAC entirely**: `authz/providers/azure/azure.go` returns
  NoOpinion for any user whose name starts with `system:`, so node and controller
  identities on a standard AKS cluster never reach the DataAction mapping

## Changing the DataAction Mapping

`getResourceAndAction` / `getDataActions` in
`authz/providers/azure/rbac/checkaccessreqhelper.go` compose the DataAction strings that
Azure RBAC evaluates. A string Guard invents is only authorizable if the RP already
publishes the matching operation.

**Verify every new action before merging:**

```bash
az provider operation show --namespace Microsoft.ContainerService -o json \
  | jq -r '.. | objects | select(.name? // "" | test("<action-suffix>")) | .name'
```

If the action is absent, the change is breaking and there is no customer-side fix:

- ARM rejects it in custom role definitions with `InvalidDataActionOrNotDataAction -
  does not match any of the actions supported by the providers`, so it cannot be
  granted explicitly.
- Built-in `Azure Kubernetes Service RBAC Reader` and `Writer` enumerate leaf actions,
  so they do not match it either. Only wildcard-bearing roles match
  (`managedClusters/*`, `pods/*`, `<resource>/*`), i.e. Admin and Cluster Admin.
- Net effect: denied for every principal on a least-privilege role, remediable only by
  widening the role to a wildcard. The RP operations manifest must ship first.

Registered as of 2026-08-13: `managedClusters/pods/{read,write,delete}`,
`managedClusters/pods/exec/action`,
`managedClusters/certificates.k8s.io/certificatesigningrequests/{read,write,delete}`.
NOT registered: `pods/{attach,portforward,proxy}/action`, `services/proxy/action`,
`nodes/proxy/action`, `certificatesigningrequests/nodeclient/action`.

Two things that make this easy to miss in review:

- **A passing unit test proves nothing here.** Asserting the composed string succeeds
  whether or not Azure can grant it. Pair every mapping change with the `az` check.
- **Standard clusters cannot reproduce it.** Splitting a subresource out of its parent
  action flips previously-allowed requests to denied, but only for principals that
  actually reach the mapping. Clusters whose kubelets authenticate with an Entra ID
  identity instead of a bootstrap token are the ones that break; `system:`-prefixed
  identities mask the regression everywhere else.

## Commands

- `/guard-staging-test` - deploy Guard + mock PDP to staging and run e2e validation
