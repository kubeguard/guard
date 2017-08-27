# Uninstall Kad
Please follow the steps below to uninstall Kad:

1. Delete the various objects created for Kad operator.
```console
$ ./hack/deploy/uninstall.sh
+ kubectl delete deployment -l app=kad -n kube-system
deployment "kad" deleted
+ kubectl delete service -l app=kad -n kube-system
service "kad" deleted
+ kubectl delete serviceaccount -l app=kad -n kube-system
No resources found
+ kubectl delete clusterrolebindings -l app=kad -n kube-system
No resources found
+ kubectl delete clusterrole -l app=kad -n kube-system
No resources found
```

2. Now, wait several seconds for Kad to stop running. To confirm that Kad operator pod(s) have stopped running, run:
```console
$ kubectl get pods --all-namespaces -l app=kad
```
