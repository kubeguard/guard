---
title: Guard Run
menu:
  product_guard_0.2.1:
    identifier: guard-run
    name: Guard Run
    parent: reference
product_name: guard
menu_name: product_guard_0.2.1
section_menu_id: reference
---
## guard run

Run server

### Synopsis

Run server

```
guard run [flags]
```

### Options

```
      --auth-providers strings               name of providers for which guard will provide authentication service (required), supported providers : Azure/Github/Gitlab/Google/Ldap/Token-Auth
      --azure.client-id string               MS Graph application client ID to use
      --azure.client-secret string           MS Graph application client secret to use
      --azure.tenant-id string               MS Graph application tenant id to use
      --azure.use-group-uid                  Use group UID for authentication instead of group display name (default true)
      --clock-check-interval duration        Interval between checking time against NTP servers, set to 0 to disable checks (default 10m0s)
      --github.base-url string               Base url for enterprise, keep empty to use default github base url
      --gitlab.base-url string               Base url for GitLab, including the API path, keep empty to use default gitlab base url.
      --google.admin-email string            Email of G Suite administrator
      --google.sa-json-file string           Path to Google service account json file
  -h, --help                                 help for run
      --ldap.auth-choice AuthChoice          LDAP user authentication mechanisms Simple/Kerberos(via GSSAPI) (default Simple)
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
      --max-clock-skew duration              Max acceptable clock skew for server clock (default 2m0s)
      --ntp-server string                    Address of NTP serer used to check clock skew (default "0.pool.ntp.org")
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

* [guard](/docs/reference/guard.md)	 - Guard by AppsCode - Kubernetes Authentication WebHook Server

