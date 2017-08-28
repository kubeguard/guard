# Uninstall Guard
Please follow the steps below to uninstall Guard:

1. Delete the various objects created for Guard operator.
```console
$ ./hack/deploy/uninstall.sh
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

2. Now, wait several seconds for Guard to stop running. To confirm that Guard operator pod(s) have stopped running, run:
```console
$ kubectl get pods --all-namespaces -l app=guard
```
