#!/bin/bash
set -x

kubectl delete deployment -l app=kad -n kube-system
kubectl delete service -l app=kad -n kube-system

# Delete RBAC objects, if --rbac flag was used.
kubectl delete serviceaccount -l app=kad -n kube-system
kubectl delete clusterrolebindings -l app=kad -n kube-system
kubectl delete clusterrole -l app=kad -n kube-system
