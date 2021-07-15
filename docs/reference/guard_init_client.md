---
title: Guard Init Client
menu:
  product_guard_{{ .version }}:
    identifier: guard-init-client
    name: Guard Init Client
    parent: reference
product_name: guard
menu_name: product_guard_{{ .version }}
section_menu_id: reference
---
## guard init client

Generate client certificate pair

```
guard init client [flags]
```

### Options

```
  -h, --help                  help for client
  -o, --organization string   Name of Organization (Azure/Github/Gitlab/Google/Ldap/Token-Auth).
      --pki-dir string        Path to directory where pki files are stored. (default "/home/tamal/.guard")
```

### SEE ALSO

* [guard init](/docs/reference/guard_init.md)	 - Init PKI

