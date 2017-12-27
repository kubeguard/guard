---
title: Install Kops
description: Kops Guard Install
menu:
  product_guard_0.1.0-rc.4:
    identifier: install-guard
    name: Install-Kops
    parent: getting-started
    weight: 30
product_name: guard
menu_name: product_guard_0.1.0-rc.4
section_menu_id: getting-started
url: /products/guard/0.1.0-rc.4/getting-started/install-kops/
aliases:
  - /products/guard/0.1.0-rc.4/install-kops/
---

> New to Guard? Please start [here](/docs/tutorial.md).

# Kops Installation Guide

Please start [here](/docs/install.md) to get a global installation overview. This document only
shows distinctions during KOPS setup of guard.

For creation of guard server config you need a free cluster ip. There is an easy trick which helps
to find it in most cases: Just find out your nonMasqueradeCIDR through `kops edit cluster --name
<cluster_name>` and then add x.x.10.96 to this range e.g. if it is 100.64.0.0 use 100.64.10.96.

If this does not work for some unknown reason, you have to describe one of your kube-api-server pods
in kube-system namespace and find out ```service-cluster-ip-range```. In this range you can take any
ip which is not already assign. You can show all ips through this command

```kubectl get svc --all-namespaces|grep ClusterIP |awk '{print
$4}'|sort```

```
$ guard init server --ips=100.64.10.96
```

Before you apply your guard config with `kubectl apply -f` verify installer.yaml (`spec/clusterIP: 100.64.10.96`) is filled with rigth ip address.

To configure your api server to use `--authentication-token-webhook-config-file` you need to edit
your kops cluster spec: `kops edit cluster --name <cluster_name>`. There you add the following
specifications.

```
spec:
  kubeAPIServer:
    authenticationTokenWebhookConfigFile: /srv/kubernetes/webhook-guard-config
  fileAssets:
  - content: |
       (OUTPUT of: guard get webhook-config your-github-org -o github --addr=100.64.10.96:9844)
    name: guard-github-auth
    path: /srv/kubernetes/webhook-guard-config
    roles:
    - Master
```

After you saved your config, you have to exchange your kubernetes master nodes. If you have a three
master node cluster, i recommend that you exchange one server with command: `kops rolling-update
cluster <cluster_name> --instance-group master-eu-west-1a--yes`. Now theoretically every third
request could work after your master node is online again. If the node does not join your cluster
or things do not work, ssh to this master node and verify kubernetes api server logs in `/var/log/`.
If some requests are working, exchange the other master nodes. This keeps your cluster working all
the time.

This document only shows difference between kops setup and [here](/docs/tutorial.md).
