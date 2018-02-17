---
title: Gitlab Authenticator | Guard
description: Authenticate into Kubernetes using Gitlab
menu:
  product_guard_0.1.0-rc.5:
    identifier: gitlab-authenticator
    parent: authenticator-guides
    name: Gitlab
    weight: 20
product_name: guard
menu_name: product_guard_0.1.0-rc.5
section_menu_id: guides
---

# Gitlab Authenticator

TO use Gitlab, you need a client cert with `Organization` set to `Gitlab`. To ease this process, use the Guard cli to issue a client cert/key pair.

```console
$ guard init client {common-name} -o Gitlab
```
![gitlab-webhook-flow](/docs/images/gitlab-webhook-flow.png)

```json
{
  "apiVersion": "authentication.k8s.io/v1beta1",
  "kind": "TokenReview",
  "status": {
    "authenticated": true,
    "user": {
      "username": "<gitlab-login>",
      "uid": "<gitlab-id>",
      "groups": [
        "<group-1>",
        "<group-2>"
      ]
    }
  }
}
```

To use Gitlab authentication, you can use your personal access token with scope `api`. You can use the following command to issue a token:

```console
$ guard get token -o gitlab
```

Guard uses the token found in `TokenReview` request object to read user's profile information and list of groups this user is member of. In the `TokenReview` response, `status.user.username` is set to user's Gitlab login, `status.user.groups` is set to the list of the groups where this user is a member.
