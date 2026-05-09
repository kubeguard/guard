# Guard

Kubernetes Webhook Authentication/Authorization server for AKS clusters with Azure AD integration.

## Architecture

- Guard runs in the **CCP namespace on the cx-underlay**, co-located with kube-apiserver
- Deployed via CCP Helm chart: `aks-rp/ccp/control-plane-core/charts/kube-control-plane/charts/kube-authwebhook/`
- Credentials via `guard-secrets` Secret (client-id, tenant-id)
- Token exchange via OBO service: `http://obo.<namespace>.svc.cluster.local`

## Known Gotchas

- **VM size in staging**: `Standard_DS2_v2` NOT allowed - use `standard_d2s_v5`
- **Region in staging**: `westus2` has VHD bugs - prefer `eastus2`
- **Base image**: Use `alpine:3.20` not `distroless/static` - needs CA certificates
- **Guard CLI flags**: `--tls-ca-file`, `--tls-cert-file`, `--tls-private-key-file` (NOT `--ca-cert-file`)
- **Webhook endpoint**: `/subjectaccessreviews` (NOT the full K8s API path)
- **Client cert**: `-o Azure` org required for mTLS
- **Azure SDK**: `BearerTokenPolicy` requires HTTPS for auth tokens - PDP must be TLS even in test

## Skills

- `/guard-staging-test` - deploy Guard + mock PDP to AKS INT/Staging and run e2e validation
