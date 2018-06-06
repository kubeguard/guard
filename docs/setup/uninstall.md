---
title: Uninstall
description: Guard Uninstall
menu:
  product_guard_0.1.3:
    identifier: uninstall-guard
    name: Uninstall
    parent: setup
    weight: 20
product_name: guard
menu_name: product_guard_0.1.3
section_menu_id: setup
---

# Uninstall Guard
Please follow the steps below to uninstall Guard:

- Delete the various objects created for Guard operator.

```console
$ curl -fsSL https://raw.githubusercontent.com/appscode/guard/0.1.3/hack/deploy/uninstall.sh | bash

+ kubectl delete deployment -l app=guard -n kube-system
deployment "guard" deleted
+ kubectl delete service -l app=guard -n kube-system
service "guard" deleted
+ kubectl delete serviceaccount -l app=guard -n kube-system
No resources found
+ kubectl delete clusterrolebindings -l app=guard -n kube-system
No resources found
+ kubectl delete clusterrole -l app=guard -n kube-system
No resources found
```

- Now, wait several seconds for Guard to stop running. To confirm that Guard operator pod(s) have stopped running, run:

```console
$ kubectl get pods --all-namespaces -l app=guard
```
