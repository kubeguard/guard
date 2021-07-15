---
title: Guard Get Installer
menu:
  product_guard_{{ .version }}:
    identifier: guard-get-installer
    name: Guard Get Installer
    parent: reference
product_name: guard
menu_name: product_guard_{{ .version }}
section_menu_id: reference
---
## guard get installer

Prints Kubernetes objects for deploying guard server

```
guard get installer [flags]
```

### Options

```
      --addr string                                Address (host:port) of guard server. (default "10.96.10.96:443")
      --auth-providers strings                     name of providers for which guard will provide authentication service (required), supported providers : Azure/Github/Gitlab/Google/Ldap/Token-Auth
      --authz-providers strings                    name of providers for which guard will provide authorization service, supported providers : Azure
      --azure.aks-authz-token-url string           url to call for AKS Authz flow
      --azure.aks-token-url string                 url to call for AKS OBO flow
      --azure.allow-nonres-discovery-path-access   allow access on Non Resource paths required for discovery, setting it false will require explicit non resource path role assignment for all users in Azure RBAC (default true)
      --azure.arm-call-limit int                   No of calls before which webhook switch to new ARM instance to avoid throttling (default 2000)
      --azure.auth-mode string                     auth mode to call graph api, valid value is either aks, obo, or client-credential (default "client-credential")
      --azure.authz-mode string                    authz mode to call RBAC api, valid value is either aks or arc
      --azure.client-id string                     MS Graph application client ID to use
      --azure.client-secret string                 MS Graph application client secret to use
      --azure.environment string                   Azure cloud environment
      --azure.graph-call-on-overage-claim          set to true to resolve group membership only when overage claim is present. setting to false will always call graph api to resolve group membership
      --azure.resource-id string                   azure cluster resource id (//subscription/<subName>/resourcegroups/<RGname>/providers/Microsoft.ContainerService/managedClusters/<clustername> for AKS or //subscription/<subName>/resourcegroups/<RGname>/providers/Microsoft.Kubernetes/connectedClusters/<clustername> for arc) to be used as scope for RBAC check
      --azure.skip-authz-check strings             name of usernames/email for which authz check will be skipped
      --azure.skip-authz-for-non-aad-users         skip authz for non AAD users (default true)
      --azure.tenant-id string                     MS Graph application tenant id to use
      --azure.use-group-uid                        Use group UID for authentication instead of group display name (default true)
      --azure.verify-clientID                      set to true to validate token's audience claim matches clientID
      --github.base-url string                     Base url for enterprise, keep empty to use default github base url
      --gitlab.base-url string                     Base url for GitLab, including the API path, keep empty to use default gitlab base url.
      --gitlab.use-group-id                        Use group ID for authentication instead of group full path
      --google.admin-email string                  Email of G Suite administrator
      --google.sa-json-file string                 Path to Google service account json file
  -h, --help                                       help for installer
      --image-pull-secret string                   Name of image pull secret
      --ldap.auth-choice AuthChoice                LDAP user authentication mechanisms Simple/Kerberos(via GSSAPI) (default Simple)
      --ldap.bind-dn string                        The connector uses this DN in credentials to search for users and groups. Not required if the LDAP server provides access for anonymous auth.
      --ldap.bind-password string                  The connector uses this password in credentials to search for users and groups. Not required if the LDAP server provides access for anonymous auth.
      --ldap.ca-cert-file string                   ca cert file that used for self signed server certificate
      --ldap.group-member-attribute string         Ldap group member attribute (default "member")
      --ldap.group-name-attribute string           Ldap group name attribute (default "cn")
      --ldap.group-search-dn string                BaseDN to start the search group
      --ldap.group-search-filter string            Filter to apply when searching the groups that user is member of (default "(objectClass=groupOfNames)")
      --ldap.is-secure-ldap                        Secure LDAP (LDAPS)
      --ldap.keytab-file string                    path to the keytab file, it's contain LDAP service principal keys
      --ldap.server-address string                 Host or IP of the LDAP server
      --ldap.server-port string                    LDAP server port (default "389")
      --ldap.service-account string                service account name
      --ldap.skip-tls-verification                 Skip LDAP server TLS verification, default : false
      --ldap.start-tls                             Start tls connection
      --ldap.user-attribute string                 Ldap username attribute (default "uid")
      --ldap.user-search-dn string                 BaseDN to start the search user
      --ldap.user-search-filter string             Filter to apply when searching user (default "(objectClass=person)")
  -n, --namespace string                           Name of Kubernetes namespace used to run guard server. (default "kube-system")
      --pki-dir string                             Path to directory where pki files are stored. (default "/home/tamal/.guard")
      --private-registry string                    Private Docker registry (default "appscode")
      --proxy-cert string                          Path to the certificate file for proxy
      --proxy-http string                          Http proxy URL to be used
      --proxy-https string                         Https proxy URL to be used
      --proxy-skip-range string                    List of URLs/CIDRs for which proxy should not to be used
      --run-on-master                              If true, runs Guard server on master instances (default true)
      --token-auth-file string                     To enable static token authentication
      --v string                                   Log level for V logs (default "3")
```

### SEE ALSO

* [guard get](/docs/reference/guard_get.md)	 - Get PKI

