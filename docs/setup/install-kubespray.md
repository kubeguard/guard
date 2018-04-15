---
title: Install Kubespray
description: Kubespray Guard Install
menu:
  product_guard_0.1.0:
    identifier: install-kubespray
    name: Install in Kubespray
    parent: setup
    weight: 15
product_name: guard
menu_name: product_guard_0.1.0
section_menu_id: setup
---

> New to Guard? Please start [here](/docs/concepts).

# Kubespray Installation Guide

[Kubespray](https://github.com/kubernetes-incubator/kubespray) (originally called Kargo) is a preferred choice for deploying _customized_ production grade Kubernetes clusters, particularly for those who are comfortable with [Ansible](https://www.ansible.com/). What makes it popular is the range of options it offers, you can deploy a cluster on GCE, Azure, OpenStack, AWS, or Baremetal, along with a choice of various network plugins.

## Pre-Requisites

- [Guard](/docs/setup/install.md) installed on OS X or Linux
- [GitHub](https://github.com/collab-uniba/socialcde4eclipse/wiki/How-to-setup-a-GitHub-organization,-project-and-team) Acccount with an organization and team(s)
- [GitHub Token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/) with _public_repo_ and _read:org_ access

## Key Points

- This document shows up the configuration for GitHub as Guard client or authenticator, the configuration for other [authenticators](https://appscode.com/products/guard/0.1.0/guides/authenticator/github/#authenticator-guides) should be similar
- The document can be followed on a running cluster or before creating a new Kubernetes Cluster
- All files referenced in this documents are in [guard](https://github.com/appscode/kubespray/tree/guard) branch of [kubespray](https://github.com/appscode/kubespray) repository


## Guard Setup
### Initialize PKI
This command creates a self-signed certificate authority (CA) and a key.
```console
$ guard init ca
$ tree .guard/
.guard/
└── pki
    ├── ca.crt
    └── ca.key

1 directory, 2 files
$
```

### Initialize Guard Server
To initialize a Guard server, you just need a free cluster IP. Choose an IP from the IP range while installing Kubernetes Cluster through Kubespray. This can be found in [k8s-cluster.yml](https://github.com/appscode/kubespray/blob/guard/inventory/group_vars/k8s-cluster.yml#L99). We have chosen IP `10.233.0.27` in this document.

```console
$ guard init server --ips=10.233.0.27
$ tree .guard/
.guard/
└── pki
    ├── ca.crt
    ├── ca.key
    ├── server.crt
    └── server.key

1 directory, 4 files
$
```

### Initialize Guard Clients
Here, we are using GitHub as a Guard client.
```console
$ guard init client <your_github_org> -o github
$ tree .guard/
.guard/
└── pki
    ├── ca.crt
    ├── ca.key
    ├── server.crt
    ├── server.key
    ├── <your_github_org>@github.crt
    └── <your_github_org>@github.key
1 directory, 6 files
$
```

### Modifications in Kubespray Repository
Clone the [Kubespray](https://github.com/kubernetes-incubator/kubespray) git repository or the already clones repository which you have used to install Kubernetes Cluster.

> Note: All the below files and commands are relative to the root of the git repository. Hyperlinks are also used a lot to make things easier.

* Enable Guard and optionally RBAC policies in [k8s-cluster.yml](https://github.com/appscode/kubespray/blob/guard/inventory/group_vars/k8s-cluster.yml#L21-26)
```console
kube_guard: true
kube_guard_rbac_policies: false
```

> Note: `kube_guard_rbac_policies` if set to `true` assumes that you have teams named `admins` and `developers` in your GitHub Organization. Feel free to edit the files to match your environment. GitHub users in `admins` group get full access to Kubernetes Cluster and users in `developers` get full access to only `default` namespace.

* `kubernetes` Ansible role: Include Guard Auth Token File
```console
$ mkdir roles/kubernetes/master/files
$ guard get webhook-config <your_github_org> -o github --addr=10.233.0.27:443 > roles/kubernetes/master/files/guard_auth_token_file
```

* `kubernetes` Ansible role: Create `guard-file.yml`
```
$ curl -sL https://raw.githubusercontent.com/appscode/kubespray/guard/roles/kubernetes/master/tasks/guard-file.yml > roles/kubernetes/master/tasks/guard-file.yml
```

* `kubernetes` Ansible role: Import `guard-file.yml` as a task in [main.yml](https://github.com/appscode/kubespray/blob/guard/roles/kubernetes/master/tasks/main.yml#L18-L19)
```console
- import_tasks: guard-file.yml
  when: kube_guard|default(false)
```

* `kubernetes` Ansible role: Update Kube API Server Deployment [template](https://github.com/appscode/kubespray/blob/guard/roles/kubernetes/master/templates/manifests/kube-apiserver.manifest.j2#L61-L63) with Guard Auth Token
```console
{% if kube_guard|default(false) %}
    - --authentication-token-webhook-config-file={{ kube_token_dir }}/guard
{% endif %}
```

* `guard-auth-webook` Ansible Role: Add an ansible role to configure Guard

This role enables Guard server on Kubernetes Master nodes.
```console
$ mkdir -pv roles/guard-auth-webook/{files,tasks}
```

* `guard-auth-webook` Ansible Role: Copy all files and directories from [guard-auth-webook](https://github.com/appscode/kubespray/tree/guard/roles/guard-auth-webook) role

At the time of writing this document, the structure looks like this.
```console
$ tree roles/guard-auth-webook
roles/guard-auth-webook
├── files
│   ├── cluster-admin-role.yml
│   ├── default-namespace-admin-role.yml
│   └── guard-installer.yml
└── tasks
    ├── main.yml
    ├── rbac-guard.yml
    └── setup-guard.yml

2 directories, 6 files
$
```

* `guard-auth-webook` Ansible Role: Create Guard Deployment File
```console
$ guard get installer \
  --auth-providers=github \
  --namespace="kube-system" \
  --addr="10.233.0.27:443" > roles/guard-auth-webook/files/guard-installer.yml
```

The generated file creates these Kubernetes _assets_ packaged in a yaml file
- serviceaccount
- clusterrole
- clusterrolebinding
- deployment
- secret
- service

Kubespray does not label the master nodes correctly. See this [issue](https://github.com/kubernetes-incubator/kubespray/issues/2108) for details. You can fix by running the below command.
```console
$ sed -i "s/node-role.kubernetes.io\/master: \"\"/app: guard/" roles/guard-auth-webook/files/guard-installer.yml
```

* Add the new Ansible Role

Update [cluster.yml](https://github.com/appscode/kubespray/blob/guard/cluster.yml#L126-L128) with the `guard-auth-webook` role
```console
- hosts: kube-master[0]
  roles:
    - { role: guard-auth-webook, when: kube_guard }
```

### Run Ansible Playbook to update or install Kubernetes Cluster
```console
$ ansible-playbook -i inventory/hosts cluster.yml
```

## Add a GitHub User
Add a Github user which is in your organization.
```console
$ kubectl config set-credentials <github-user> --token=<your-github-token>
```

## Check if everything went as desired
* When RBAC policies are disabled.
```console
$ kubectl get pods --user <github-user>
Error from server (Forbidden): pods is forbidden: User "<github-user>" cannot list pods in the namespace "default"
$
```
The `Forbidden` message here is a good sign, this means you have been authenticated but you do not have proper authorization. Now, you can start creating [RBAC](https://kubernetes.io/docs/admin/authorization/rbac/) policies.

* When RBAC policies are enabled (your output may slightly vary)
```console
$ kubectl --user <github-user> get nodes
NAME                    STATUS    ROLES     AGE       VERSION
kubernetes-master0      Ready     master    1d        v1.9.5
kubernetes-master1      Ready     master    1d        v1.9.5
kubernetes-worker0      Ready     node      1d        v1.9.5
kubernetes-worker1      Ready     node      1d        v1.9.5
```

## Troubleshooting
### Ensure guard file is present in Kubernetes API Server
```console
$ kubectl -n kube-system exec -it kube-apiserver-kubernetes-lab-master0 -- ls /etc/kubernetes/tokens/guard
/etc/kubernetes/tokens/guard
$

$ grep authentication-token-webhook-config-file /etc/kubernetes/manifests/kube-apiserver.manifest
    - --authentication-token-webhook-config-file=/etc/kubernetes/tokens/guard
$
```

### Check Logs of API Server
```console
$ kubectl -n kube-system logs kube-apiserver-kubernetes-lab-master0
```

### Slack
Join AppsCode [Slack](https://appscode.slack.com) and hit us on `#guard` channel
