---
title: Guard Init Server
menu:
  product_guard_{{ .version }}:
    identifier: guard-init-server
    name: Guard Init Server
    parent: reference
product_name: guard
menu_name: product_guard_{{ .version }}
section_menu_id: reference
---
## guard init server

Generate server certificate pair

```
guard init server [flags]
```

### Options

```
      --domains strings   Alternative Domain names
  -h, --help              help for server
      --ips ipSlice       Alternative IP addresses (default [127.0.0.1])
      --pki-dir string    Path to directory where pki files are stored. (default "/Users/tamal/.guard")
```

### SEE ALSO

* [guard init](/docs/reference/guard_init.md)	 - Init PKI

