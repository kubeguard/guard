---
title: Troubleshoot Guard Installer
menu:
  product_guard_0.1.2:
    identifier: troubleshoot-guard
    name: Troubleshooting
    parent: setup
    weight: 25
product_name: guard
menu_name: product_guard_0.1.2
section_menu_id: setup
---

> New to Guard? Please start [here](/docs/concepts).

## Troubleshooting

### How do I debug and fix bad certificate issues?

Say, you are seeing logs like below in Guard server logs:

```
I0830 16:41:59.919947       1 logs.go:19] FLAG: --alsologtostderr="false"
I0830 16:41:59.919987       1 logs.go:19] FLAG: --analytics="true"
I0830 16:41:59.920016       1 logs.go:19] FLAG: --ca-cert-file="/srv/guard/pki/ca.crt"
I0830 16:41:59.920026       1 logs.go:19] FLAG: --cert-file="/srv/guard/pki/tls.crt"
I0830 16:41:59.920032       1 logs.go:19] FLAG: --help="false"
I0830 16:41:59.920039       1 logs.go:19] FLAG: --key-file="/srv/guard/pki/tls.key"
I0830 16:41:59.920067       1 logs.go:19] FLAG: --log_backtrace_at=":0"
I0830 16:41:59.920075       1 logs.go:19] FLAG: --log_dir=""
I0830 16:41:59.920080       1 logs.go:19] FLAG: --logtostderr="true"
I0830 16:41:59.920086       1 logs.go:19] FLAG: --ops-addr=":56790"
I0830 16:41:59.920091       1 logs.go:19] FLAG: --stderrthreshold="2"
I0830 16:41:59.920098       1 logs.go:19] FLAG: --v="3"
I0830 16:41:59.920105       1 logs.go:19] FLAG: --vmodule=""
I0830 16:41:59.920112       1 logs.go:19] FLAG: --web-address=":9844"
I0830 16:42:45.430823       1 logs.go:19] http: TLS handshake error from 1.1.2.6:34028: remote error: tls: bad certificate
I0830 16:43:00.483658       1 logs.go:19] http: TLS handshake error from 1.1.2.6:34062: remote error: tls: bad certificate
```

```
admin@ip-172-20-48-207:~$ sudo tail -f /var/log/kube-apiserver.log | grep auth
E0830 16:56:46.430468       6 authentication.go:58] Unable to authenticate the request due to an error: [invalid bearer token, [invalid bearer token, [invalid bearer token, invalid bearer token, Post https://10.7.3.3:9844/apis/authentication.k8s.io/v1beta1/tokenreviews: x509: certificate is valid for 1.27.55.255, not 10.7.3.3]]]
```

To debug this issue, follow the steps below:

1. First note the ip address in the authentication webhook config file. This is the ip address used by Kubernetes api server to connect to the guard server.

2. Now check the common name(CN) and subject alternative names (SANS) in the server.crt. If that does not include the ip addressfrom step 1, we need to regenerate the server certificates.

```
$ guard init server --ips=<ip1>,<ip2> --domains=<dns1>,<dns2>
```

Then regenerate the installer.yaml . This includes a Secret called `guard-pki` which is used to mount the server.crt.

Then `kubectl apply` the new installer.yaml. This should restart the guard server with the updated cert and fix the `bad certificate` issue.
