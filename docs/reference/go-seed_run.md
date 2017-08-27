## kad run

Run server

### Synopsis


Run server

```
kad run [flags]
```

### Options

```
      --caCertFile string    File containing CA certificate
      --certFile string      File container server TLS certificate
      --kad              Send analytical events to Google Kad (default true)
  -h, --help                 help for run
      --keyFile string       File containing server TLS private key
      --ops-addr string      Address to listen on for web interface and telemetry. (default ":56790")
      --web-address string   Http server address (default ":9844")
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [kad](kad.md)	 - Kad by AppsCode - Essential kad for OSS

