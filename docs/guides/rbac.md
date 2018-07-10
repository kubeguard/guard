---
title: RBAC Roles | Guard
description: RBAC Roles
menu:
  product_guard_0.2.1:
    identifier: rbac-guides
    name: RBAC Roles
    parent: guides
    weight: 15
product_name: guard
menu_name: product_guard_0.2.1
section_menu_id: guides
---

# RBAC Roles

Kubernetes 1.6+ comes with a set of pre-defined set of [user-facing roles](https://kubernetes.io/docs/admin/authorization/rbac/#user-facing-roles). You can create `ClusterRoleBinding`s or `RoleBinding`s to grant permissions to your Github teams or Google groups. Say, you have a Github team called `ops`. You want to make the members of this Github team admin of a cluster. You can do that using the following command:

```console
echo "
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: ops-team
subjects:
- kind: Group
  name: ops
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
" | kubectl apply -f -
```
