---
title: Google Authenticator | Guard
description: Authenticate into Kubernetes using Google
menu:
  product_guard_0.1.0-rc.4:
    identifier: google-authenticator
    parent: authenticator-guides
    name: Google
    weight: 15
product_name: guard
menu_name: product_guard_0.1.0-rc.4
section_menu_id: guides
---

# Google Authenticator
TO use Google, you need a client cert with `CommonName` set to Google Apps (now G Suite) domain and `Organization` set to `Google`. To ease this process, use the Guard cli to issue a client cert/key pair.
```console
$ guard init client {domain-name} -o Google
```

![google-webhook-flow](/docs/images/google-webhook-flow.png)
```json
{
  "apiVersion": "authentication.k8s.io/v1beta1",
  "kind": "TokenReview",
  "status": {
    "authenticated": true,
    "user": {
      "username": "john@mycompany.com",
      "uid": "<google-id>",
      "groups": [
        "groups-1@mycompany.com",
        "groups-2@mycompany.com"
      ]
    }
  }
}
```
To use Google authentication, you need a token with the following OAuth scopes:
 - https://www.googleapis.com/auth/userinfo.email
 - https://www.googleapis.com/auth/admin.directory.group.readonly

You can use the following command to issue a token:
```
$ guard get token -o google
```
This will run a local HTTP server to issue a token with appropriate OAuth2 scopes. Guard uses the token found in `TokenReview` request object to read user's profile information and list of Google Groups this user is member of. In the `TokenReview` response, `status.user.username` is set to user's Google email, `status.user.groups` is set to email of Google groups under the domain found in client cert of which this user is a member of.
