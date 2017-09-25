## guard init client

Generate client certificate pair

### Synopsis


Generate client certificate pair

```
guard init client [flags]
```

### Options

```
  -h, --help                  help for client
  -o, --organization string   Name of Organization (Github or Google).
      --pki-dir string        Path to directory where pki files are stored. (default user's home directory)
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
* [guard init](guard_init.md)	 - Init PKI

