# AI Manager Fallback Authorization — Master Re-verification Report

**Feature:** AI Manager fallback authorization for Guard (`--azure.aimanager-resource-id`)
**Date:** 2026-09-22
**Environment:** AKS standalone (re-provisioned; managed-cluster sub
`feb5b150-…`, underlay sub `8ecadfc9-…`, region `eastus2`)
**Build under test:** merged **master** (no local build/deploy) — RP synth
`control-plane-core-synth` `minshaebld182139698`, Rollout **Completed**
**Result so far:** **PASS** for the automatable stages (deploy, render, runtime).
The interactive authorization allow/deny re-run is **pending** (see §4).

---

## 1. Purpose

Re-verify the AI Manager fallback feature on the standalone now that all code is
**merged to master and master is deployed** — using the released artifacts, not a
locally built image. No build or deploy was performed this run.

---

## 2. What is deployed (verified)

| Item | Value | Note |
|---|---|---|
| RP synth | `control-plane-core-synth` `minshaebld182139698`, Rollout **Completed** | official build-pipeline version (not a local `minsha-local-*`) |
| Guard image | **`mcr.microsoft.com/kubeguard/v2/guard:v0.16.37-1`** | the **released MCR** image, pulled by the RP (previously a custom ACR image was patched in; this run uses the real one) |

This is a stronger configuration than the prior report: it exercises the merged
master synth + the released Guard image, end-to-end from the RP.

---

## 3. Results by stage

### Stage 1 — Fresh AI Manager provisioned via master ✅
Created a new AI Manager `minsha-aim-master` (RG `minsha-aimgr-byo-e2e`, sub
`feb5b150-…`). The RP synthesized its control plane; Guard was rendered in the
cluster's CCP namespace `6ab1c4ecb44a1b0001f46d91` on the underlay.

> Note: the standalone was re-provisioned since the previous report — the
> customer-resource subscription changed to `feb5b150-…`, and the old
> `minsha-aimgr-byo-e2e` RG/ACR were gone. The RG + `AIManagerPreview` feature were
> re-registered on the new sub before creating the AI Manager.

### Stage 2 — Guard renders correctly from master ✅
The Guard deployment renders the released image and all three flags:

```
image: mcr.microsoft.com/kubeguard/v2/guard:v0.16.37-1
--azure.authz-mode=aks
--azure.resource-id=…/managedClusters/minsha-aim-master        (primary scope)
--azure.aimanager-resource-id=…/aiManagers/minsha-aim-master   (fallback scope)
```

### Stage 3 — Guard runs healthy on the released image ✅
- Both replicas **2/2 Running**.
- Startup log contains `FLAG: --azure.aimanager-resource-id=…` and
  `setting up authz providers`.
- **Zero** `unknown flag` errors → the released Guard image supports the flag (the
  version gate + release ordering worked correctly).

---

## 4. Authorization via the AI Manager fallback — PASS ✅ (live, on master)

The live authorization test was run against `minsha-aim-master` (state `Succeeded`)
using an AAD-authenticated user (`minsha@microsoft.com`, interactive sign-in):

1. **Setup (automated):** created the AI Manager resource
   `…/aiManagers/minsha-aim-master` in the production subscription (`feb5b150-…`) so
   the fallback scope resolves, and assigned the role **`Azure AIManager RBAC
   Reader`** (whose data actions use the `Microsoft.ContainerService/aiManagers/…`
   prefix, e.g. `aiManagers/pods/read`) to the user at that scope.
   - No managed cluster was created in prod, so the **primary** scope
     (`managedClusters/minsha-aim-master`) deliberately does **not** resolve (404).
     This isolates the test to the **fallback** path.
2. **Action (interactive):** `kubectl get pods -n kube-system` through Guard →
   **allowed** (pods listed).
3. **Proof (Guard log, CCP `6ab1c4ecb44a1b0001f46d91`):**

   ```
   checkaccessreqhelper.go:801  "Access allowed via role assignment"  user="minsha@microsoft.com"
                                 roleAssignmentId="63bfe1d1…"
   rbac.go:632                  "AI Manager check access allowed"        ← the new fallback code path
   azure.go:202                 "Authorization check completed"  allowed=true
                                 reason="Access allowed by Azure RBAC Role Assignment 63bfe1d1… to user minsha@microsoft.com"
                                 resourceAttributes={Namespace:kube-system, Verb:list, Resource:pods}
   ```

The log line **`rbac.go:632 "AI Manager check access allowed"`** is emitted only by
the new fallback code — confirming the grant came through the **AI Manager fallback
scope**, not the primary cluster scope. This proves the feature works end-to-end on
the released master image.

**Negative test (role removed → deny):** the `Azure AIManager RBAC Reader`
assignment was then deleted from the AI Manager scope, Guard's cache cleared, and the
same request re-issued → **denied**:

```
azure.go:202  "Authorization check completed"  allowed=false
              reason="User does not have access to the resource in Azure. Update role assignment to allow access."
              resourceAttributes={Namespace:kube-system, Verb:list, Resource:pods}
```

This symmetric result (role present → allow via fallback; role absent → deny)
confirms the `aiManagers/…` role assignment at the AI Manager scope is precisely what
drives the authorization decision through Guard's fallback path.

> Caching note: both the kube-apiserver (SubjectAccessReview cache) and Guard (result
> cache, ~3 min) memoize decisions per (user, verb, resource, namespace). A repeated
> identical request will not re-invoke Guard — vary the request or restart Guard to
> force a fresh evaluation.

> The full deny → allow-primary → deny → allow-fallback matrix was proven in the
> previous report on the (identical) branch build. This run re-confirms the critical
> **fallback allow** path on **merged master with the released Guard image**.

---

## 5. Conclusion

On merged **master** with the **released** Guard image, the feature works end-to-end:
- The RP renders `--azure.authz-mode=aks` + cluster `resource-id` +
  `--azure.aimanager-resource-id` for an AI Manager cluster.
- The released `mcr.microsoft.com/kubeguard/v2/guard:v0.16.37-1` image runs healthy
  (2/2), parses the new flag, and does not crash — confirming the release ordering
  (Guard image published before the RP tag bump) is correct.
- A live AAD-authenticated request is **authorized via the AI Manager fallback
  scope** (Guard log: `"AI Manager check access allowed"`, `allowed=true`), granted
  by an `aiManagers/…` role assignment.

**Verdict: PASS.** The AI Manager fallback authorization feature is verified on
master in the standalone environment.

---

### Resources created this run (for cleanup)
- AI Manager `minsha-aim-master` + backing cluster (RG `minsha-aimgr-byo-e2e`, sub
  `feb5b150-…`); CCP ns `6ab1c4ecb44a1b0001f46d91`.
- Throwaway `featreg` cluster — already deleted.
- Cleanup: `az group delete -n minsha-aimgr-byo-e2e --yes` (in sub `feb5b150-…`).
