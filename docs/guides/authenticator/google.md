---
title: Google Authenticator | Guard
description: Authenticate into Kubernetes using Google
menu:
  product_guard_0.1.0-rc.5:
    identifier: google-authenticator
    parent: authenticator-guides
    name: Google
    weight: 15
product_name: guard
menu_name: product_guard_0.1.0-rc.5
section_menu_id: guides
---

# Google Authenticator
To use Google, you need a client cert with `CommonName` set to Google Apps (now G Suite) domain and `Organization` set to `Google`. To ease this process, use the Guard cli to issue a client cert/key pair.
```console
$ guard init client {domain-name} -o Google
```

## G Suite Domain-Wide Delegation of Authority
Guard server needs to determine the list of groups for any user in a G suite domain. This requires the domain administrator to grant Guard server with domain-wide access to its users' data — this is referred as domain-wide delegation of authority. The following procedure has been adapted from official documentation found [here](https://developers.google.com/admin-sdk/directory/v1/guides/delegation).

### Create the service account and its credentials
G Suite domain administrator needs to create a service account and its credentials. During this procedure you need to gather information that will be later passed to Guard server installer.

- Open the [Service accounts page](https://console.developers.google.com/permissions/serviceaccounts). If prompted, select a project.
- Click Create service account.
- In the Create service account window, type a name for the service account, and select Furnish a new private key and Enable Google Apps Domain-wide Delegation. Then click Create.

You should now have gathered your service account's Private Key file, Client ID and email address. You are ready to delegate domain-wide authority to your service account.

### Delegate domain-wide authority to your service account
The service account that you created needs to be granted access to the G Suite domain’s user data that Guard server needs to access. The following tasks have to be performed by an administrator of the G Suite domain:

- Go to your G Suite domain’s [Admin console page for managing API client access](https://admin.google.com/ManageOauthClients).
- In the Client name field enter the service account's Client ID.
- In the One or More API Scopes field enter _https://www.googleapis.com/auth/admin.directory.group.readonly_
- Click the Authorize button.

Your service account now has domain-wide access to the Google Admin SDK Directory API for all the users of your domain. Only users with access to the Admin APIs can access the Admin SDK Directory API, therefore your service account needs to impersonate one of those users to access the Admin SDK Directory API.

## Deploy Guard Server
For detailed instructions follow the guide [here](/docs/setup/install.md). To generate correct installer YAMLs for Guard server, pass the flags `--google.admin-email` and `--google.sa-json-file`.

```console
# generate Kubernetes YAMLs for deploying guard server
$ guard get installer \
  --google.admin-email=<email-of-a-g-suite-admin> \
  --google.sa-json-file=<path-json-key-file> > installer.yaml
$ kubectl apply -f installer.yaml
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
