---
title: LDAP Authenticator | Guard
description: Authenticate into Kubernetes using LDAP
menu:
  product_guard_0.3.0:
    identifier: ldap-authenticator
    parent: authenticator-guides
    name: LDAP
    weight: 35
product_name: guard
menu_name: product_guard_0.3.0
section_menu_id: guides
---

# LDAP Authenticator

Guard installation guide can be found [here](/docs/setup/install.md). To use LDAP, create a client cert with `Organization` set to `Ldap`. For LDAP `CommonName` is optional. To ease this process, use the Guard cli to issue a client cert/key pair.

```console
# If CommonName is not provided, then default CommonName `ldap` is used
$ guard init client [CommonName] -o Ldap
```

### Deploy Guard Server

To generate installer YAMLs for guard server you can use the following command.

```console
# generate Kubernetes YAMLs for deploying guard server
$  guard get installer \
        --auth-providers="ldap" \
        --ldap.server-address=<server_address> \
        --ldap.server-port=<server_port> \
        --ldap.bind-dn=<bind_dn> \
        --ldap.bind-password=<bind_password> \
        --ldap.user-search-dn=<user_search_dn> \
        --ldap.user-search-filter=<user_search_filter> \
        --ldap.user-attribute=<user_attribute> \
        --ldap.group-search-dn=<group_search_dn> \
        --ldap.group-search-filter=<group_search_filter> \
        --ldap.group-name-attribute=<group_name_attribute> \
        --ldap.group-member-attribute=<group_member_attribute> \
        --ldap.skip-tls-verification=<true/false> \
        --ldap.start-tls=<true/false>\
        --ldap.is-secure-ldap=<true/false>\
        --ldap.ca-cert-file=<path_to_the_ca_cert_file>
        --ldap.auth-choice=<Simple/Kerberos>
        > installer.yaml

$ kubectl apply -f installer.yaml
```

Additional flags for LDAP:

```console
--ldap.server-address=<server_address>

# If the port is not supplied, then default port `389` is used
--ldap.server-port=<server_port>

# To start tls connection
--ldap.start-tls

# For secure LDAP (LDAPS)
--ldap.is-secure-ldap

# The connector uses bind DN and password as credential to search for users and groups.
# Not required if the LDAP server provides access for anonymous auth.
--ldap.bind-dn=<bind_dn>
--ldap.bind-password=<bind_password>

# BaseDN to start the user search
--ldap.user-search-dn=<user_search_dn>

# Filter to apply when searching user
# If the filter is not supplied, then default filter `(objectClass=person)` is used
--ldap.user-search-filter=<user_search_filter>

# LDAP username attribute
# If the attribute is not supplied, then default attribute `uid` is used
--ldap.user-attribute=<user_attribute>

# BaseDN to start the group search
--ldap.group-search-dn=<group_search_dn>

# Filter to apply when searching the group
# If the filter is not supplied, then default filter `(objectClass=groupOfNames)` is used
--ldap.group-search-filter=<group_search_filter>

# LDAP group name attribute
# If the attribute is not supplied, then default attribute `cn` is used
--ldap.group-name-attribute=<group_name_attribute>

# LDAP group member attribute
# If the attribute is not supplied, then default attribute `member` is used
--ldap.group-member-attribute=<group_member_attribute>

# To skip LDAP server TLS verification, provide this flag
--ldap.skip-tls-verification=<true/false>

# Ca cert file that used for self signed LDAP server certificate
--ldap.ca-cert-file=<path_to_the_ca_cert_file>

# LDAP user authentication mechanism, Simple or Kerberos
--ldap.auth-choice=<0/1>

# path to the keytab file, it's contain LDAP service principal keys
--ldap.keytab-file=<path_to_the_keytab_file>

# service account name, if empty then service principal name from keytab file will be used
--ldap.service-account=<service_account_name>
```

Environment variable needed to set for LDAP:

```console
# The connector uses bind DN and password as credential to search for users and groups.
# Not required if the LDAP server provides access for anonymous auth.
$ export LDAP_BIND_DN=<bind_dn>
$ export LDAP_BIND_PASSWORD=<bind_password>
```

> **Note:** User search filter is applied in this form : `(&<user_search_filter>(<user_attribute>=<user_name>))` and group search filter is applied in this form : `(&<group_search_filter>(<group_member_attribute>=<user_dn>))`

### Issue Token

**Simple authentication:** Use following guard command to get token:

