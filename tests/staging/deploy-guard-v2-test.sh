#!/bin/bash

# Copyright The Guard Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

RG="${RG:-akolomeetc}"
CLUSTER="${CLUSTER:-akolomeetc-v2test}"

echo "=== Fetching cluster config ==="
CLUSTER_ID=$(az aks show -g "$RG" -n "$CLUSTER" --query id -o tsv)
FQDN=$(az aks show -g "$RG" -n "$CLUSTER" --query fqdn -o tsv)
TENANT_ID=$(az aks show -g "$RG" -n "$CLUSTER" --query aadProfile.tenantId -o tsv)
AKS_AUTHZ_TOKEN_URL="https://${FQDN}:443/authz/token"

echo "  CLUSTER_ID: $CLUSTER_ID"
echo "  FQDN: $FQDN"
echo "  TENANT_ID: $TENANT_ID"
echo "  AKS_AUTHZ_TOKEN_URL: $AKS_AUTHZ_TOKEN_URL"

echo ""
echo "=== Getting kubeconfig ==="
az aks get-credentials -g "$RG" -n "$CLUSTER" --overwrite-existing --admin

echo ""
echo "=== Granting RBAC admin to current user ==="
USER_OID=$(az ad signed-in-user show --query id -o tsv)
az role assignment create \
    --assignee-object-id "$USER_OID" \
    --assignee-principal-type User \
    --role "Azure Kubernetes Service RBAC Cluster Admin" \
    --scope "$CLUSTER_ID" \
    -o json 2>&1 | jq '{role: .roleDefinitionName}' || true

echo ""
echo "=== Building Guard binary and generating TLS certs ==="
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
GUARD_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
PKI_DIR="/tmp/guard-v2-test-pki"
mkdir -p "$PKI_DIR"

# Build guard binary for cert generation
cd "$GUARD_DIR"
GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -o bin/guard-local .

# Generate certs
./bin/guard-local init ca --pki-dir="$PKI_DIR"
./bin/guard-local init server --pki-dir="$PKI_DIR" \
    --ips=127.0.0.1 \
    --domains=guard-v2-test.guard-v2-test.svc,localhost

echo ""
echo "=== Creating namespace and TLS secret ==="
kubectl create namespace guard-v2-test --dry-run=client -o yaml | kubectl apply -f -
kubectl create secret tls guard-test-pki \
    --cert="$PKI_DIR/pki/server.crt" \
    --key="$PKI_DIR/pki/server.key" \
    --namespace=guard-v2-test \
    --dry-run=client -o yaml | kubectl apply -f -

# Also add CA cert to the secret
kubectl create secret generic guard-test-pki \
    --from-file=ca.crt="$PKI_DIR/pki/ca.crt" \
    --from-file=tls.crt="$PKI_DIR/pki/server.crt" \
    --from-file=tls.key="$PKI_DIR/pki/server.key" \
    --from-file=server.crt="$PKI_DIR/pki/server.crt" \
    --from-file=server.key="$PKI_DIR/pki/server.key" \
    --namespace=guard-v2-test \
    --dry-run=client -o yaml | kubectl apply -f -

echo ""
echo "=== Deploying Guard with CheckAccess v2 ==="
sed -e "s|PLACEHOLDER_TENANT_ID|$TENANT_ID|g" \
    -e "s|PLACEHOLDER_TOKEN_URL|$AKS_AUTHZ_TOKEN_URL|g" \
    -e "s|PLACEHOLDER_CLUSTER_ID|$CLUSTER_ID|g" \
    "$SCRIPT_DIR/guard-v2-test.yaml" | kubectl apply -f -

echo ""
echo "=== Waiting for Guard pod to be ready ==="
kubectl rollout status deployment/guard-v2-test -n guard-v2-test --timeout=120s

echo ""
echo "=== Guard pod status ==="
kubectl get pods -n guard-v2-test -o wide

echo ""
echo "=== Tailing Guard logs (Ctrl+C to stop) ==="
echo "Look for:"
echo "  - 'Using CheckAccess v2 API' (V(0) - routing decision)"
echo "  - 'Acquiring PDP token' (V(5) - token acquisition)"
echo "  - 'CheckAccess v2 request succeeded' (V(5) - PDP response)"
echo "  - NO 'invalid expiry' errors (validates the fix)"
echo ""
kubectl logs -f deployment/guard-v2-test -n guard-v2-test
