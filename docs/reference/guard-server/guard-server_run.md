---
title: Run
menu:
  product_guard_0.1.0-rc.5:
    identifier: guard-server-run
    name: Run
    parent: guard-server
product_name: guard
section_menu_id: reference
menu_name: product_guard_0.1.0-rc.5
---
## guard-server run

Run server

### Synopsis

Run server

```
guard-server run [flags]
```

### Options

```
      --azure.client-id string               MS Graph application client ID to use
      --azure.client-secret string           MS Graph application client secret to use
      --azure.tenant-id string               MS Graph application tenant id to use
      --clock-check-interval duration        Interval between checking time against NTP servers (default 5m0s)
      --google.admin-email string            Email of G Suite administrator
      --google.sa-json-file string           Path to Google service account json file
  -h, --help                                 help for run
      --ldap.auth-choice int                 LDAP user authentication mechanism, 0 for simple authentication, 1 for kerberos(via GSSAPI)
      --ldap.bind-dn string                  The connector uses this DN in credentials to search for users and groups. Not required if the LDAP server provides access for anonymous auth.
      --ldap.bind-password string            The connector uses this password in credentials to search for users and groups. Not required if the LDAP server provides access for anonymous auth.
      --ldap.ca-cert-file string             ca cert file that used for self signed server certificate
      --ldap.group-member-attribute string   Ldap group member attribute (default "member")
      --ldap.group-name-attribute string     Ldap group name attribute (default "cn")
      --ldap.group-search-dn string          BaseDN to start the search group
      --ldap.group-search-filter string      Filter to apply when searching the groups that user is member of (default "(objectClass=groupOfNames)")
      --ldap.is-secure-ldap                  Secure LDAP (LDAPS)
      --ldap.keytab-file string              path to the keytab file, it's contain LDAP service principal keys
      --ldap.server-address string           Host or IP of the LDAP server
      --ldap.server-port string              LDAP server port (default "389")
      --ldap.service-account string          service account name
      --ldap.skip-tls-verification           Skip LDAP server TLS verification, default : false
      --ldap.start-tls                       Start tls connection
      --ldap.user-attribute string           Ldap username attribute (default "uid")
      --ldap.user-search-dn string           BaseDN to start the search user
      --ldap.user-search-filter string       Filter to apply when searching user (default "(objectClass=person)")
      --max-clock-skew duration              Max acceptable clock skew for server clock (default 5s)
      --secure-addr string                   host:port used to serve secure apis (default ":8443")
      --tls-ca-file string                   File containing CA certificate
      --tls-cert-file string                 File container server TLS certificate
      --tls-private-key-file string          File containing server TLS private key
      --token-auth-file string               To enable static token authentication
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --analytics                        Send analytical events to Google Guard (default true)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [guard-server](/docs/reference/guard-server/guard-server.md)	 - Guard server by AppsCode - Kubernetes Authentication WebHook Server

