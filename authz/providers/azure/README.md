# Azure Authorization Provider

This directory contains the Azure RBAC authorization provider for Guard, which enables Kubernetes authorization decisions based on Azure Role-Based Access Control (RBAC).

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [When to Use Guard](#when-to-use-guard)
- [CheckAccess API v2](#checkaccess-api-v2)
- [Custom Resource Definition (CRD) Support](#custom-resource-definition-crd-support)
- [Building the Project](#building-the-project)
- [Linting](#linting)
- [Testing](#testing)
- [E2E Testing](#e2e-testing)
- [Deployment](#deployment)
- [Staging Environment Setup](#staging-environment-setup)
- [Local Testing with CLI](#local-testing-with-cli)
- [Validation and Metrics](#validation-and-metrics)

## Architecture Overview

Guard is a Kubernetes Webhook Authorization server that bridges Azure RBAC with Kubernetes authorization. When configured as the Kubernetes API server's authorization webhook, Guard intercepts `SubjectAccessReview` requests and translates them into Azure CheckAccess API calls.

### Component Structure

```
authz/providers/azure/
├── azure.go              # Main orchestrator (Authorizer struct)
├── options/
│   ├── options.go        # Configuration flags and validation
│   └── fleet.go          # Azure Fleet Manager validation
├── rbac/
│   ├── rbac.go           # Core RBAC logic, token management
│   ├── checkaccess_v2.go # CheckAccess v2 SDK implementation
│   ├── checkaccessreqhelper.go  # Request/response handling
│   └── tokencredential_adapter.go  # Token provider adapter
└── data/
    └── datastore.go      # In-memory cache (BigCache)
```

### Authorization Flow

```
Kubernetes API Server
        │
        ▼
POST /subjectaccessreviews
        │
        ▼
Guard Authzhandler.ServeHTTP()
        │
        ├─► Validate client certificate (org = "azure")
        │
        ▼
Authorizer.Check()
        │
        ├─► System account check → NoOp
        ├─► Skip list check → NoOp
        ├─► Cache lookup → Return cached result
        ├─► Discovery path access → Allow if configured
        ├─► Token refresh if expired
        │
        ▼
Azure CheckAccess API
        │
        ├─► Primary: Cluster resource scope
        ├─► Fallback 1: Managed namespace scope
        └─► Fallback 2: Fleet manager scope
        │
        ▼
Cache result → Return SubjectAccessReviewStatus
```

### Supported Cluster Types

| AuthzMode | Cluster Type       | Resource ID Format                           |
| --------- | ------------------ | -------------------------------------------- |
| `aks`     | Managed Clusters   | `Microsoft.ContainerService/managedClusters` |
| `arc`     | Connected Clusters | `Microsoft.Kubernetes/connectedClusters`     |
| `fleet`   | Fleet Manager      | `Microsoft.ContainerService/fleets`          |

### Token Providers

Guard uses different token providers based on the authorization mode:

- **AKS/Fleet Mode**: `AKSTokenProvider` - Calls cluster's `/authz/token` endpoint
- **ARC Mode**: `ClientCredentialTokenProvider` or `MSITokenProvider`

## When to Use Guard

### Guard IS Required

1. **Azure RBAC Authorization on AKS**

   - Clusters created with `az aks create --enable-azure-rbac`
   - Users are authorized via Azure AD RBAC role assignments
   - Guard translates Kubernetes requests to Azure CheckAccess calls

2. **Azure AD Group-Based Authorization**

   - Kubernetes RBAC alone doesn't resolve Azure AD group memberships
   - Guard resolves groups and passes them to Kubernetes

3. **Fleet Manager Clusters**
   - Clusters managed by Azure Fleet Manager
   - Guard supports fleet-level and member-level authorization

### Guard is NOT Required

1. **Standard Kubernetes RBAC Only**

   - Using only Kubernetes RBAC with client certificates
   - No external identity integration

2. **AKS Without Azure RBAC**

   - Clusters created without `--enable-azure-rbac`
   - Using standard kubeconfig authentication

3. **ServiceAccount-Only Workloads**
   - Pod-to-pod communication with ServiceAccount tokens
   - No user identity involved

### Important Limitation

Guard runs **inside** the AKS cluster as a pod. When running locally for development or testing, the `/authz/token` endpoint requires authentication that is only available to in-cluster ServiceAccounts. See [E2E Testing](#e2e-testing) for testing alternatives.

## CheckAccess API v2

Guard supports two versions of the Azure CheckAccess API:

### v1 API (Legacy)

- Direct HTTP calls to Azure RBAC endpoint
- URL: `{resourceId}/providers/Microsoft.Authorization/checkaccess?api-version=2018-09-01-preview`
- Manual request/response handling

### v2 API (Recommended)

- Uses official Azure SDK: `github.com/Azure/checkaccess-v2-go-sdk`
- Better error handling and structured responses
- Same batching behavior (200 actions per request)

### Enabling v2 API

```bash
guard run \
  --azure.use-checkaccess-v2=true \
  --azure.pdp-endpoint="https://westus2.authorization.azure.net" \
  --azure.pdp-scope="https://authorization.azure.net/.default" \
  # ... other flags
```

### Configuration Flags for v2

| Flag                                 | Description                   | Required |
| ------------------------------------ | ----------------------------- | -------- |
| `--azure.use-checkaccess-v2`         | Enable v2 API                 | Yes      |
| `--azure.pdp-endpoint`               | PDP service endpoint URL      | Yes      |
| `--azure.pdp-scope`                  | OAuth scope for v2 tokens     | Yes      |
| `--azure.checkaccess-v2-api-version` | Optional API version override | No       |

### Key Differences

| Aspect         | v1                  | v2                       |
| -------------- | ------------------- | ------------------------ |
| Client         | Custom HTTP         | Official SDK             |
| Authentication | Bearer token header | `azcore.TokenCredential` |
| Error Handling | HTTP status codes   | Structured errors        |
| Metrics        | Same                | Same (consistent)        |

## Custom Resource Definition (CRD) Support

Guard can authorize access to Custom Resources (CRDs) not defined in the standard Kubernetes API.

### Enabling CRD Support

```bash
guard run \
  --azure.discover-resources=true \
  --azure.allow-custom-resource-type-check=true \
  # ... other flags
```

### How It Works

1. **Resource Discovery**: Guard periodically discovers available resources from the Kubernetes API
2. **Operations Map**: Builds a map of group → resource → verb → action
3. **CRD Detection**: If a requested resource is not in the operations map, it's treated as a custom resource
4. **Action Mapping**: Custom resources use `<clusterType>/customresources/<action>` format

### Azure Role Assignment for CRDs

Create Azure role definitions with data actions for custom resources:

```json
{
  "Name": "Custom Resource Reader",
  "IsCustom": true,
  "Actions": [],
  "DataActions": [
    "Microsoft.ContainerService/managedClusters/customresources/read"
  ],
  "NotDataActions": []
}
```

### Available Data Actions for CRDs

| Action                                 | Verb                     |
| -------------------------------------- | ------------------------ |
| `<clusterType>/customresources/read`   | get, list, watch         |
| `<clusterType>/customresources/write`  | create, update, patch    |
| `<clusterType>/customresources/delete` | delete, deletecollection |

### Subresource Support

Guard also supports subresource authorization when enabled:

```bash
--azure.allow-subresource-type-check=true
```

Supported subresources:

- `pods/logs`
- `pods/exec`
- `pods/portforward`
- `pods/proxy`
- `pods/ephemeralcontainers`
- `pods/attach`
- `deployments/scale`
- `statefulsets/scale`
- `replicasets/scale`

## Building the Project

### Prerequisites

- Docker
- Go 1.25+ (used inside Docker container)
- Make

### Build Commands

```bash
# Format code and build
make

# Build for current platform
make build

# Build for specific platforms
make build-linux_amd64
make build-darwin_arm64
make build-windows_amd64

# Build all platforms
make all-build

# Clean build artifacts
make clean
```

### Output

Binaries are output to: `bin/guard-{OS}-{ARCH}[.exe]`

Examples:

- `bin/guard-linux-amd64`
- `bin/guard-darwin-arm64`
- `bin/guard-windows-amd64.exe`

## Linting

Guard uses `golangci-lint` with additional linters.

### Running the Linter

```bash
make lint
```

### Enabled Linters

- `goconst` - Finds repeated strings that could be constants
- `gofmt` - Checks code formatting
- `goimports` - Checks import ordering
- `unparam` - Finds unused function parameters

### Configuration

The linter runs with:

- 10-minute timeout
- Vendor mode: `GOFLAGS="-mod=vendor"`
- Excludes: `generated.*\.go$`, `client`, `vendor` directories

## Testing

### Unit Tests

```bash
# Run all unit tests
make unit-tests

# Run with coverage (via hack/test.sh)
GUARD_DATA_DIR=/tmp/.guard make unit-tests
```

### Test Configuration

Unit tests require the `GUARD_DATA_DIR` environment variable set.

### CI Pipeline

The complete CI check runs:

```bash
make ci
```

This executes:

1. `make verify` - Verify generated files and modules
2. `make check-license` - Verify Apache 2.0 license headers
3. `make lint` - Run linters
4. `make build` - Build binary
5. `make unit-tests` - Run unit tests

## E2E Testing

### Prerequisites

- Kubernetes cluster (minikube, kind, or real cluster)
- Docker
- Kubeconfig configured
- Environment file at `hack/config/.env` (optional)

### Running E2E Tests

```bash
# Standard e2e tests
make e2e-tests

# With custom parameters
make e2e-tests TEST_ARGS="--selfhosted-operator=false" GINKGO_ARGS="--flakeAttempts=2"

# Run in parallel
make e2e-parallel
```

### Test Framework

- **Framework**: Ginkgo v2 (BDD testing)
- **Assertions**: Gomega

### What E2E Tests Cover

The e2e tests validate Guard deployment for different authentication providers:

1. GitHub authentication
2. GitLab authentication
3. Azure authentication
4. LDAP authentication
5. Token-based authentication
6. Google authentication
7. All providers combined

### Test Structure

```
test/e2e/
├── e2e_suite_test.go    # Suite setup/teardown
├── installer_test.go    # Provider installation tests
├── options_test.go      # Test configuration
├── framework/           # Test utilities
│   ├── framework.go
│   ├── deployment.go
│   ├── service.go
│   └── ...
└── matcher/             # Custom Gomega matchers
```

## Deployment

### Certificate Generation

Guard requires TLS certificates for secure communication with the Kubernetes API server.

```bash
# Generate CA certificate
guard init ca

# Generate server certificate
guard init server --ips=<server-ip> --domains=<server-domain>

# Generate client certificate for Kubernetes
guard init client -o azure
```

### Generate Installer YAML

```bash
guard get installer \
  --auth-providers=azure \
  --authz-providers=azure \
  --azure.authz-mode=aks \
  --azure.resource-id="/subscriptions/.../managedClusters/<name>" \
  --azure.aks-authz-token-url="https://<fqdn>:443/authz/token" \
  > guard-installer.yaml
```

### Apply to Cluster

```bash
kubectl apply -f guard-installer.yaml
```

### Configure Kubernetes API Server

Add the following flags to the Kubernetes API server:

```yaml
--authorization-mode=Webhook
--authorization-webhook-config-file=/path/to/guard-webhook.kubeconfig
```

Generate the webhook configuration:

```bash
guard get webhook-config azure -o <org-name> --addr=<guard-service-addr>
```

### AKS-Specific Deployment

For AKS clusters with Azure RBAC enabled:

```bash
# Create AKS cluster with Azure RBAC
az aks create \
  --resource-group <rg> \
  --name <cluster> \
  --enable-aad \
  --enable-azure-rbac

# Get cluster credentials
az aks get-credentials -g <rg> -n <cluster>

# Extract authz token URL
FQDN=$(az aks show -g <rg> -n <cluster> --query fqdn -o tsv)
AKS_AUTHZ_TOKEN_URL="https://${FQDN}:443/authz/token"
```

### Configuration Flags Reference

| Flag                                         | Description                    | Default |
| -------------------------------------------- | ------------------------------ | ------- |
| `--azure.authz-mode`                         | Cluster type: aks, arc, fleet  | -       |
| `--azure.resource-id`                        | Azure resource ID of cluster   | -       |
| `--azure.aks-authz-token-url`                | Token endpoint URL (aks/fleet) | -       |
| `--azure.arm-call-limit`                     | ARM throttling threshold       | 2000    |
| `--azure.skip-authz-check`                   | Users/groups to skip           | []      |
| `--azure.skip-authz-for-non-aad-users`       | Allow non-AAD users            | true    |
| `--azure.allow-nonres-discovery-path-access` | Allow /api, /openapi           | true    |
| `--azure.discover-resources`                 | Enable resource discovery      | false   |
| `--azure.discover-resources-frequency`       | Discovery refresh interval     | 5m      |
| `--azure.allow-custom-resource-type-check`   | Enable CRD support             | false   |
| `--azure.allow-subresource-type-check`       | Enable subresource perms       | false   |
| `--azure.audit-sar`                          | Log all SAR requests           | false   |

### Prometheus Metrics

Guard exposes metrics at `/metrics`:

| Metric                                             | Description                   |
| -------------------------------------------------- | ----------------------------- |
| `guard_azure_checkaccess_success_total`            | Successful CheckAccess calls  |
| `guard_azure_checkaccess_failure_total`            | Failed calls (by status code) |
| `guard_azure_checkaccess_throttling_failure_total` | HTTP 429 throttling events    |
| `guard_azure_checkaccess_request_duration_seconds` | Call latency histogram        |
| `guard_azure_checkaccess_context_timeout`          | Timeout occurrences           |

### Troubleshooting

**Token Refresh Failures**

- Verify AKS cluster has Azure RBAC enabled
- Check ServiceAccount has proper permissions
- Verify `/authz/token` endpoint is accessible

**Authorization Denied**

- Check Azure role assignments for the user
- Verify resource ID format is correct
- Enable `--azure.audit-sar=true` for debugging

**Rate Limiting (HTTP 429)**

- Reduce request volume
- Check `--azure.arm-call-limit` setting
- Monitor `guard_azure_checkaccess_throttling_failure_total` metric

## Staging Environment Setup

This section describes how to set up a test AKS cluster on the staging environment for CheckAccess v2 validation.

### Prerequisites

- Azure CLI installed and configured
- Access to `AKS INT/Staging Test` subscription
- Sufficient permissions to create AKS clusters and role assignments

### Step 1: Configure Azure CLI

```bash
# Login to Azure
az login

# Set the staging subscription
az account set --subscription 'AKS INT/Staging Test'

# Verify subscription
az account show --query '{name:name, id:id}' -o table
```

### Step 2: Create Resource Group

```bash
# Set environment variables
export RG="guard-test-rg"
export CLUSTER="guard-test-cluster"
export LOCATION="westus2"

# Create resource group
az group create -l $LOCATION -n $RG
```

### Step 3: Create AKS Cluster with Azure RBAC

```bash
# Create AKS cluster with Azure AD and Azure RBAC enabled
az aks create \
    --resource-group $RG \
    --name $CLUSTER \
    --location $LOCATION \
    --generate-ssh-keys \
    --enable-aad \
    --enable-azure-rbac \
    --node-count 1 \
    --node-vm-size Standard_DS2_v2

# Wait for cluster creation (5-10 minutes)
az aks wait --resource-group $RG --name $CLUSTER --created
```

### Step 4: Get Cluster Details

```bash
# Get cluster resource ID
CLUSTER_ID=$(az aks show -g $RG -n $CLUSTER --query id -o tsv)
echo "Cluster Resource ID: $CLUSTER_ID"

# Get cluster FQDN
FQDN=$(az aks show -g $RG -n $CLUSTER --query fqdn -o tsv)
echo "FQDN: $FQDN"

# Get tenant ID
TENANT_ID=$(az aks show -g $RG -n $CLUSTER --query aadProfile.tenantId -o tsv)
echo "Tenant ID: $TENANT_ID"

# Construct authz token URL
AKS_AUTHZ_TOKEN_URL="https://${FQDN}:443/authz/token"
echo "Authz Token URL: $AKS_AUTHZ_TOKEN_URL"

# Verify Azure RBAC is enabled
az aks show -g $RG -n $CLUSTER --query aadProfile.enableAzureRbac -o tsv
```

### Step 5: Assign RBAC Roles for Testing

```bash
# Get current user's object ID
USER_OID=$(az ad signed-in-user show --query id -o tsv)
echo "User OID: $USER_OID"

# Assign Azure Kubernetes Service RBAC Cluster Admin role
az role assignment create \
    --assignee-object-id "$USER_OID" \
    --assignee-principal-type User \
    --role "Azure Kubernetes Service RBAC Cluster Admin" \
    --scope "$CLUSTER_ID"

# Verify role assignment
az role assignment list \
    --assignee "$USER_OID" \
    --scope "$CLUSTER_ID" \
    --query "[].{role:roleDefinitionName, scope:scope}" \
    -o table
```

### Step 6: Get Cluster Credentials

```bash
# Get kubeconfig
az aks get-credentials -g $RG -n $CLUSTER --overwrite-existing

# Extract CA certificate for local testing
az aks get-credentials -g $RG -n $CLUSTER --file ./kubeconfig_tmp --overwrite-existing
awk '/certificate-authority-data:/{print $2}' ./kubeconfig_tmp | base64 --decode > ./ca.crt

# Verify cluster access
kubectl get nodes
```

### Step 7: Get User Token for Testing

```bash
# Request bearer token for AKS server app
TOKEN=$(az account get-access-token \
    --resource 6dae42f8-4368-4678-94ff-3960e28e3630 \
    --query accessToken -o tsv)

# Test API access
curl --silent --show-error \
    --cacert ./ca.crt \
    -H "Authorization: Bearer $TOKEN" \
    "https://${FQDN}/version"
```

### Step 8: Export Environment for Guard Testing

```bash
# Create environment file for Guard testing
cat > guard-test.env << EOF
export RG="$RG"
export CLUSTER="$CLUSTER"
export CLUSTER_ID="$CLUSTER_ID"
export FQDN="$FQDN"
export TENANT_ID="$TENANT_ID"
export AKS_AUTHZ_TOKEN_URL="$AKS_AUTHZ_TOKEN_URL"
export USER_OID="$USER_OID"
export PDP_ENDPOINT="https://westus2.authorization.azure.net"
export PDP_SCOPE="https://authorization.azure.net/.default"
EOF

echo "Environment saved to guard-test.env"
echo "Source it with: source guard-test.env"
```

### Cleanup

```bash
# Delete the test cluster and resource group when done
az group delete -n $RG --yes --no-wait
```

## Local Testing with CLI

This section provides CLI commands for testing Guard authorization locally.

### Build Guard Binary

```bash
cd /path/to/guard-arxhive

# Build for your platform
make build-darwin_arm64   # Apple Silicon
make build-darwin_amd64   # Intel Mac
make build-linux_amd64    # Linux

# Verify build
./bin/guard-darwin-arm64 version
```

### Generate TLS Certificates

```bash
# Create PKI directory
export PKI_DIR="/tmp/guard-test"
mkdir -p $PKI_DIR

# Initialize CA
./bin/guard-darwin-arm64 init ca --pki-dir=$PKI_DIR

# Generate server certificate
./bin/guard-darwin-arm64 init server \
    --pki-dir=$PKI_DIR \
    --ips=127.0.0.1 \
    --domains=localhost,guard-server

# Generate client certificate for testing
./bin/guard-darwin-arm64 init client azure --pki-dir=$PKI_DIR

# Verify certificates
ls -la $PKI_DIR/pki/
```

### Run Guard Server with CheckAccess v2

```bash
# Source environment variables (from staging setup)
source guard-test.env

# Start Guard server
./bin/guard-darwin-arm64 run \
    --ca-cert-file=$PKI_DIR/pki/ca.crt \
    --cert-file=$PKI_DIR/pki/server.crt \
    --key-file=$PKI_DIR/pki/server.key \
    --secure-addr=:8443 \
    --auth-providers=azure \
    --azure.tenant-id="$TENANT_ID" \
    --azure.auth-mode=aks \
    --azure.aks-token-url="$AKS_AUTHZ_TOKEN_URL" \
    --authz-providers=azure \
    --azure.authz-mode=aks \
    --azure.resource-id="$CLUSTER_ID" \
    --azure.aks-authz-token-url="$AKS_AUTHZ_TOKEN_URL" \
    --azure.use-checkaccess-v2=true \
    --azure.pdp-endpoint="$PDP_ENDPOINT" \
    --azure.pdp-scope="$PDP_SCOPE" \
    --azure.skip-authz-for-non-aad-users=true \
    --azure.allow-nonres-discovery-path-access=true \
    -v=5
```

### Health Check

```bash
# Check server health
curl -k https://localhost:8443/healthz

# Check readiness
curl -k https://localhost:8443/readyz
```

### SubjectAccessReview - Allow Test

```bash
# Create SAR request for user with RBAC Admin role
cat > /tmp/sar-allow.json << EOF
{
  "apiVersion": "authorization.k8s.io/v1",
  "kind": "SubjectAccessReview",
  "spec": {
    "user": "test-user@example.com",
    "groups": ["system:authenticated"],
    "extra": {
      "oid": ["$USER_OID"]
    },
    "resourceAttributes": {
      "namespace": "default",
      "verb": "list",
      "group": "",
      "resource": "pods"
    }
  }
}
EOF

# Send request
curl -k \
    --cert $PKI_DIR/pki/azure.crt \
    --key $PKI_DIR/pki/azure.key \
    --cacert $PKI_DIR/pki/ca.crt \
    -X POST \
    -H "Content-Type: application/json" \
    -d @/tmp/sar-allow.json \
    https://localhost:8443/subjectaccessreviews
```

**Expected Response (Allow):**

```json
{
  "apiVersion": "authorization.k8s.io/v1",
  "kind": "SubjectAccessReview",
  "status": {
    "allowed": true,
    "reason": "Access allowed by Azure RBAC Role Assignment..."
  }
}
```

### SubjectAccessReview - Deny Test

```bash
# Use a random UUID for a user with no role assignments
cat > /tmp/sar-deny.json << EOF
{
  "apiVersion": "authorization.k8s.io/v1",
  "kind": "SubjectAccessReview",
  "spec": {
    "user": "no-access-user@example.com",
    "groups": [],
    "extra": {
      "oid": ["00000000-0000-0000-0000-000000000000"]
    },
    "resourceAttributes": {
      "namespace": "kube-system",
      "verb": "delete",
      "group": "apps",
      "resource": "deployments"
    }
  }
}
EOF

curl -k \
    --cert $PKI_DIR/pki/azure.crt \
    --key $PKI_DIR/pki/azure.key \
    --cacert $PKI_DIR/pki/ca.crt \
    -X POST \
    -H "Content-Type: application/json" \
    -d @/tmp/sar-deny.json \
    https://localhost:8443/subjectaccessreviews
```

**Expected Response (Deny):**

```json
{
  "apiVersion": "authorization.k8s.io/v1",
  "kind": "SubjectAccessReview",
  "status": {
    "allowed": false,
    "denied": true,
    "reason": "User does not have access to the resource..."
  }
}
```

### SubjectAccessReview - Missing OID (Error Case)

```bash
cat > /tmp/sar-error.json << EOF
{
  "apiVersion": "authorization.k8s.io/v1",
  "kind": "SubjectAccessReview",
  "spec": {
    "user": "test-user@example.com",
    "groups": ["system:authenticated"],
    "resourceAttributes": {
      "namespace": "default",
      "verb": "list",
      "resource": "pods"
    }
  }
}
EOF

# Missing oid - should return NoOp or error depending on config
curl -k \
    --cert $PKI_DIR/pki/azure.crt \
    --key $PKI_DIR/pki/azure.key \
    --cacert $PKI_DIR/pki/ca.crt \
    -X POST \
    -H "Content-Type: application/json" \
    -d @/tmp/sar-error.json \
    https://localhost:8443/subjectaccessreviews
```

### Custom Resource Authorization Test

```bash
# Test custom resource access (requires --azure.allow-custom-resource-type-check=true)
cat > /tmp/sar-crd.json << EOF
{
  "apiVersion": "authorization.k8s.io/v1",
  "kind": "SubjectAccessReview",
  "spec": {
    "user": "test-user@example.com",
    "groups": ["system:authenticated"],
    "extra": {
      "oid": ["$USER_OID"]
    },
    "resourceAttributes": {
      "namespace": "default",
      "verb": "list",
      "group": "templates.gatekeeper.sh",
      "resource": "constrainttemplates"
    }
  }
}
EOF

curl -k \
    --cert $PKI_DIR/pki/azure.crt \
    --key $PKI_DIR/pki/azure.key \
    --cacert $PKI_DIR/pki/ca.crt \
    -X POST \
    -H "Content-Type: application/json" \
    -d @/tmp/sar-crd.json \
    https://localhost:8443/subjectaccessreviews
```

## Validation and Metrics

### Prometheus Metrics Endpoints

```bash
# Fetch all Guard Azure metrics
curl -k https://localhost:8443/metrics | grep guard_azure
```

### Key Metrics to Monitor

| Metric                                                   | Type      | Description                 |
| -------------------------------------------------------- | --------- | --------------------------- |
| `guard_azure_checkaccess_success_total`                  | Counter   | Successful v2 API calls     |
| `guard_azure_checkaccess_failure_total{code}`            | Counter   | Failed calls by HTTP status |
| `guard_azure_check_access_requests_total{code}`          | Counter   | Total requests by status    |
| `guard_azure_checkaccess_request_duration_seconds{code}` | Histogram | Request latency             |
| `guard_azure_checkaccess_throttling_failure_total`       | Counter   | Throttled (429) requests    |
| `guard_azure_checkaccess_context_timeout`                | Counter   | Context timeout events      |

### Validation Queries

```bash
# Check success rate
curl -k https://localhost:8443/metrics | grep guard_azure_checkaccess_success_total

# Check for failures
curl -k https://localhost:8443/metrics | grep guard_azure_checkaccess_failure_total

# Check latency distribution
curl -k https://localhost:8443/metrics | grep guard_azure_checkaccess_request_duration_seconds
```

### Log Verbosity Levels

| Level | Flag    | Description                                              |
| ----- | ------- | -------------------------------------------------------- |
| Info  | `-v=5`  | Authorization decisions, cache operations, token refresh |
| Debug | `-v=7`  | Request/response details, batch operations               |
| Trace | `-v=10` | Full request/response bodies with correlation IDs        |

### Key Log Messages to Monitor

```bash
# v2 API initialization
grep "CheckAccess v2" guard.log

# Successful authorization
grep "Access allowed" guard.log

# Denied authorization
grep "Access denied" guard.log

# Fallback scenarios
grep "Falling back" guard.log

# Batch processing
grep "batch" guard.log
```

### Validation Checklist

| Check              | Command                     | Expected                           |
| ------------------ | --------------------------- | ---------------------------------- |
| Server starts      | Logs show "listening on"    | No startup errors                  |
| PDP connectivity   | First authorization request | No connection errors               |
| Token acquisition  | Logs after first request    | "Token refreshed successfully"     |
| Allow works        | SAR with valid user         | `"allowed": true`                  |
| Deny works         | SAR with no-access user     | `"allowed": false, "denied": true` |
| Metrics emit       | `/metrics` endpoint         | `guard_azure_*` metrics present    |
| Latency reasonable | Duration histogram          | p99 < 5s                           |
| Batching works     | >200 actions test           | Multiple batch logs                |

### Unit Tests for CheckAccess v2

```bash
# Run CheckAccess v2 unit tests
GUARD_DATA_DIR=/tmp/.guard go test -v ./authz/providers/azure/rbac/... -run TestCheckAccessV2

# Run all Azure authz tests
GUARD_DATA_DIR=/tmp/.guard go test -v ./authz/providers/azure/...

# Run with coverage
GUARD_DATA_DIR=/tmp/.guard go test -v -cover ./authz/providers/azure/...
```
