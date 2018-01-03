---
title: Github Authenticator | Guard
description: Authenticate into Kubernetes using Github
menu:
  product_guard_0.1.0-rc.4:
    identifier: github-authenticator
    parent: authenticator-guides
    name: Github
    weight: 10
product_name: guard
menu_name: product_guard_0.1.0-rc.4
section_menu_id: guides
---

# Github Authenticator

TO use Github, you need a client cert with `CommonName` set to Github organization name and `Organization` set to `Github`. To ease this process, use the Guard cli to issue a client cert/key pair.

```console
$ guard init client {org-name} -o Github
```

![github-webhook-flow](/docs/images/github-webhook-flow.png)

```json
{
  "apiVersion": "authentication.k8s.io/v1beta1",
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

To use Github authentication, you can use your personal access token with permissions to read `public_repo` and `read:org`. You can use the following command to issue a token:

```console
$ guard get token -o github
```

Guard uses the token found in `TokenReview` request object to read user's profile information and list of teams this user is member of. In the `TokenReview` response, `status.user.username` is set to user's Github login, `status.user.groups` is set to teams of the organization in client cert of which this user is a member of.
