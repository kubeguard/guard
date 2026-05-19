# E2E tests

This directory contains Guard's end-to-end test suite.

The recommended local workflow is to use [`hack/scripts/run-e2e-local-image.sh`](../../hack/scripts/run-e2e-local-image.sh), which:

1. builds the current Guard binary from your working tree
2. builds a local Guard container image
3. loads that image into your Kind cluster
4. runs the Ginkgo E2E suite against that exact image

## Prerequisites

At minimum, have these available in the environment where you run the script:

- `go`
- `kubectl`
- `kind` or `kind.exe`
- a container runtime supported by the script:
  - `docker`
  - `docker.exe`
  - `podman`

You also need a reachable Kubernetes cluster context. The examples below use Kind.

## Quick start

Run the full suite against a locally built Guard image:

```bash
hack/scripts/run-e2e-local-image.sh
```

Run a focused subset of specs:

```bash
GINKGO_FOCUS='Set up guard for github should be successful' \
  hack/scripts/run-e2e-local-image.sh
```

Pass extra Ginkgo or test flags after `--` automatically via the script arguments:

```bash
hack/scripts/run-e2e-local-image.sh --ginkgo.v
```

## Important environment variables

The script supports these commonly used variables:

- `GUARD_IMAGE`: image name/tag to build and test
- `CONTAINER_BIN`: container CLI (`docker`, `docker.exe`, or `podman`)
- `KIND_BIN`: Kind CLI (`kind` or `kind.exe`)
- `KUBE_CONTEXT`: kube context used by `kubectl` and the E2E binary
- `KIND_CLUSTER_NAME`: Kind cluster name
- `KIND_LOAD_STRATEGY`: `auto`, `docker-image`, or `image-archive`
- `GINKGO_FOCUS`: regex used to focus specs
- `KUBECONFIG`: optional kubeconfig path

The script auto-detects many of these, but setting them explicitly is often helpful for cross-platform setups.

## Common platform examples

### Linux or macOS with Docker

This is the simplest case when both your container runtime and Kind are available directly in the current shell.

```bash
KUBE_CONTEXT=kind-kind \
CONTAINER_BIN=docker \
KIND_BIN=kind \
KIND_LOAD_STRATEGY=docker-image \
./hack/scripts/run-e2e-local-image.sh
```

### Linux or macOS with Podman

Archive loading is usually the safest option when using Podman.

```bash
KUBE_CONTEXT=kind-kind \
CONTAINER_BIN=podman \
KIND_BIN=kind \
KIND_LOAD_STRATEGY=image-archive \
./hack/scripts/run-e2e-local-image.sh
```

### Windows with WSL, where Docker and Kind are both available inside WSL

If `docker`, `kubectl`, and `kind` are all directly available inside WSL, the setup looks similar to Linux.

```bash
KUBE_CONTEXT=kind-kind \
CONTAINER_BIN=docker \
KIND_BIN=kind \
KIND_LOAD_STRATEGY=docker-image \
./hack/scripts/run-e2e-local-image.sh
```

### Windows with WSL, where Kind runs on Windows and Podman runs in WSL

This is a possible setup if you have a Windows-backed Kind context (e.g. `kind-kind-cluster`) and want to use Podman
from WSL for building and saving the image.

Use:

- `podman` from WSL for build/save
- `kind.exe` from Windows for loading the image archive into the cluster
- a Windows-backed Kind context such as `kind-kind-cluster`

```bash
KUBE_CONTEXT=kind-kind-cluster \
KIND_CLUSTER_NAME=kind-cluster \
CONTAINER_BIN=podman \
KIND_BIN=kind.exe \
KIND_LOAD_STRATEGY=image-archive \
./hack/scripts/run-e2e-local-image.sh
```

## Azure Entra SDK E2Es

The suite includes real-token E2Es for Azure + Entra SDK:

- access token validation
- PoP token validation

These tests are opt-in and are skipped unless the required environment variables are present.

### Shared Azure environment variables

Both Azure Entra SDK E2Es use:

- `AZURE_E2E_ENVIRONMENT`
- `AZURE_E2E_CLIENT_ID`
- `AZURE_E2E_TENANT_ID`

For the current Microsoft public cloud examples:

```bash
export AZURE_E2E_ENVIRONMENT=AzureCloud
export AZURE_E2E_CLIENT_ID=6256c85f-0aad-4d50-b960-e6e9b21efe35
export AZURE_E2E_TENANT_ID='<your-tenant-id>'
```

The example values for:

- `AZURE_E2E_CLIENT_ID`
- `kubelogin --client-id`

come from the Azure Arc kubelogin documentation:

- <https://azure.github.io/kubelogin/concepts/azure-arc.html>

### Azure access token E2E

Required additional variable:

- `AZURE_E2E_ACCESS_TOKEN`

You can mint a regular Azure access token with Azure CLI:

```bash
export AZURE_E2E_ACCESS_TOKEN="$(az account get-access-token \
  --resource 6256c85f-0aad-4d50-b960-e6e9b21efe35 \
  --tenant "$AZURE_E2E_TENANT_ID" \
  --query accessToken -o tsv)"
```

Run only the access-token spec:

```bash
GINKGO_FOCUS='Set up guard for azure with Entra SDK validates a real access token' \
KUBE_CONTEXT=kind-kind \
CONTAINER_BIN=docker \
KIND_BIN=kind \
KIND_LOAD_STRATEGY=docker-image \
./hack/scripts/run-e2e-local-image.sh
```

### Azure PoP token E2E

Required additional variables:

- `AZURE_E2E_POP_HOSTNAME`
- `AZURE_E2E_POP_TOKEN`

`AZURE_E2E_POP_HOSTNAME` must match the value Guard uses for `--azure.pop-hostname`.
For the current staging setup, that value is the user object ID plus tenant ID separated by `@`:

```bash
export AZURE_E2E_POP_HOSTNAME='<your-user-object-id>@<your-tenant-id>'
```

Mint the PoP token outside the test code and pass it through the environment. For example, using WSL `kubelogin`:

```bash
export AZURE_E2E_POP_TOKEN="$(kubelogin get-token \
  --login interactive \
  --client-id 3f4439ff-e698-4d6d-84fe-09c9d574f06b \
  --tenant-id "$AZURE_E2E_TENANT_ID" \
  --server-id 6256c85f-0aad-4d50-b960-e6e9b21efe35 \
  --pop-enabled \
  --pop-claims u="$AZURE_E2E_POP_HOSTNAME" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["status"]["token"])')"
```

Run only the PoP-token spec:

```bash
GINKGO_FOCUS='Set up guard for azure with Entra SDK validates a real PoP token' \
KUBE_CONTEXT=kind-kind \
CONTAINER_BIN=docker \
KIND_BIN=kind \
KIND_LOAD_STRATEGY=docker-image \
./hack/scripts/run-e2e-local-image.sh
```

## Run both Azure Entra SDK specs together

Once all Azure variables are set, you can run both Azure Entra SDK specs with a single focus regex:

```bash
GINKGO_FOCUS='Set up guard for azure with Entra SDK validates a real (access|PoP) token' \
KUBE_CONTEXT=kind-kind \
CONTAINER_BIN=docker \
KIND_BIN=kind \
KIND_LOAD_STRATEGY=docker-image \
./hack/scripts/run-e2e-local-image.sh
```

## Notes

- The script always builds the current working tree, so uncommitted changes are included.
- The script normalizes unqualified local image names to `localhost/...`.
- When using `kind.exe` from WSL, the script automatically converts the archive path to a Windows path before loading.
- The Azure real-token E2Es are intentionally environment-driven. The test code does not fetch tokens itself.

