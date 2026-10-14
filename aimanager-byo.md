# AI Manager — Bring Your Own Cluster (BYO)

PR reference: https://github.com/kubeguard/guard/pull/444/

## Background

An earlier PR (#444) added first-class support for the AI Manager as its own
authorization target. It introduced a new RBAC mode, `aimanager`, which makes
Guard build CheckAccess requests scoped to the AI Manager resource ID
(`Microsoft.ContainerService/aiManagers/<name>`) and emit data actions with the
`Microsoft.ContainerService/aiManagers/...` prefix. In that mode the AI Manager
is the *only* authorization scope.

## What this feature adds

We now want to support a "BYO" (bring-your-own-cluster) mode. In this mode a
**regular cluster** (an AKS managed cluster) *joins* an AI Manager in a way very
similar to how a member cluster joins a fleet. The important consequence is that
such a cluster has **two authorization scopes at once**:

1. **Its own cluster RBAC** — the normal managed-cluster authorization is still
   evaluated first.
2. **A fallback to the AI Manager RBAC** — if the cluster-scope check does not
   grant access, Guard additionally checks access against the joined AI
   Manager's resource ID.

A request is allowed if **either** scope grants it (OR semantics), with the
cluster scope tried first and the AI Manager scope tried as a fallback.

This mirrors the existing fleet-member fallback in `CheckAccess`
(`rbac.go`), where a managed cluster that has joined a fleet falls back to the
fleet (`fleetManagerResourceId`) scope using `fleets/members` actions.

## Input / configuration

- The AKS RP will pass an **additional** option, `--azure.aimanager-resource-id`:
  the **AI Manager resource ID** to fall back to (analogous to
  `--azure.fleet-resource-id` for fleet members).
- The primary authz mode remains the regular cluster mode (`aks`); the AI
  Manager resource ID is supplemental, not a replacement for
  `--azure.resource-id`.

## Required behavior (parity with existing AI Manager authZ)

The BYO fallback must reproduce everything the standalone `aimanager` mode
already does, but as a *secondary* scope after the cluster check:

- **Scope**: build the CheckAccess `Resource.Id` from the AI Manager resource
  ID, appending the namespace segment for namespaced requests
  (`.../aiManagers/<name>/namespaces/<ns>`).
- **Actions**: generate data actions using the AI Manager cluster type so action
  IDs use the `Microsoft.ContainerService/aiManagers/...` prefix (not the
  managed-cluster prefix), the same way the fleet fallback regenerates actions
  with `fleets/members`.
- **Namespace model**: AI Manager namespaces are first-class ARM resources of
  type `Microsoft.ContainerService/aiManagers/namespaces`, so the plain
  `namespaces/<ns>` segment is correct. AI Manager must remain **excluded from
  the `managedNamespaces` fallback** (as PR #444 established).
- **Both API paths**: implement the fallback in both the v1 (`rbac.go`,
  `CheckAccess`) and v2 (`checkaccess_v2.go`, `checkAccessV2`) code paths, since
  they duplicate the scope/fallback logic.


## AKS RP changes

AKS RP needs to pass `--azure.aimanager-resource-id` to Guard (via the
kube-authwebhook guard-deployment template), gated on the cluster having joined
an AI Manager.

## Design task

Look at the current AI Manager authorization (mode `aimanager`, the
`aiManagers` constant, `getClusterType`, the data-action prefix logic, and the
namespace scope handling) and, for this new BYO feature, cover all of that
existing AI Manager authZ behavior — but invoked as a fallback scope on a
regular cluster, driven by the new AI Manager resource ID option, following the
same structure as the existing fleet-member fallback.


## E2E test

Goal: prove that on an AI-Manager-managed cluster, Guard renders and runs with
`--azure.authz-mode=aks` + `--azure.resource-id=<cluster>` +
`--azure.aimanager-resource-id=<aiManager>`, and that the flag is safely
version-gated so it is never handed to a Guard image that can't parse it.

> A *regular* (managed) AI Manager cluster exercises the same Guard code path as a
> BYO cluster, so it is sufficient to test with `aksdev aimanager create` — no BYO
> join is required.

### Environment

- `~/standalone` — the standalone test environment
  - `azureconfig.yaml` — standalone configuration (region `eastus2`,
    `build_version`, etc.)
  - `svc.yaml` — the SVC/underlay cluster kubeconfig
  - `cx1.yaml` — the CX/CCP-hosting underlay kubeconfig (guard runs here, in each
    cluster's hex CCP namespace)
- Test subscription: `8ecadfc9-d1a3-4ea4-b844-0d9f87e4d7c8`
- Guard code: `/home/minsha/codes/guard`
- AKS RP code: `/home/minsha/codes/aks-rp`
- Tools required: `az` (logged in), `docker`, `kubectl`, `aksdev` (at
  `/usr/local/bin/aksdev`), plus the RP build toolchain (`aksbuilder`).

### Gotchas (read first — these cost hours)

- **snap Docker can't mount `/tmp`.** The ev2 deploy inside `deploybox` fails with
  `exit 127 … no such file or directory`. Run every `localdeploy` with
  `TMPDIR=$HOME/ev2tmp` (a `$HOME` path snap Docker *can* mount). Create it once:
  `mkdir -p ~/ev2tmp`.
- **`localdeploy` needs `~/azureconfig.yaml`.** Symlink it:
  `ln -sf ~/standalone/azureconfig.yaml ~/azureconfig.yaml`.
- **The guard chart is rendered by the `control-plane-core-synth` Eno
  Synthesizer, NOT overlaymgr.** You must `localdeploy --workspace
  ccp/control-plane-core` for template changes to take effect. Deploying only
  overlaymgr will silently leave the old guard template live.
- **Eno only re-synthesizes NEW clusters when a Synthesizer changes.** Existing
  CCPs keep the old render. Always create a *fresh* AI Manager after deploying to
  verify.
- **The deployer agent pod must be schedulable.** If the underlay is CPU-saturated
  (each test cluster's CCP holds ~1-2 CPU across kube-apiservers), the new
  `deployer` pod goes `Pending` and `localdeploy` times out after 30 min with
  `Rollout version … is not updated`. Fix by deleting unused test clusters to free
  CPU (`aksdev cluster delete` / `aksdev aimanager delete`), then re-run.
- **eastus2 capacity.** `AKSCapacityHeavyUsage` can block new AKS clusters (AI
  Manager backing clusters) from provisioning. Not a code issue — retry later or
  free capacity.

### Steps

**0. One-time setup**

```bash
mkdir -p ~/ev2tmp
ln -sf ~/standalone/azureconfig.yaml ~/azureconfig.yaml

# Dev ACR for images (Standard SKU so anonymous pull can be enabled).
SUB=8ecadfc9-d1a3-4ea4-b844-0d9f87e4d7c8
RG=minsha-aimgr-byo-e2e
ACR=minshaaimgrbyoe2e
az group create -n "$RG" -l eastus2 --subscription "$SUB"
az acr create  -n "$ACR" -g "$RG" --sku Standard --admin-enabled true
az acr update  -n "$ACR" --anonymous-pull-enabled true
```

**1. Build + push the Guard image with your code.**
The template gates the flag at `>=0.16.37-0`, so tag the image so its semver
satisfies the gate (i.e. `v0.16.37`+). The RP guard tag lives in
`toolkit/versioning/manifests/ccp/kubeguard-guard/kubeguard-guard.yaml`
(`defaultImageTag`, currently `v0.16.36-1`).

```bash
cd ~/codes/guard
make container VERSION=v0.16.37          # builds ghcr.io/kubeguard/guard:v0.16.37_linux_amd64
az acr login -n minshaaimgrbyoe2e
docker tag  ghcr.io/kubeguard/guard:v0.16.37_linux_amd64 \
            minshaaimgrbyoe2e.azurecr.io/kubeguard/v2/guard:v0.16.37-1
docker push minshaaimgrbyoe2e.azurecr.io/kubeguard/v2/guard:v0.16.37-1
```

**2. Apply the RP changes and regenerate snapshots/fixtures.**
The RP diff is: the guard-deployment template (managed-AIManager branch renders
`authz-mode=aks` + cluster `resource-id` + version-gated `aimanager-resource-id`),
plus `AIManagerValues.IsByoCluster` plumbing in helmvalues. Then regenerate golden
files (these run in the dev-env container, so `az acr login` to the dev ACR first):

```bash
cd ~/codes/aks-rp
az acr login -n aksdevinfrareadonlyeastus            # dev-env image pull
make generate-helm-fixtures
make render-control-plane-chart-snapshots
make render-ccp-plugin-adapter-chart-snapshots
make render-control-plane-chart-eno-resource-snapshots
# Expect: only snapshots/aimanager-cluster.yaml/.../guard-deployment.yaml changes.
# With RP guard tag v0.16.36 (< 0.16.37) the snapshot must render authz-mode=aks +
# cluster resource-id and NO aimanager-resource-id flag (gate closed — correct).
```

**3. Deploy to the standalone** (control-plane-core synth is the one that matters;
overlaymgr optional). Watch for the `Pending` deployer / capacity gotchas above.

```bash
cd ~/codes/aks-rp
TMPDIR=$HOME/ev2tmp bash hack/aksbuilder.sh localdeploy \
  --workspace ccp/control-plane-core --e2eazureconfig ~/standalone/azureconfig.yaml --usespoperator
# Optional: also deploy overlaymgr the same way (--workspace overlaymgr/server).
# Success = aksapp version bumped, Rollout Completed:
KUBECONFIG=~/standalone/cx1.yaml kubectl get aksapp -A | grep control-plane-core-synth
```

**4. Create a fresh AI Manager and inspect the rendered Guard.**
First-time only, register the preview feature (registration is subscription-scoped;
`aksdev cluster create` carrying the feature is the simplest way):

```bash
# Register AIManagerPreview on the sub (one throwaway cluster; delete it after).
aksdev cluster create featreg --azureconfig ~/standalone/azureconfig.yaml \
  --resourcegroup minsha-aimgr-byo-e2e --location eastus2 \
  --subscription-features AIManagerPreview
  
aksdev cluster delete featreg --azureconfig ~/standalone/azureconfig.yaml \
  --resourcegroup minsha-aimgr-byo-e2e --location eastus2

# Find its CCP hex namespace on cx1 (newest) and read the guard args:
NS=$(KUBECONFIG=~/standalone/cx1.yaml kubectl get ns --sort-by=.metadata.creationTimestamp \
      | grep -E '^[0-9a-f]{16,}' | tail -1 | awk '{print $1}')
KUBECONFIG=~/standalone/cx1.yaml kubectl get deploy guard -n "$NS" \
  -o jsonpath='{range .spec.template.spec.containers[0].args[*]}{@}{"\n"}{end}' \
  | grep -E 'authz-mode|resource-id|aimanager'
```

Expected (once the RP guard tag is >= v0.16.37, or after step 5's image patch):
```
--azure.authz-mode=aks
--azure.resource-id=.../managedClusters/<backing-cluster>
--azure.aimanager-resource-id=.../aiManagers/minsha-aim-v2
```

**5. Verify Guard actually runs with the flag.**
The standalone's default guard image is `v0.16.36-1`, which does **not** know the
flag — so if you force the flag on (or bump the RP tag), stock guard CrashLoops with
`unknown flag: --azure.aimanager-resource-id`. Point the CCP guard at your
v0.16.37 image to confirm your code parses it:

```bash
KUBECONFIG=~/standalone/cx1.yaml kubectl set image deploy/guard -n "$NS" \
  guard=minshaaimgrbyoe2e.azurecr.io/kubeguard/v2/guard:v0.16.37-1
KUBECONFIG=~/standalone/cx1.yaml kubectl get pods -n "$NS" | grep guard   # expect 2/2 Running
KUBECONFIG=~/standalone/cx1.yaml kubectl logs -n "$NS" deploy/guard -c guard \
  | grep -E 'unknown flag|FLAG: --azure.aimanager-resource-id|setting up authz'
# Pass = no "unknown flag", all three flags present, "setting up authz providers".
```

### Verifying the authorization end-to-end (the actual test)

Guard evaluates CheckAccess against **production** Azure RBAC — it uses the
`--azure.resource-id` (cluster, primary) and `--azure.aimanager-resource-id`
(AI Manager, fallback) strings as **literal scopes against prod ARM**. The
standalone RP mints these IDs under the test subscription, so the scope only
resolves if the **same resource ID also exists in production**. This lets us prove
both the primary check and the fallback independently.

**Setup — download the cluster kubeconfig + get an AAD identity.**

```bash
# The AI Manager's backing managed cluster lives in its managed RG.
aksdev aimanager get minsha-aim-e2e --azureconfig ~/standalone/azureconfig.yaml \
  --resourcegroup minsha-aimgr-byo-e2e --location eastus2   # → managedResourceGroupName

# Download the clusterUser (AAD) kubeconfig for the backing cluster.
aksdev cluster kubeconfig minsha-aim-e2e --user clusterUser \
  --azureconfig ~/standalone/azureconfig.yaml \
  -g AIM_minsha-aimgr-byo-e2e_minsha-aim-e2e_eastus2 --location eastus2 \
  -s 8ecadfc9-d1a3-4ea4-b844-0d9f87e4d7c8 > /tmp/aim-e2e.kubeconfig

# kubelogin is required for AAD auth (interactive device-code / consent).
export KUBECONFIG=/tmp/aim-e2e.kubeconfig
kubectl get pods -n kube-system   # triggers interactive AAD sign-in
```

Watch Guard's decision live (webhook load-balances across BOTH replicas, so select
by label, not a single pod name; results are cached ~3 min per key):

```bash
NS=<ccp-hex-namespace>   # the cluster's CCP namespace on cx1
KUBECONFIG=~/standalone/cx1.yaml kubectl logs -n "$NS" -l app=guard -c guard \
  --since=10m --prefix -f | grep -iE 'checkaccess|does not have|allowed=|Cache'
```

**Step A — Baseline deny (no matching prod resource).**
Even with the `Azure Kubernetes Service RBAC Admin/Reader` role assigned, `kubectl
get pods` is **denied**. Guard's log shows the primary CheckAccess against
`managedClusters/minsha-aim-e2e` returns **404** (the cluster doesn't exist in prod
ARM), then the AI Manager fallback returns `allowed=false` (no role there yet).
→ Confirms Guard denies when neither scope resolves/grants.

**Step B — Primary check works (create the managed cluster in prod).**
Create the managed cluster with the **exact same resource ID** the standalone minted,
using `az aks create` (the ARM entry is what matters; it need not finish provisioning):

```bash
az aks create \
  -g AIM_minsha-aimgr-byo-e2e_minsha-aim-e2e_eastus2 \
  -n minsha-aim-e2e --location eastus2 \
  --node-count 1 --node-vm-size Standard_D2s_v3 \
  --enable-managed-identity --no-ssh-key --no-wait
```

Now the primary scope (`managedClusters/minsha-aim-e2e`) resolves, so the existing
`AKS RBAC Admin/Reader` role (assigned on the managed RG, which contains the cluster)
is honored. Wait out Guard's cache (~3 min), retry `kubectl get pods` → **allowed**.
→ **Proves the primary cluster-scope authorization works.**

**Step C — Fallback proven (delete the cluster → deny again → grant on AI Manager).**
Delete the prod managed cluster so the primary scope 404s again:

```bash
az aks delete -g AIM_minsha-aimgr-byo-e2e_minsha-aim-e2e_eastus2 -n minsha-aim-e2e --yes
```

Retry after the cache expires → **denied again** (primary gone, AI Manager scope has
no role). Then create the AI Manager resource in prod (so its scope resolves) and
assign an **AI Manager** RBAC role — data actions must use the
`Microsoft.ContainerService/aiManagers/...` prefix, since the fallback emits
`aiManagers/...` actions (the `managedClusters/...` reader role does NOT match):

```bash
AIM_ID="/subscriptions/8ecadfc9-d1a3-4ea4-b844-0d9f87e4d7c8/resourceGroups/minsha-aimgr-byo-e2e/providers/Microsoft.ContainerService/aiManagers/minsha-aim-e2e"

# Create the AI Manager resource ID in prod (supported preview api-versions:
# 2026-04-02-preview / 2026-05-02-preview / 2026-09-02-preview).
az resource create --id "$AIM_ID" --api-version 2026-05-02-preview \
  --location eastus2 --properties '{}'

# Assign the AI Manager RBAC role (aiManagers/... data actions) at the AI Manager scope.
az role assignment create --assignee minsha@microsoft.com \
  --role "Azure AIManager RBAC Reader" --scope "$AIM_ID"
```

Wait out the cache, retry `kubectl get pods` → **allowed** — and Guard's log shows the
allow came from the **AI Manager fallback** CheckAccess POST to
`.../aiManagers/minsha-aim-e2e/namespaces/<ns>/.../checkaccess`.
→ **Proves the AI Manager fallback authorization works.**

> Summary of the three transitions verified:
> deny (no prod resource) → allow (primary, cluster created) → deny (cluster deleted)
> → allow (fallback, AI Manager role assigned).

### Release ordering (important)

The version gate makes rollout safe **only if** the Guard image ships first:

1. Merge + publish the **Guard** image to MCR as `>= v0.16.37`.
2. Only then bump the RP's
   `toolkit/versioning/manifests/ccp/kubeguard-guard/kubeguard-guard.yaml`
   `defaultImageTag` to `>= v0.16.37`.

Until the RP tag reaches `v0.16.37`, the template gate keeps the
`--azure.aimanager-resource-id` flag hidden, so older Guard images never crash.

### Cleanup

```bash
az group delete -n minsha-aimgr-byo-e2e --yes   # removes clusters, AI Managers, ACR
```


