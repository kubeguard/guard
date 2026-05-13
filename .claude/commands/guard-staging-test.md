# Staging Test: CheckAccess v2 with Mock PDP

Deploy the current Guard branch to a real AKS cluster in `<SUBSCRIPTION>` subscription and validate CheckAccess v2 end-to-end against a mock PDP service. Execute all steps and show raw logs as confirmation.

### One-time setup

```bash
az account set --subscription '<SUBSCRIPTION>'
az group create -l eastus2 -n <RG>
az acr create --name <ACR> --resource-group <RG> --sku Basic --location eastus2
```

### 1. Build and push images

```bash
# Guard image (use alpine for CA certificates, NOT distroless/static)
GOOS=linux GOARCH=amd64 go build -o bin/guard-linux-amd64 .
cat > bin/Dockerfile.test << 'EOF'
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY guard-linux-amd64 /guard
RUN chmod +x /guard
USER 65534
ENTRYPOINT ["/guard"]
EOF
az acr build --registry <ACR> --image guard:pr-test --file bin/Dockerfile.test bin/

# Mock PDP image
GOOS=linux GOARCH=amd64 go build -o tests/mock-server/mock-server-linux-amd64 ./tests/mock-server/
az acr build --registry <ACR> --image mock-pdp:latest --file tests/mock-server/Dockerfile tests/mock-server/
```

### 2. Create AKS cluster

```bash
az aks create \
    --resource-group <RG> \
    --name guard-v2-staging \
    --enable-aad --enable-azure-rbac \
    --attach-acr <ACR> \
    --node-count 1 \
    --node-vm-size standard_d2s_v5 \
    --location eastus2

az aks get-credentials -g <RG> -n guard-v2-staging --admin --overwrite-existing
```

### 3. Deploy mock PDP + Guard

The mock server replaces external token and authorization services. It runs in dual mode:
- HTTP :8080 for token exchange endpoints
- HTTPS :8443 for PDP checkAccess v2 endpoint (Azure SDK requires TLS for authenticated requests)

