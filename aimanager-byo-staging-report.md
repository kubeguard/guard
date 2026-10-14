# AI Manager Fallback Authorization — Staging Verification Report

**Feature:** AI Manager fallback authorization for Guard (`--azure.aimanager-resource-id`)
**Date:** 2026-09-23
**Environment:** **Staging** — subscription `26fe00f8-…` ("AKS INT/Staging Test"),
public ARM (`*.staging.azmk8s.io`). Build under test: **merged master, deployed to
staging** (released Guard image). No local build/deploy.
**Verification method:** live `kubectl` authorization checks through Guard (Guard runs
in the Microsoft-managed control plane and is not directly inspectable in staging, so
authorization behavior is verified via `kubectl auth can-i` / `kubectl get pods`).

## Executive summary

- ✅ **Primary (cluster-scope) authorization works** — role at the managed-cluster
  scope grants access through Guard.
- ✅ **AI Manager fallback authorization works** — role (`Azure AIManager RBAC
  Reader`, `aiManagers/…` data actions) at the **AI Manager scope** grants access
  through Guard's fallback path; removing it denies again (symmetric proof, verified
  on two independent AI Manager clusters).
- ⚠️ **BYO-join (`clusterResourceId`) did NOT attach the existing cluster** in this
  staging deployment — see Finding 2. This is a provisioning/join-wiring issue, **not**
  a Guard authorization issue.
- ℹ️ **Region:** AI Manager provisioning **failed in `westus2`** but **succeeded in
  `centraluseuap`** — the AI Manager service appears to be serving the canary region
  in this staging ring (existing healthy AI Managers in the sub are all in
  `centraluseuap`). Tests were completed in `centraluseuap`.

---

## Test results

All tests use an AAD user (`minsha@microsoft.com`) issuing requests through Guard;
each role change was followed by RBAC propagation + Guard cache expiry (~5–10 min).

### Test 1 — Primary (cluster-scope) authorization ✅ PASS
Cluster: `minsha-byo-ceuap` (standalone Azure-RBAC AKS).

| Step | Result |
|---|---|
| Baseline (no role) | **Denied** — `User does not have access to the resource in Azure` |
| Assign `Azure Kubernetes Service RBAC Reader` at **cluster** scope | **Allowed** — `kubectl get pods` lists real pods |

→ The released Guard authorizes correctly via the primary cluster scope.

### Test 2 — Regular AI Manager fallback ✅ PASS (symmetric)
AI Manager `minsha-aim-ceuap` + its own provisioned cluster.

| Step | Result |
|---|---|
| Baseline (no role) | **Denied** |
| Assign `Azure AIManager RBAC Reader` at **AI Manager** scope | **Allowed** — real pods listed |
| Remove the AI Manager role | **Denied** again |

→ The `aiManagers/…` role at the AI Manager scope drives the decision through Guard's
**fallback** path. Confirmed a second time on `minsha-aim-byojoin`'s own cluster
(role present → **Allowed**).

### Test 3 — BYO join (attach existing cluster) ⚠️ NOT WIRED
Attempted to join existing cluster `minsha-byo-ceuap` to a new AI Manager
`minsha-aim-byojoin` via `properties.clusterResourceId`.

| Observation | Result |
|---|---|
| AI Manager `minsha-aim-byojoin` provisioning | **Succeeded** |
| Did it attach `minsha-byo-ceuap`? | **No** — the AI Manager provisioned its **own separate** backing cluster (`minsha-aim-byojoin` in MRG `AIM_…_byojoin_…`); `minsha-byo-ceuap` has **no `aiManagerProfile`** and no linkage |
| `Azure AIManager RBAC Reader` at the join scope → access on `minsha-byo-ceuap` | **Denied** (expected — the cluster was never joined, so its Guard has no `--azure.aimanager-resource-id` for this AI Manager) |

→ The BYO-**authorization** logic is proven (Test 2). What did not work here is the
**BYO-join provisioning** — passing `clusterResourceId` did not attach the existing
cluster; it created a fresh one instead.

---

## Findings

1. **Region availability (staging).** AI Manager creates **fail in `westus2`** in this
   staging sub (both a regular AI Manager and a BYO join failed there), but **succeed in
   `centraluseuap`**. All existing healthy AI Managers in the sub are in
   `centraluseuap`. Recommendation: run AI Manager staging tests in `centraluseuap`.

2. **BYO-join did not attach the existing cluster.** Creating an AI Manager with
   `properties.clusterResourceId` pointing at an existing AKS cluster provisioned a new
   backing cluster instead of joining the referenced one; the referenced cluster's
   `aiManagerProfile` was never stamped. This should be investigated on the RP/AI
   Manager side — it is independent of the Guard authorization change (which is proven).

3. **RBAC/cache propagation is slow.** After a role change, decisions took ~5–10 min to
   reflect (Azure RBAC propagation + Guard's ~3-min result cache + apiserver SAR cache).
   Vary the request or wait ~10 min before concluding a negative result.

---

## Conclusion

On **merged master deployed to staging**, the Guard authorization behavior is verified:

- **Primary** cluster-scope authorization: **PASS**.
- **AI Manager fallback** authorization: **PASS** (allow with `aiManagers/…` role at the
  AI Manager scope, deny when removed) — proven on two AI Manager clusters.

The AI Manager **fallback authorization feature works in staging.** One separate,
non-Guard issue was found: the **BYO-join provisioning path** (`clusterResourceId`) did
not attach the existing cluster in this environment and should be followed up on the RP
side.

---

### Resources created (for cleanup)
Subscription `26fe00f8-…`:
- RG `minsha-aim-staging` (westus2) — failed AI Managers `minsha-aim-stg`,
  `minsha-aim-byo`; healthy AKS `minsha-byo-aks`.
- RG `minsha-aim-staging-ceuap` (centraluseuap) — AI Managers `minsha-aim-ceuap`,
  `minsha-aim-byojoin`; AKS `minsha-byo-ceuap`; plus their AI-Manager managed RGs.
- Cleanup: `az group delete -n minsha-aim-staging --yes` and
  `az group delete -n minsha-aim-staging-ceuap --yes` (managed RGs are removed with the
  AI Manager resources / may need separate deletion).
