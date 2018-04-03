---
title: Guard Get Token
menu:
  product_guard_0.1.0-rc.5:
    identifier: guard-get-token
    name: Guard Get Token
    parent: reference
product_name: guard
menu_name: product_guard_0.1.0-rc.5
section_menu_id: reference
---
## guard get token

Get tokens for Appscode/Azure/Github/Gitlab/Google/Ldap/Token-Auth

### Synopsis

Get tokens for Appscode/Azure/Github/Gitlab/Google/Ldap/Token-Auth

```
guard get token [flags]
```

### Options

```
  -h, --help                      help for token
      --ldap.auth-choice int      LDAP user authentication mechanism, 0 for simple authentication, 1 for kerberos(via GSSAPI)
      --ldap.disable-pa-fx-fast   Disable PA-FX-Fast, Active Directory does not commonly support FAST negotiation so you will need to disable this on the client (default true)
      --ldap.krb5-config string   Path to the kerberos configuration file (default "/etc/krb5.conf")
      --ldap.password string      Password
      --ldap.realm string         Realm, set the realm to empty string to use the default realm from config
      --ldap.spn string           Service principal name
      --ldap.username string      Username
  -o, --organization string       Name of Organization (Appscode/Azure/Github/Gitlab/Google/Ldap/Token-Auth).
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

* [guard get](/docs/reference/guard_get.md)	 - Get PKI

