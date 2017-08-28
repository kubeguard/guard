#!/bin/bash
set -x

kubectl delete deployment -l app=guard -n kube-system
kubectl delete service -l app=guard -n kube-system

# Delete RBAC objects, if --rbac flag was used.
kubectl delete serviceaccount -l app=guard -n kube-system
kubectl delete clusterrolebindings -l app=guard -n kube-system
kubectl delete clusterrole -l app=guard -n kube-system
