---
title: Github Authenticator | Guard
description: Authenticate into Kubernetes using Github
menu:
  product_guard_0.2.0:
    identifier: github-authenticator
    parent: authenticator-guides
    name: Github
    weight: 15
product_name: guard
menu_name: product_guard_0.2.0
section_menu_id: guides
---

# Github Authenticator

Guard installation guide can be found [here](/docs/setup/install.md). To use Github, you need a client cert with `CommonName` set to Github organization name and `Organization` set to `Github`. To ease this process, use the Guard cli to issue a client cert/key pair.

```console
$ guard init client {common-name} -o Github
```

### Deploy Guard Server

To generate installer YAMLs for guard server you can use the following command.

```console
$ guard get installer \
    --auth-providers="github" \
    > installer.yaml

$ kubectl apply -f installer.yaml

```

Additional flags for github:

```console
# Base url for enterprise, keep empty to use default github base url
--github.base-url=<base_url>
```

### Issue Token
To use Github authentication, you can use your personal access token with permissions to read `public_repo` and `read:org`. You can use the following command to issue a token:

```console
$ guard get token -o github
```

![github-token](/docs/images/github-token.png)

Guard uses the token found in `TokenReview` request object to read user's profile information and list of teams this user is member of. In the `TokenReview` response, `status.user.username` is set to user's Github login, `status.user.groups` is set to teams of the organization in client cert of which this user is a member of.

![github-webhook-flow](/docs/images/github-webhook-flow.png)

```json
{
  "apiVersion": "authentication.k8s.io/v1",
  "kind": "TokenReview",
  "status": {
    "authenticated": true,
    "user": {
      "username": "<github-login>",
      "uid": "<github-id>",
      "groups": [
        "<team-1>",
        "<team-2>"
      ]
    }
  }
}
```

### Configure Kubectl
```console
kubectl config set-credentials <user_name> --token=<token>
```

Or You can add user in .kube/confg file

```console
...
users:
- name: <user_name>
  user:
    token: <token>
```

```console
$ kubectl get pods --all-namespaces --user <user_name>
NAMESPACE     NAME                               READY     STATUS    RESTARTS   AGE
kube-system   etcd-minikube                      1/1       Running   0          7h
kube-system   kube-addon-manager-minikube        1/1       Running   0          7h
kube-system   kube-apiserver-minikube            1/1       Running   1          7h
kube-system   kube-controller-manager-minikube   1/1       Running   0          7h
kube-system   kube-dns-6f4fd4bdf-f7csh           3/3       Running   0          7h

```
