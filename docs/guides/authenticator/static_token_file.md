---
title: Static Token File Authentication | Guard
description: Authenticate into Kubernetes using static token file
menu:
  product_guard_{{ .version }}:
    identifier: static-token-file-authentication
    parent: authenticator-guides
    name: Static Token File
    weight: 10
product_name: guard
menu_name: product_guard_{{ .version }}
section_menu_id: guides
---

# Static Token File Authentication

TO use static token file authentication, you need to set `--token-auth-file` flag of your guard server to a [token file](https://kubernetes.io/docs/admin/authentication/#static-token-file).

You can use the following command with `--token-auth-file` to generate YAMLs for deploying guard server with static token file authentication.

```console
$ guard get installer \
    --auth-providers="token-auth" \
    --token-auth-file=<path_to_the_token_file> \
    > installer.yaml

$ kubectl apply -f installer.yaml
```
![github-webhook-flow](/docs/images/token-auth-webhook-flow.png)

```json
{
  "apiVersion": "authentication.k8s.io/v1",
  "kind": "TokenReview",
  "status": {
    "authenticated": true,
    "user": {
      "username": "<user-name>",
      "uid": "<user-id>",
      "groups": [
        "<group-1>",
        "<group-2>"
      ]
    }
  }
}
```

Guard uses the token found in `TokenReview` request object to get user's information and list of groups this user is member of. In the `TokenReview` response, `status.user.username`, `status.user.uid` and `status.user.groups` are set to username, userid and groups found in token file.

### Token file
Token file is a csv file containing four columns: token, username, user uid and group names. Group names column may be empty or contain multiple names. Token must be unique for each user.

|username |uid      |token                 |List groups user is member of
|---------|---------|----------------------|----------------------------------
|user1    |1123     |alkskjhfdku3jkfhm     |test,dev
|user2    |566      |kjasdfgjkewyucxmj12   |dev
|user3    |7654     |lskdfjldskfnkjhf      |

For above user's, token file is given below:
```console
$ cat token.csv
alkskjhfdku3jkfhm,user1,1123,"test,dev"
kjasdfgjkewyucxmj12,user2,566,dev
lskdfjldskfnkjhf,user3,7654,

```
### Configure Kubectl
```console
kubectl config set-credentials <user_name> --token=<token>
```

Or You can add user in .kube/config file

```yaml
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

> **Note:** If you set up guard only for static token authentication , then you will need a client cert with `Organization` set to `token-auth`. if you set up guard for static token authentication and other auth provider (for example, `--auth-providers="token-auth,github"`), then at first guard will check for static token authentication if not succeeded then it will check for other provider. And for multiple auth providers, if you set permissions based on group names, then please be aware of same group name from different authenticators.