```bash
RG=<RG> CLUSTER=guard-v2-staging
CLUSTER_ID=$(az aks show -g $RG -n $CLUSTER --query id -o tsv)
TENANT_ID=$(az aks show -g $RG -n $CLUSTER --query aadProfile.tenantId -o tsv)
PKI_DIR="/tmp/guard-v2-test-pki"

# Build Guard binary for cert generation
GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -o bin/guard-local .

# Generate TLS certs (shared by Guard + mock PDP)
rm -rf "$PKI_DIR" && mkdir -p "$PKI_DIR"
./bin/guard-local init ca --pki-dir=$PKI_DIR
./bin/guard-local init server --pki-dir=$PKI_DIR --ips=127.0.0.1 \
    --domains=mock-pdp.guard-v2-test.svc.cluster.local,mock-pdp.guard-v2-test.svc,guard-v2-test.guard-v2-test.svc,localhost

# Generate client cert for sending test requests
./bin/guard-local init client azure -o Azure --pki-dir=$PKI_DIR

# Create namespace and secrets
kubectl create namespace guard-v2-test
kubectl create secret generic guard-test-pki \
    --from-file=ca.crt=$PKI_DIR/pki/ca.crt \
    --from-file=server.crt=$PKI_DIR/pki/server.crt \
    --from-file=server.key=$PKI_DIR/pki/server.key \
    -n guard-v2-test --dry-run=client -o yaml | kubectl apply -f -
kubectl create secret generic mock-pdp-pki \
    --from-file=ca.crt=$PKI_DIR/pki/ca.crt \
    --from-file=server.crt=$PKI_DIR/pki/server.crt \
    --from-file=server.key=$PKI_DIR/pki/server.key \
    -n guard-v2-test --dry-run=client -o yaml | kubectl apply -f -

# Deploy mock PDP
kubectl apply -f tests/staging/mock-pdp-service.yaml
kubectl rollout status deployment/mock-pdp -n guard-v2-test --timeout=60s

# Deploy Guard with v2 pointing to mock PDP
sed -e "s|PLACEHOLDER_TENANT_ID|$TENANT_ID|g" \
    -e "s|PLACEHOLDER_TOKEN_URL|http://mock-pdp.guard-v2-test.svc.cluster.local:8080/v1/test-ccpid/authztoken|g" \
    -e "s|PLACEHOLDER_CLUSTER_ID|$CLUSTER_ID|g" \
    tests/staging/guard-v2-test.yaml | kubectl apply -f -

# Patch Guard for HTTPS PDP + CA trust
kubectl patch deployment guard-v2-test -n guard-v2-test --type=strategic -p '{
  "spec": {"template": {"spec": {
    "initContainers": [{"name": "ca-setup", "image": "alpine:3.20",
      "command": ["sh", "-c", "cat /etc/ssl/certs/ca-certificates.crt /etc/mock-ca/ca.crt > /shared-certs/ca-bundle.crt"],
      "volumeMounts": [
        {"name": "mock-pdp-ca", "mountPath": "/etc/mock-ca", "readOnly": true},
        {"name": "shared-certs", "mountPath": "/shared-certs"}
      ]}],
    "containers": [{"name": "guard",
      "image": "<ACR>.azurecr.io/guard:pr-test",
      "env": [{"name": "SSL_CERT_FILE", "value": "/shared-certs/ca-bundle.crt"}],
      "args": [
        "run",
        "--tls-ca-file=/etc/guard/pki/ca.crt",
        "--tls-cert-file=/etc/guard/pki/server.crt",
        "--tls-private-key-file=/etc/guard/pki/server.key",
        "--secure-addr=:8443",
        "--auth-providers=azure",
        "--azure.tenant-id='"$TENANT_ID"'",
        "--azure.auth-mode=aks",
        "--azure.aks-token-url=http://mock-pdp.guard-v2-test.svc.cluster.local:8080/v1/test-ccpid/token",
        "--authz-providers=azure",
        "--azure.authz-mode=aks",
        "--azure.resource-id='"$CLUSTER_ID"'",
        "--azure.aks-authz-token-url=http://mock-pdp.guard-v2-test.svc.cluster.local:8080/v1/test-ccpid/authztoken",
        "--azure.use-checkaccess-v2=true",
        "--azure.pdp-endpoint=https://mock-pdp.guard-v2-test.svc.cluster.local/checkaccess/v2",
        "--azure.pdp-scope=https://authorization.azure.net/.default",
        "--azure.skip-authz-for-non-aad-users=true",
        "--azure.allow-nonres-discovery-path-access=true",
        "-v=5"
      ],
      "volumeMounts": [
        {"name": "guard-pki", "mountPath": "/etc/guard/pki", "readOnly": true},
        {"name": "shared-certs", "mountPath": "/shared-certs", "readOnly": true}
      ]}],
    "volumes": [
      {"name": "guard-pki", "secret": {"secretName": "guard-test-pki"}},
      {"name": "mock-pdp-ca", "secret": {"secretName": "mock-pdp-pki"}},
      {"name": "shared-certs", "emptyDir": {}}
    ]
  }}}
}'

kubectl rollout status deployment/guard-v2-test -n guard-v2-test --timeout=120s
```

### 4. Run e2e test

```bash
USER_OID=$(az ad signed-in-user show --query id -o tsv)
UPN=$(az ad signed-in-user show --query userPrincipalName -o tsv)

kubectl port-forward -n guard-v2-test svc/guard-v2-test 8443:8443 &
PF_PID=$!
sleep 3

curl -sk -X POST https://localhost:8443/subjectaccessreviews \
    --cert $PKI_DIR/pki/azure@Azure.crt \
    --key $PKI_DIR/pki/azure@Azure.key \
    --cacert $PKI_DIR/pki/ca.crt \
    -H "Content-Type: application/json" \
    -d '{
      "apiVersion": "authorization.k8s.io/v1",
      "kind": "SubjectAccessReview",
      "spec": {
        "user": "'"$UPN"'",
        "groups": ["system:authenticated"],
        "extra": {"oid": ["'"$USER_OID"'"]},
        "resourceAttributes": {"verb": "list", "resource": "pods", "namespace": "default"}
      }
    }' | jq '.status'

kill $PF_PID 2>/dev/null
```

### 5. Validate with raw logs

