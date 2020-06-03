---
title: Azure Authorizer | Guard
description: Authorize into Kubernetes using Azure
menu:
  product_guard_{{ .version }}:
    identifier: azure-authorizer
    parent: authorizer-guides
    name: Azure
    weight: 10
product_name: guard
menu_name: product_guard_{{ .version }}
section_menu_id: guides
---

# Azure Authorizer

Guard installation guide can be found [here](/docs/setup/install.md). To use Azure, create a client cert with `Organization` set to `Azure`.For Azure `CommonName` is optional. To ease this process, use the Guard cli to issue a client cert/key pair.

```console
$ guard init client [CommonName] -o Azure
```

Azure authenticator guide can be found [here](/docs/guides/authenticator/azure.md).


## ARC mode

Guard can be configured with arc mode which uses service principal (azure.client-id configured for authentication) with read access on subscription of kubernetes cluster.

### Deploy guard server

To generate installer YAMLs for guard server you can use the following command.

```console
# generate Kubernetes YAMLs for deploying guard server
$ guard get installer \
    -- all authentication options as per Azure authenticatoin guide
    --authz-providers=azure \
    --azure.authz-mode=arc \
    --azure.resource-id=<arc k8s cluster arm resource id> \
    --azure.skip-authz-check=<comma separated list of user email ids for which Azure RBAC will be skipped>
    --azure.authz-resolve-group-memberships=true \
    --azure.skip-authz-for-non-aad-users=true \
    --azure.allow-nonres-discovery-path-access=true \
    > installer.yaml

$ kubectl apply -f installer.yaml
```

> **Note**
> Azure authorization can be enabled only with Azure authentication.
> Create single installer.yaml with both authentication and authorization options together.
> ARC mode can be enabled with client credential mode or On-Behalf-Of (OBO) mode.
> Keep azure.skip-authz-for-non-aad-users=true for certificate users (non AAD users) to work with Azure authorization. You are required to set separate Kubernetes RBAC authorizer for certificate users.

## Further Reading:
- https://docs.microsoft.com/en-us/azure/role-based-access-control/overview
- https://docs.microsoft.com/en-us/azure/role-based-access-control/best-practices
- https://aka.ms/AzureArcK8sDocs
