---
title: Azure Authenticator | Guard
description: Authenticate into Kubernetes using Azure
menu:
  product_guard_0.1.0-rc.5:
    identifier: azure-authenticator
    parent: authenticator-guides
    name: Azure
    weight: 10
product_name: guard
menu_name: product_guard_0.1.0-rc.5
section_menu_id: guides
---

# Azure Authenticator

TO use Azure,

1.  Create a client cert with `Organization` set to `Azure`.For Azure `CommonName` is optional. To ease this process, use the Guard cli to issue a client cert/key pair.
    
    ```console
    $ guard init client {common-name} -o Azure
    ```

2.  Send additional `--azure-client-id`,`--azure-client-secret` and `--azure-tenant-id` flags to guard server. You can use following command to create YAMLs for this setup.
     ```console
     # generate Kubernetes YAMLs for deploying guard server
     $ guard get installer --azure-client-id=[APPLICATION_ID] --azure-client-secret=[APPLICATION_SECRET] --azure-tenant-id=[TENANT_ID] > installer.yaml
     $ kubectl apply -f installer.yaml

     ```
     Procedure to find `APPLICATION_ID`, `APPLICATION_SECRET` are given below. Replace the TENANT_ID with your azure tenant id.

### Configure Azure Active Directory App

1.  Sign in to the [Azure portal](https://portal.azure.com/)

2.  Create an Azure Active Directory Web App / API application

    ![create-app-registration](/docs/images/azure/create-app-registration.png)
    
3.  Use the **Application ID** as `APPLICATION_ID`

    ![application-id](/docs/images/azure/application-id.png)

4.  Click on the **Settings**, click on the **key** , generate a key and use this key as `APPLICATION_SECRET`

    ![secret-key](/docs/images/azure/secret-key.png)
    
5.  Click on the **Manifest** , set `groupMembershipClaims` to `All` and **save** the mainfest

    ![update-manifest](/docs/images/azure/update-manifest.png)
    
6.  Add **Microsoft graph** api with permission `Read directory data` and `Sign in and read user profile`.
    
    ![add-api](/docs/images/azure/add-api.png)
    
7.  Create a second Azure Active Directory native application

    ![create-native-app](/docs/images/azure/create-native-app.png)
    
8.  Use the **Application ID** of this native app as `CLIENT_ID`

    ![client-id](/docs/images/azure/client-id.png)

9.  Add application created at step 2 with permission `Access [Application_Name_Created_At_Step_2]`
    
    ![add-guard-app](/docs/images/azure/add-guard-api.png)

## Configure kubectl

```console
kubectl config set-credentials "USER_NAME" --auth-provider=azure \
  --auth-provider-arg=environment=AzurePublicCloud \
  --auth-provider-arg=client-id=CLIENT_ID \
  --auth-provider-arg=tenant-id=TENANT_ID \
  --auth-provider-arg=apiserver-id=APPLICATION_ID
```

Procedure to find `APPLICATION_ID`, `APPLICATION_SECRET` and `CLIENT_ID` are given above. Replace the USER_NAME and TENANT_ID with your azure username and tenant id.

Or You can add user in `.kube/config` file

```yaml
...
users:
- name: USER_NAME
  user:
    auth-provider:
      config:
        apiserver-id: APPLICATION_ID
        client-id: CLIENT_ID
        tenant-id: TENANT_ID
        environment: AzurePublicCloud
      name: azure
```

The access token is acquired when first `kubectl` command is executed

   ```
   kubectl get pods

   To sign in, use a web browser to open the page https://aka.ms/devicelogin and enter the code DEC7D48GA to authenticate.
   ```

After signing in a web browser, the token is stored in the configuration, and it will be reused when executing next commands.