```bash
# Guard v2 flow logs
kubectl logs deployment/guard-v2-test -n guard-v2-test --tail=30 | grep -v FLAG:

# Mock PDP logs
kubectl logs deployment/mock-pdp -n guard-v2-test

# Prometheus metrics (api_version label)
kubectl port-forward -n guard-v2-test svc/guard-v2-test 8443:8443 &>/dev/null &
PF_PID=$!; sleep 2
curl -sk https://localhost:8443/metrics | grep api_version
kill $PF_PID 2>/dev/null
```

Expected raw output (confirmed 2026-05-09 on `guard-v2-staging` in `<SUBSCRIPTION>`):

**Guard logs:**

```
I0509 03:44:00.500121  1 server.go:218] setting up authz providers
I0509 03:44:00.500234  1 server.go:234] Initializing authorization cache: size=50MB, ttl=3m
I0509 03:44:20.031595  1 authzhandler.go:39] Recieved subject access review request
I0509 03:44:20.032085  1 azure.go:113] Creating Azure global authz client
I0509 03:44:20.032186  1 rbac.go:323] "Cache miss" key="<USER>@<DOMAIN>/default/-/pods/read"
I0509 03:44:20.073290  1 rbac.go:289] "Token refreshed successfully" expiresAt="2026-05-09 04:39:20 +0000 UTC"
I0509 03:44:20.073311  1 rbac.go:474] "Using CheckAccess v2 API"
I0509 03:44:20.073347  1 checkaccess_v2.go:271] "Performing primary CheckAccess v2" actionsCount=1
I0509 03:44:20.073470  1 tokencredential_adapter.go:46] "Acquiring PDP token" provider="AKSTokenProvider" scope="https://authorization.azure.net/.default"
I0509 03:44:20.115041  1 tokencredential_adapter.go:58] "PDP token acquired successfully" expiresOn="2026-05-09 04:44:20 +0000 UTC"
I0509 03:44:20.185278  1 checkaccess_v2.go:110] "CheckAccess v2 request succeeded" durationSeconds=0.111899644 decisionsCount=1
I0509 03:44:20.185316  1 checkaccess_v2.go:190] "Access allowed via v2 API" roleAssignmentId="/.../roleAssignments/08f06925-..." roleDefinitionId="/.../roleDefinitions/069a667f-..."
I0509 03:44:20.185412  1 azure.go:202] "Authorization check completed" allowed=true
2026/05/09 03:44:20 "POST https://localhost:8443/subjectaccessreviews HTTP/2.0" - 200 644B in 154.267133ms
```

Verify: `"Using CheckAccess v2 API"` present, `"PDP token acquired successfully"` with future expiresOn (no `"invalid expiry"`), `"CheckAccess v2 request succeeded"` with decisionsCount > 0, HTTP 200 not 500.

**Mock PDP logs:**

```
2026/05/09 03:43:58.295313 Mock Azure server starting HTTPS on :8443 (PDP endpoint)
2026/05/09 03:43:58.295385 Mock Azure server starting HTTP on :8080 (OBO token endpoints)
2026/05/09 03:44:20.072969 [OBO-TOKEN] Issued authz token for path /v1/test-ccpid/authztoken (total: 1)
2026/05/09 03:44:20.114813 [OBO-TOKEN] Issued authz token for path /v1/test-ccpid/authztoken (total: 2)
2026/05/09 03:44:20.184797 [V2-CHECKACCESS] Request #1: user=29511a87-... actions=1 resource=/.../managedClusters/guard-v2-staging/namespaces/default allowed=true
```

Verify: both HTTP/HTTPS started, OBO tokens issued (2 calls), V2 checkAccess received.

**Prometheus metrics:**

```
guard_azure_authz_check_access_requests_total{api_version="v2",code="200"} 1
guard_azure_authz_checkaccess_success_total{api_version="v2"} 1
guard_azure_checkaccess_request_duration_seconds_sum{api_version="v2",code="200"} 0.111899644
guard_azure_checkaccess_request_duration_seconds_count{api_version="v2",code="200"} 1
```

Verify: `api_version="v2"` label present, duration reasonable (not 0, not >10s).

### 6. Cleanup

```bash
az group delete -n <RG> --yes --no-wait
```

## Notes

- This test validates: code path routing, token acquisition, ExpiresOn validation, v2 SDK request/response parsing, error handling, cache behavior
- It does NOT validate end-to-end PDP authorization (requires production deployment)
- See `CLAUDE.md` for known gotchas
