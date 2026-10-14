# AI Manager Fallback Authorization — E2E Test Report

**Feature:** AI Manager fallback authorization for Guard (Kubernetes authn/authz webhook)
**Date:** 2026-09-15
**Environment:** AKS standalone (test subscription `8ecadfc9-…`, region `eastus2`)
**Result:** **PASS** — all verifiable stages passed; the interactive authorization
allow/deny transitions were confirmed by hand.

---

## 1. What was tested

Guard is the AKS authentication/authorization webhook. This feature adds a new
option, `--azure.aimanager-resource-id`, so that a regular AKS cluster which has
joined an **AI Manager** can authorize a request against **either**:

1. its own cluster scope (primary), **or**
2. the AI Manager scope (fallback),

with the first `Allowed` winning. This mirrors the existing fleet-member fallback.

---

## 2. Components under test

| Component | Version / identity |
|---|---|
| Guard code | branch `minsha/aimanager/byo` (2 commits ahead of `master`) |
| Guard image | `…/kubeguard/v2/guard:v0.16.37-1` (built from the branch) |
| AKS RP synth | `control-plane-core-synth`, `minsha-local-1789506681`, Rollout **Completed** |
| Test cluster | AI Manager `minsha-aim-e2e` (+ its backing managed cluster), state **Succeeded** |

---

## 3. Results by stage

### Stage 1 — Guard code quality ✅
- `go build ./...` — clean
- `go test ./authz/providers/azure/...` — all pass
- `gofmt` — clean

### Stage 2 — RP integration deployed to standalone ✅
The AKS RP chart change (pass `--azure.aimanager-resource-id`, version-gated to a
Guard image that supports it) was built and deployed to the standalone. The
`control-plane-core-synth` synthesizer rolled out **Completed**.

### Stage 3 — Guard renders and runs correctly ✅
On the AI Manager cluster's control plane, the Guard deployment renders the expected
image and flags, and runs healthy (both replicas `2/2 Running`, **zero** crashes):

```
image: …/kubeguard/v2/guard:v0.16.37-1
--azure.authz-mode=aks
--azure.resource-id=…/managedClusters/minsha-aim-e2e            (primary scope)
--azure.aimanager-resource-id=…/aiManagers/minsha-aim-e2e       (fallback scope)
```

The presence of the `--azure.aimanager-resource-id` flag in the running binary's
startup log — with **no** "unknown flag" error — proves the deployed Guard is the
new code (a Guard without this feature would crash on the flag).

### Stage 4 — Authorization behavior (primary + fallback) ✅ (verified interactively)
Guard evaluates CheckAccess against **production** Azure RBAC, using the resource-id
strings as literal scopes. Using an AAD user (`Azure Kubernetes Service RBAC` role)
running `kubectl get pods` through Guard, the following transitions were observed —
each confirmed in Guard's logs:

| # | Condition | Expected | Observed |
|---|---|---|---|
| A | No matching resource in prod | Deny | **Deny** — primary scope returns 404, fallback has no role |
| B | Managed cluster created in prod (primary scope resolves) | Allow | **Allow** — cluster-scope role honored (primary path) |
| C | Managed cluster deleted (primary 404s again) | Deny | **Deny** |
| D | AI Manager resource + `Azure AIManager RBAC Reader` role assigned | Allow | **Allow** — Guard log shows the grant came from the AI Manager fallback CheckAccess (`…/aiManagers/minsha-aim-e2e/namespaces/…/checkaccess`) |

The A→B→C→D sequence independently proves both the **primary** cluster-scope
authorization and the **AI Manager fallback** authorization.

---

## 4. Key findings / decisions

1. **Version gating is required.** The new flag must be gated behind a Guard image
   version that supports it — an older Guard image crashes on the unknown flag. The
   RP template only emits the flag for Guard ≥ the release containing this feature.
   → **Release ordering:** publish the Guard image first, then bump the RP's Guard
   image tag.

2. **The fallback role must use AI Manager data actions.** A grant via the fallback
   requires a role whose data actions use the `Microsoft.ContainerService/aiManagers/…`
   prefix (e.g. `Azure AIManager RBAC Reader`). The standard
   `…/managedClusters/…` reader roles do **not** match the AI Manager scope.

3. **Resource IDs must match production.** Because Guard calls production Azure RBAC,
   the cluster / AI Manager resource IDs minted in the standalone must also exist in
   production for the scope to resolve.

---

## 5. Conclusion

The AI Manager fallback authorization feature works end-to-end:
- Code builds and passes tests.
- The RP integration deploys cleanly and renders the correct Guard configuration.
- The deployed Guard runs the new code without error.
- Authorization correctly denies when no scope grants access, allows via the
  **primary** cluster scope, and — when the cluster scope does not apply — allows via
  the **AI Manager fallback** scope.

**Recommendation:** proceed with merge, observing the release ordering in finding #1.
