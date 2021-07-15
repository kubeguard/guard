---
title: Guard Get Token
menu:
  product_guard_{{ .version }}:
    identifier: guard-get-token
    name: Guard Get Token
    parent: reference
product_name: guard
menu_name: product_guard_{{ .version }}
section_menu_id: reference
---
## guard get token

Get tokens for Azure/Github/Gitlab/Google/Ldap/Token-Auth

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
  -o, --organization string       Name of Organization (Azure/Github/Gitlab/Google/Ldap/Token-Auth).
```

### SEE ALSO

* [guard get](/docs/reference/guard_get.md)	 - Get PKI

