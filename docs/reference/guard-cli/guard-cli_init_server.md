---
title: Init Server
menu:
  product_guard_0.1.0-rc.5:
    identifier: guard-cli-init-server
    name: Init Server
    parent: guard-cli
product_name: guard
section_menu_id: reference
menu_name: product_guard_0.1.0-rc.5
---
## guard-cli init server

Generate server certificate pair

### Synopsis

Generate server certificate pair

```
guard-cli init server [flags]
```

### Options

```
      --domains stringSlice   Alternative Domain names
  -h, --help                  help for server
      --ips ipSlice           Alternative IP addresses (default [127.0.0.1])
      --pki-dir string        Path to directory where pki files are stored. (default "$HOME/.guard")
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

* [guard-cli init](/docs/reference/guard-cli/guard-cli_init.md)	 - Init PKI

