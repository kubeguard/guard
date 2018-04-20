---
title: Azure Authenticator | Guard
description: Authenticate into Kubernetes using Azure
menu:
  product_guard_0.1.0:
    identifier: azure-authenticator
    parent: authenticator-guides
    name: Azure
    weight: 25
product_name: guard
menu_name: product_guard_0.1.0
section_menu_id: guides
---

# Azure Authenticator

Guard installation guide can be found [here](/docs/setup/install.md). To use Azure, create a client cert with `Organization` set to `Azure`.For Azure `CommonName` is optional. To ease this process, use the Guard cli to issue a client cert/key pair.

```console
$ guard init client [CommonName] -o Azure
```

### Deploy guard server

To generate installer YAMLs for guard server you can use the following command.

```console
# generate Kubernetes YAMLs for deploying guard server
$ guard get installer \
    --auth-providers = "azure" \
    --azure.client-id=<application_id> \
    --azure.tenant-id=<tenant_id> \
    > installer.yaml

$ kubectl apply -f installer.yaml
```
> **Note:** guard take `<application_secret>` from environment variable **AZURE_CLIENT_SECRET**

Procedure to find `<application_id>`, `<application_secret>` are given below. Replace the `<tenant_id>` with your azure tenant id.

### Configure Azure Active Directory App

[![Configuring Microsoft Azure auth provider for Kubernetes using Guard](https://img.youtube.com/vi/n2kKwAFYuiM/0.jpg)](https://www.youtube-nocookie.com/embed/n2kKwAFYuiM)

Configuring Azure AD as a auth provider requires an initial setup by `Global Administrator` of such AD. This involves a complex multi-step process. Please see the video above to setup your Azure AD.

1.  Sign in to the [Azure portal](https://portal.azure.com/). Please make sure that you are a `Global Administrator` of your Azure AD. If not, please contact your Azure AD administrator to perform these steps.

    ![aad-dir-role](/docs/images/azure/dir-role.png)

2.  Create an Azure Active Directory Web App / API application

    ![create-app-registration](/docs/images/azure/create-app-registration.png)

3.  Use the **Application ID** as `<application_id>`

    ![application-id](/docs/images/azure/application-id.png)

4.  Click on the **Settings**, click on the **key** , generate a key and use this key as `<application_secret>`

    ![secret-key](/docs/images/azure/secret-key.png)

5.  Add **Microsoft Graph** api with _application permission_ `Read directory data` and _delegated permission_ `Read directory data` and `Sign in and read user profile`.

    ![add-api](/docs/images/azure/add-api.png)
    ![add-api-2](/docs/images/azure/add-api-2.png)

6. Now grant grant the permission from step 5 for all account to this application by clicking the `Grant Permissions` button. Afterwards, check the permissions for this application to confirm that grant operation was successful.

    ![guard-grant-perm](/docs/images/azure/guard-grant-perm.png)
    ![guard-permissions](/docs/images/azure/guard-permissions.png)

7.  Create a second Azure Active Directory native application

    ![create-native-app](/docs/images/azure/create-native-app.png)

8.  Use the **Application ID** of this native app as `<client_id>`

    ![client-id](/docs/images/azure/client-id.png)

9.  Add application created at step 2 with permission `Access <Application_Name_Created_At_Step_2>`

    ![add-guard-app](/docs/images/azure/add-guard-api.png)

## Configure kubectl

```console
kubectl config set-credentials <user_name> --auth-provider=azure \
  --auth-provider-arg=environment=AzurePublicCloud \
  --auth-provider-arg=client-id=<client_id> \
  --auth-provider-arg=tenant-id=<tenant_id> \
  --auth-provider-arg=apiserver-id=<application_id>
```

Procedure to find `<application_id>`, `<application_secret>` and `<client_id>` are given above. Replace the <user_name> and <tenant_id> with your azure username and tenant id.

Or You can add user in `.kube/config` file

```yaml
...
users:
- name: <user_name>
  user:
    auth-provider:
      config:
        apiserver-id: <application_id>
        client-id: <client_id>
        tenant-id: <tenant_id>
        environment: AzurePublicCloud
      name: azure
```

The access token is acquired when first `kubectl` command is executed

   ```
   $ kubectl get pods --user <user_name>

   To sign in, use a web browser to open the page https://aka.ms/devicelogin and enter the code DEC7D48GA to authenticate.
   ```

After signing in a web browser, the token is stored in the configuration, and it will be reused when executing next commands.

## Further Reading:
- https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-integrating-applications
- https://github.com/kubernetes/client-go/blob/master/plugin/pkg/client/auth/azure/README.md
