---
title: Install Kops
description: Kops Guard Install
menu:
  product_guard_0.2.1:
    identifier: install-kops
    name: Install in Kops
    parent: setup
    weight: 15
product_name: guard
menu_name: product_guard_0.2.1
section_menu_id: setup
---

> New to Guard? Please start [here](/docs/concepts).

# Kops Installation Guide

[Kops](https://github.com/kubernetes/kops) is a popular installer for production grade Kubernetes clusters. Please start [here](/docs/setup/install.md) to get an overview of installation steps. This document only shows distinctions during Kops setup of guard.

## During Initialize PKI
For creation of guard server config you need a free cluster ip. There is an easy trick which helps to find it in most cases: Just find out your nonMasqueradeCIDR through `kops edit cluster --name <cluster_name>` and then add x.x.10.96 to this range e.g. if it is 100.64.0.0 use 100.64.10.96.

If this does not work for some unknown reason, you have to describe one of your kube-api-server pods in kube-system namespace and find out `service-cluster-ip-range`. In this range you can use any ip which is not already assigned. You can show all ips through this command:

```console
kubectl get svc --all-namespaces|grep ClusterIP |awk \'{print $4}\'|sort
```

```console
guard init server --ips=100.64.10.96
```

## During Deploy Guard server
Before you apply your guard config with `kubectl apply -f` verify installer.yaml (`spec/clusterIP: 100.64.10.96`) is filled with rigth ip address.

## During Configure Kubernetes API Server
To configure your api server to use `--authentication-token-webhook-config-file` you need to edit
your kops cluster spec: `kops edit cluster --name <cluster_name>`. There you add the following
specifications:

```yaml
spec:
  kubeAPIServer:
    authenticationTokenWebhookConfigFile: /srv/kubernetes/webhook-guard-config
  fileAssets:
  - content: |
       (OUTPUT of: guard get webhook-config your-github-org -o github --addr=100.64.10.96:443)
    name: guard-github-auth
    path: /srv/kubernetes/webhook-guard-config
    roles:
    - Master
```

After you saved your config, you have to exchange your k8s master nodes. If you have a three
master HA cluster, i recommend that you exchange one server with command: `kops rolling-update
cluster <cluster_name> --instance-group master-eu-west-1a --yes`. Now theoretically every third
request could work after your master node is online again. If the node does not join your cluster
or things do not work, ssh to this master node and verify kubernetes api server logs in `/var/log/`.
If some requests are working, exchange the other master nodes. This keeps your cluster working all
the time.

This document only shows difference between kops setup and [here](/docs/setup/install.md).
