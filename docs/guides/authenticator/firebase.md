---
title: Firebase Authenticator | Guard
description: Authenticate into Kubernetes using Firebase
menu:
  product_guard_{{ .version }}:
    identifier: firebase-authenticator
    parent: authenticator-guides
    name: Firebase
    weight: 20
product_name: guard
menu_name: product_guard_{{ .version }}
section_menu_id: guides
---

# Firebase Authenticator
To use Guard with Firebase, you need to have a Firebase project and a Service Account that has the `firebaseauth.viewer` auth role.
You can use the [Firebase console](https://console.firebase.google.com/) to view your existing projects or create new ones.
Use the [GCP console](https://console.cloud.google.com/iam-admin/serviceaccounts) to create Service account:
- Click Create Service Account.
- In the Service account name field, enter a name. (The service account ID is completed automatically, and you do not need to include a description)
- Click Create.
- Select 'Firebase Authentication Viewer' Role from the Service account permissions drop-down box.
- User permissions are not required for Guard to work with Firebase. You can skip this step.
- Click Create Key. Ensure the key type is set to JSON.

You should now have gathered your service account's Private Key file, Client ID and email address. You are ready to setup Gurad with Firebase.

## Deploy Guard Server
For detailed instructions follow the guide [here](/docs/setup/install.md). To generate correct installer YAMLs for Guard server, pass the flag `--firebase.sa-json-file`.

```console
# generate Kubernetes YAMLs for deploying guard server
$ guard get installer \
  --namespace=<namespace> \
  --auth-providers=firebase \
  --firebase.sa-json-file==<path-json-key-file> > installer.yaml

$ kubectl apply -f installer.yaml
```

## Setup Kube API server to use Guard with Firebase
```console
$ guard get webhook-config  \
  -o firebase \
  --addr=guard.<namespace>.svc:443 > server-config.yaml

kubectl create secret generic api-server-auth --from-file=server-config.yaml
# mount the secret in api server pod
```

![firebase-webhook-flow](/docs/images/firebase-webhook-flow.png)
```json
{
  "apiVersion": "authentication.k8s.io/v1",
  "kind": "TokenReview",
  "status": {
    "authenticated": true,
    "user": {
      "username": "john@mycompany.com",
      "uid": "<firebase-uid>",
    }
  }
}
```
Guard uses the token found in `TokenReview` request object to read user's email address and UID in Firebase. In the `TokenReview` response, `status.user.username` is set to user's email, `status.user.uid` is set to user's UID.
