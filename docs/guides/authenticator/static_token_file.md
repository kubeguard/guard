---
title: Static Token File Authentication | Guard
description: Authenticate into Kubernetes using static token file
menu:
  product_guard_0.1.0-rc.5:
    identifier: static-token-file-authentication
    parent: authenticator-guides
    name: Static Token File
    weight: 10
product_name: guard
menu_name: product_guard_0.1.0-rc.5
section_menu_id: guides
---

# Static Token File Authentication

TO use static token file authentication, you need to set `--token-auth-file` flag of your guard server to a [token file](https://kubernetes.io/docs/admin/authentication/#static-token-file).

You can use the following command with `--token-auth-file` to generate YAMLs for deploying guard server with static token file authentication.

```console
$ guard get installer --token-auth-file=[PATH_TO_TOKEN_FILE]
```
![github-webhook-flow](/docs/images/token-auth-webhook-flow.png)

```json
{
  "apiVersion": "authentication.k8s.io/v1beta1",
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
kubectl config set-credentials [USERNAME] --token=[TOKEN]
```

Or You can add user in .kube/config file

```yaml
...
users:
- name: [USERNAME]
  user:
    token: [TOKEN]
```
