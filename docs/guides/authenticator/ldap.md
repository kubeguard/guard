---
title: LDAP Authenticator | Guard
description: Authenticate into Kubernetes using LDAP
menu:
  product_guard_0.1.0-rc.5:
    identifier: ldap-authenticator
    parent: authenticator-guides
    name: LDAP
    weight: 10
product_name: guard
menu_name: product_guard_0.1.0-rc.5
section_menu_id: guides
---

# LDAP Authenticator

TO use LDAP,

1.  Create a client cert with `Organization` set to `Ldap`.For LDAP `COMMON_NAME` is optional. To ease this process, use the Guard cli to issue a client cert/key pair.
    
    ```console
    # If COMMON_NAME is not provided, then default COMMON_NAME `ldap` is used
    $ guard init client [COMMON_NAME] -o Ldap
    ```

2.  Send following additional flags to guard server.
    ```console
    --ldap.server-address=[SERVER_ADDRESS]
    
    # If the port is not supplied, then default port `389` is used
    --ldap.server-port=[PORT]
    
    # To start tls connection
    --ldap.start-tls
    
    # For secure LDAP (LDAPS)
    --ldap.is-secure-ldap
    
    # The connector uses bind DN and password as credential to search for users and groups.
    # Not required if the LDAP server provides access for anonymous auth.
    --ldap.bind-dn=[BIND_DN]
    --ldap.bind-password=[BIND_PASSWORD]
    
    # BaseDN to start the user search
    --ldap.user-search-dn=[USER_SEARCH_DN]
    
    # Filter to apply when searching user
    # If the filter is not supplied, then default filter `(objectClass=person)` is used
    --ldap.user-search-filter=[USER_SEARCH_FILTER]
    
    # LDAP username attribute
    # If the attribute is not supplied, then default attribute `uid` is used
    --ldap.user-attribute=[USER_ATTRIBUTE]
    
    # BaseDN to start the group search
    --ldap.group-search-dn=[GROUP_SEARCH_DN]
    
    # Filter to apply when searching the group
    # If the filter is not supplied, then default filter `(objectClass=groupOfNames)` is used
    --ldap.group-search-filter=[GROUP_SEARCH_FILTER]
    
    # LDAP group name attribute
    # If the attribute is not supplied, then default attribute `cn` is used
    --ldap.group-name-attribute=[GROUP_NAME_ATTRIBUTE]
    
    # LDAP group member attribute
    # If the attribute is not supplied, then default attribute `member` is used
    --ldap.group-member-attribute=[GROUP_MEMBER_ATTRIBUTE]  
    
    # To skip LDAP server TLS verification, provide this flag
    --ldap.skip-tls-verification
    
    # Ca cert file that used for self signed LDAP server certificate
    --ldap.ca-cert-file
        
    ```
    
    Or you can use following command to create YAMLs for this setup.
     ```console
     # generate Kubernetes YAMLs for deploying guard server
     $  guard get installer \
            --ldap.server-address=[SERVER_ADDRESS] \
            --ldap.server-port=[PORT] \
            --ldap.bind-dn=[BIND_DN] \
            --ldap.bind-password=[BIND_PASSWORD] \
            --ldap.user-search-dn=[USER_SEARCH_DN] \
            --ldap.user-search-filter=[USER_SEARCH_FILTER] \
            --ldap.user-attribute=[USER_ATTRIBUTE] \
            --ldap.group-search-dn=[GROUP_SEARCH_DN] \
            --ldap.group-search-filter=[GROUP_SEARCH_FILTER] \
            --ldap.group-name-attribute=[GROUP_NAME_ATTRIBUTE] \
            --ldap.group-member-attribute=[GROUP_MEMBER_ATTRIBUTE] \
            --ldap.skip-tls-verification=[true/false] \
            --ldap.start-tls=[true/false]\
            --ldap.is-secure-ldap=[true/false]\
            --ldap.ca-cert-file=[PATH_TO_THE_CA_CERT_FILE]
            > installer.yaml

     $ kubectl apply -f installer.yaml
     
     ```
     
     > **Note:** User search filter is applied in this form : `(&[USER_SEARCH_FILTER]([USER_ATTRIBUTE]=[USERNAME]))` and group search filter is applied in this form : `(&[GROUP_SEARCH_FILTER]([GROUP_MEMBER_ATTRIBUTE]=[USER_DN]))`
     
### Configure Kubectl
```console
kubectl config set-credentials [USERNAME] --token=[TOKEN]
```

Or You can add user in ~/.kube/config file
```yaml
...
users:
- name: [USERNAME]
  user:
    token: [TOKEN]
```

### How to generate token

For LDAP, user is authenticated using user DN and password. User DN is collected using `[USERNAME]` and `[USER_ATTRIBUTE]`.
So, use base64 encoded string of `[USERNAME]:[PASSWORD]` as token.

|username |password |username:password     |token
|---------|---------|----------------------|------------------
|user12   |12345    |user12:12345          |dXNlcjEyOjEyMzQ1