```console
$ guard get token \
    -o ldap \
    --ldap.auth-choice=Simple \
    --ldap.username=<user_name> \
    --ldap.password=<password>

I0330 11:37:12.375526   24687 logs.go:19] FLAG: --alsologtostderr="false"
I0330 11:37:12.376448   24687 logs.go:19] FLAG: --analytics="true"
I0330 11:37:12.376465   24687 logs.go:19] FLAG: --help="false"
I0330 11:37:12.376476   24687 logs.go:19] FLAG: --ldap.auth-choice="Simple"
I0330 11:37:12.376497   24687 logs.go:19] FLAG: --ldap.disable-pa-fx-fast="true"
I0330 11:37:12.376518   24687 logs.go:19] FLAG: --ldap.krb5-config="/etc/krb5.conf"
I0330 11:37:12.376534   24687 logs.go:19] FLAG: --ldap.password=<password>
I0330 11:37:12.376552   24687 logs.go:19] FLAG: --ldap.realm=""
I0330 11:37:12.376582   24687 logs.go:19] FLAG: --ldap.spn=""
I0330 11:37:12.376594   24687 logs.go:19] FLAG: --ldap.username=<user_name>
I0330 11:37:12.376609   24687 logs.go:19] FLAG: --log_backtrace_at=":0"
I0330 11:37:12.376619   24687 logs.go:19] FLAG: --log_dir=""
I0330 11:37:12.376629   24687 logs.go:19] FLAG: --logtostderr="false"
I0330 11:37:12.376638   24687 logs.go:19] FLAG: --organization="ldap"
I0330 11:37:12.376647   24687 logs.go:19] FLAG: --stderrthreshold="0"
I0330 11:37:12.376656   24687 logs.go:19] FLAG: --v="0"
I0330 11:37:12.376666   24687 logs.go:19] FLAG: --vmodule=""
Current Kubeconfig is backed up as /home/ac/.kube/config.bak.2018-03-30T11-37.
Configuration has been written to /home/ac/.kube/config

$ cat ~/.kube/config
...
users:
- name: <user_name>
  user:
    token: <token>

$ kubectl get pods --all-namespaces --user <user_name>
NAMESPACE     NAME                               READY     STATUS    RESTARTS   AGE
kube-system   etcd-minikube                      1/1       Running   0          7h
kube-system   kube-addon-manager-minikube        1/1       Running   0          7h
kube-system   kube-apiserver-minikube            1/1       Running   1          7h
kube-system   kube-controller-manager-minikube   1/1       Running   0          7h
kube-system   kube-dns-6f4fd4bdf-f7csh           3/3       Running   0          7h
```

**Kerberos authentication:** Use following guard command to get token:

```console
$ guard get token \
    -o ldap \
    --ldap.auth-choice=Kerberos \
    --ldap.username=<user_name> \
    --ldap.password=<password> \
    --ldap.krb5-config=<path_to_the_krb5_config_file> \
    --ldap.realm=<realm> \
    --ldap.disable-pa-fx-fast=<true/false> \
    --ldap.spn=<service_principle_name>

I0330 11:37:12.375526   24687 logs.go:19] FLAG: --alsologtostderr="false"
I0330 11:37:12.376448   24687 logs.go:19] FLAG: --analytics="true"
I0330 11:37:12.376465   24687 logs.go:19] FLAG: --help="false"
I0330 11:37:12.376476   24687 logs.go:19] FLAG: --ldap.auth-choice="Kerberos"
I0330 11:37:12.376497   24687 logs.go:19] FLAG: --ldap.disable-pa-fx-fast=<true/false>
I0330 11:37:12.376518   24687 logs.go:19] FLAG: --ldap.krb5-config=<path_to_the_krb5_config_file>
I0330 11:37:12.376534   24687 logs.go:19] FLAG: --ldap.password=<password>
I0330 11:37:12.376552   24687 logs.go:19] FLAG: --ldap.realm=<realm>
I0330 11:37:12.376582   24687 logs.go:19] FLAG: --ldap.spn=<service_principle_name>
I0330 11:37:12.376594   24687 logs.go:19] FLAG: --ldap.username=<user_name>
I0330 11:37:12.376609   24687 logs.go:19] FLAG: --log_backtrace_at=":0"
I0330 11:37:12.376619   24687 logs.go:19] FLAG: --log_dir=""
I0330 11:37:12.376629   24687 logs.go:19] FLAG: --logtostderr="false"
I0330 11:37:12.376638   24687 logs.go:19] FLAG: --organization="ldap"
I0330 11:37:12.376647   24687 logs.go:19] FLAG: --stderrthreshold="0"
I0330 11:37:12.376656   24687 logs.go:19] FLAG: --v="0"
I0330 11:37:12.376666   24687 logs.go:19] FLAG: --vmodule=""
Current Kubeconfig is backed up as /home/ac/.kube/config.bak.2018-03-30T11-37.
Configuration has been written to /home/ac/.kube/config

$ cat ~/.kube/config
...
users:
- name: <user_name>
  user:
    token: <token>

$ kubectl get pods --all-namespaces --user <user_name>
NAMESPACE     NAME                               READY     STATUS    RESTARTS   AGE
kube-system   etcd-minikube                      1/1       Running   0          7h
kube-system   kube-addon-manager-minikube        1/1       Running   0          7h
kube-system   kube-apiserver-minikube            1/1       Running   1          7h
kube-system   kube-controller-manager-minikube   1/1       Running   0          7h
kube-system   kube-dns-6f4fd4bdf-f7csh           3/3       Running   0          7h
```

**Flag** details for get token:
```console
# LDAP user authentication mechanism
#   - 0 for simple authentication
#   - 1 for kerberos(via GSSAPI)
--ldap.auth-choice=<Simple/Kerberos>

# Username
--ldap.username=<user_name>

# Password
--ldap.password=<password>

# Path to the kerberos configuration file (default "/etc/krb5.conf")
-ldap.krb5-config=<path_to_the_krb5_config_file>

# Realm, set the realm to empty string to use the default realm from config
--ldap.realm=<realm>

# Service principal name
--ldap.spn=<service_principle_name>

# Disable PA-FX-Fast, Active Directory does not commonly support FAST negotiation
# so you will need to disable this on the client (default true)
--ldap.disable-pa-fx-fast=<true/false>

```

### How guard generate token for simple authentication

For LDAP, user is authenticated using user DN and password. User DN is collected using `<user_name>` and `<user_attribute>`.
So, guard use base64 encoded string of `<user_name>:<password>` as token.

|username |password |username:password     |token
|---------|---------|----------------------|------------------
|user12   |12345    |user12:12345          |dXNlcjEyOjEyMzQ1
