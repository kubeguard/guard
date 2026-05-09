#!/bin/bash
set -euo pipefail

echo "=== Port-forwarding to Guard pod ==="
kubectl port-forward -n guard-v2-test svc/guard-v2-test 8443:8443 &
PF_PID=$!
sleep 3

cleanup() { kill $PF_PID 2>/dev/null; }
trap cleanup EXIT

echo ""
echo "=== Health check ==="
curl -sk https://localhost:8443/healthz && echo " OK" || echo " FAILED"

echo ""
echo "=== Getting user token ==="
USER_OID=$(az ad signed-in-user show --query id -o tsv)
UPN=$(az ad signed-in-user show --query userPrincipalName -o tsv)

echo "  User: $UPN"
echo "  OID: $USER_OID"

echo ""
echo "=== Sending SubjectAccessReview (list pods - should ALLOW) ==="
curl -sk -X POST https://localhost:8443/apis/authorization.k8s.io/v1/subjectaccessreviews \
    -H "Content-Type: application/json" \
    -d '{
      "apiVersion": "authorization.k8s.io/v1",
      "kind": "SubjectAccessReview",
      "spec": {
        "user": "'"$UPN"'",
        "groups": ["system:authenticated"],
        "extra": {"oid": ["'"$USER_OID"'"]},
        "resourceAttributes": {
          "verb": "list",
          "resource": "pods",
          "namespace": "default"
        }
      }
    }' | jq '.status'

echo ""
echo "=== Sending SubjectAccessReview (delete nodes - should ALLOW for cluster admin) ==="
curl -sk -X POST https://localhost:8443/apis/authorization.k8s.io/v1/subjectaccessreviews \
    -H "Content-Type: application/json" \
    -d '{
      "apiVersion": "authorization.k8s.io/v1",
      "kind": "SubjectAccessReview",
      "spec": {
        "user": "'"$UPN"'",
        "groups": ["system:authenticated"],
        "extra": {"oid": ["'"$USER_OID"'"]},
        "resourceAttributes": {
          "verb": "delete",
          "resource": "nodes"
        }
      }
    }' | jq '.status'

echo ""
echo "=== Check Guard logs for v2 activity ==="
kubectl logs deployment/guard-v2-test -n guard-v2-test --tail=20 | grep -E "CheckAccess v2|Acquiring PDP|invalid expiry|Using CheckAccess"
