---
title: Install
description: Guard Install
menu:
  product_guard_0.3.0:
    identifier: install-guard
    name: Install
    parent: setup
    weight: 10
product_name: guard
menu_name: product_guard_0.3.0
section_menu_id: setup
---

> New to Guard? Please start [here](/docs/concepts).

# Installation Guide

Guard binary works as a cli and server. In cli mode, you can use `guard` to generate various configuration to easily deploy Guard server. Guard server uses TLS client auth to secure the communication channel between Kubernetes api server and Guard server. You can run Guard server external to a Kubernetes cluster. This document shows you how to `self-host` Guard server in a Kubernetes cluster. To that end, we run Guard server using a predefined Service ClusterIP `10.96.10.96` and port `443`. This ClusterIP is chosen so that it falls in the default --service-cidr range for [Kubeadm](https://kubernetes.io/docs/admin/kubeadm/). If the service CIDR range for your cluster is different, please pick an appropriate ClusterIP.

If you want to set up guard via [__Kops__](https://github.com/kubernetes/kops) visit [here](/docs/setup/install-kops.md) to see differences in setup.

If you want to set up guard via [__Kubespray__](https://github.com/kubernetes-incubator/kubespray) visit [here](/docs/setup/install-kubespray.md).


## Install Guard as CLI
Download pre-built binaries from [appscode/guard Github releases](https://github.com/appscode/guard/releases) and put the binary to some directory in your `PATH`. To install on Linux 64-bit and MacOS 64-bit you can run the following commands:

```console
# Linux amd 64-bit:
wget -O guard https://github.com/appscode/guard/releases/download/0.3.0/guard-linux-amd64 \
  && chmod +x guard \
  && sudo mv guard /usr/local/bin/

# Mac 64-bit
wget -O guard https://github.com/appscode/guard/releases/download/0.3.0/guard-darwin-amd64 \
  && chmod +x guard \
  && sudo mv guard /usr/local/bin/
```

If you prefer to install Guard cli from source code, you will need to set up a GO development environment following [these instructions](https://golang.org/doc/code.html). Then, install `guard` CLI using `go get` from source code.

```bash
go get github.com/appscode/guard
```

Please note that this will install Guard cli from master branch which might include breaking and/or undocumented changes.

## Initialize PKI

Guard uses TLS client certs to secure the communication between guard server and Kubernetes api server. Guard also uses the `CommonName` and `Organization` in client certificate to identify which auth provider to use. Follow the steps below to initialize a self-signed ca, generate a pair of server and client certificates.

```console
# initialize self signed ca
$ guard init ca
Wrote ca certificates in  $HOME/.guard/pki

# generate server certificate pair
$ guard init server --ips=10.96.10.96
Wrote server certificates in  $HOME/.guard/pki

# generate client certificate pair for Github organization `appscode`
$ guard init client appscode -o github
Wrote client certificates in  $HOME/.guard/pki

$ guard init client appscode.com -o google
Wrote client certificates in  $HOME/.guard/pki

$ guard init client qacode -o appscode
Wrote client certificates in  $HOME/.guard/pki

# generate client certificate pair for Gitlab
$ guard init client -o gitlab
Wrote client certificates in  $HOME/.guard/pki

# for azure, commonName is optional
$ guard init client -o azure
Wrote client certificates in  $HOME/.guard/pki

# generate client certificate pair for LDAP
$ guard init client appscode -o ldap
Wrote client certificates in  $HOME/.guard/pki

$ ls -l $HOME/.guard/pki
total 32
-rwxr-xr-- 1 tamal tamal 1054 Aug 28 07:42 qacode@appscode.crt
-rw------- 1 tamal tamal 1679 Aug 28 07:42 qacode@appscode.key
-rwxr-xr-- 1 tamal tamal 1054 Aug 28 07:42 appscode.com@google.crt
-rw------- 1 tamal tamal 1679 Aug 28 07:42 appscode.com@google.key
-rwxr-xr-- 1 tamal tamal 1050 Aug 28 07:12 appscode@github.crt
-rw------- 1 tamal tamal 1675 Aug 28 07:12 appscode@github.key
-rwxr-xr-- 1 tamal tamal 1050 Aug 28 07:12 gitlab@gitlab.crt
-rw------- 1 tamal tamal 1675 Aug 28 07:12 gitlab@gitlab.key
-rwxr-xr-- 1 tamal tamal 1050 Aug 28 07:12 azure@azure.crt
-rw------- 1 tamal tamal 1675 Aug 28 07:12 azure@azure.key
-rwxr-xr-- 1 tamal tamal 1050 Aug 28 07:12 ldap@ldap.crt
-rw------- 1 tamal tamal 1675 Aug 28 07:12 ldap@ldap.key
-rwxr-xr-- 1 tamal tamal 1005 Aug 28 07:12 ca.crt
-rw------- 1 tamal tamal 1675 Aug 28 07:12 ca.key
-rwxr-xr-- 1 tamal tamal 1046 Aug 28 07:12 server.crt
-rw------- 1 tamal tamal 1675 Aug 28 07:12 server.key
```

As you can see, Guard stores the generated certificates in `.guard` subdirectory of home directory of user executing these commands. You can change the location by either setting `GUARD_DATA_DIR` environment variable or by passing the path to a different directory using [`--pki-dir`](https://github.com/appscode/guard/pull/25) flag. For example:

```console
$ export GUARD_DATA_DIR=/tmp/guard

$ guard init ca
I0621 14:09:30.582429   17769 types.go:16] Using data dir /tmp/guard found in GUARD_DATA_DIR env variable
I0621 14:09:30.606327   17769 logs.go:19] FLAG: --alsologtostderr="false"
I0621 14:09:30.606350   17769 logs.go:19] FLAG: --analytics="true"
I0621 14:09:30.606356   17769 logs.go:19] FLAG: --help="false"
I0621 14:09:30.606363   17769 logs.go:19] FLAG: --log_backtrace_at=":0"
I0621 14:09:30.606369   17769 logs.go:19] FLAG: --log_dir=""
I0621 14:09:30.606377   17769 logs.go:19] FLAG: --logtostderr="false"
I0621 14:09:30.606385   17769 logs.go:19] FLAG: --pki-dir="/tmp/guard"
I0621 14:09:30.606392   17769 logs.go:19] FLAG: --stderrthreshold="0"
I0621 14:09:30.606407   17769 logs.go:19] FLAG: --v="0"
I0621 14:09:30.606415   17769 logs.go:19] FLAG: --vmodule=""
Wrote ca certificates in  /tmp/guard/pki
```

Guard can use [supported authenticator](/docs/guides/) to authenticate users for a Kubernetes cluster. A Kubernetes cluster can use one of these organization to authenticate users. But you can configure a single Guard server to perform authentication for multiple clusters, where each cluster uses a different auth provider.

## Deploy Guard server
Now deploy Guard server so that your Kubernetes api server can access it. Use the command below to generate YAMLs for your particular setup. Then use `kubectl apply -f` to install Guard server.

```console
# generate Kubernetes YAMLs for deploying guard server
$ guard get installer \
    --auth-providers=<auth_providers_name> \
    > installer.yaml

$ kubectl apply -f installer.yaml
```

By default, the installer.yaml will deploy Guard server on master instances. If your cluster is provisioned by Kubespray, change
the node selector in installer.yaml to `"node-role.kubernetes.io/master": "true"` due to [kubernetes-incubator/kubespray#2108](https://github.com/kubernetes-incubator/kubespray/issues/2108).

## Configure Kubernetes API Server
To use webhook authentication, you need to set `--authentication-token-webhook-config-file` flag of your Kubernetes api server to a [kubeconfig file](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication) describing how to access the Guard webhook service. You can use the following command to generate a sample `kubeconfig` file.

```console
# print auth token webhook config file. Change the server address to your guard server address.
$ guard get webhook-config appscode -o github --addr=10.96.10.96:443

apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE56QTRNamt4TkRJek16aGFGdzB5TnpBNE1qY3hOREl6TXpoYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBclF3aGRweVE3KzBaVzBDMFpEblI1VS85CjdaVmgyODRtYjdYSmVVaHM0QmI0UHE2Mk1remZpc2lzdXZQZmJzK2Y3dW1oeXhyOGVLak10RlJyc3ZQakhUeDUKTG5rM0hvdE1wNGtCbGJjOTl4dExqN0VwazlXSVNNUW42OEo4K0NCU2N3aFZ1Zlg4NndrNFo3cTZuYXdEUTlRbApyMG1qNWZCVFZ1K3gzYWhXa1F4Rzczd3QxRUdZRkRFMFR1UlV6OXpDclk2bFZzdnZJcXlHL04wWHlUVHNMeFQxCmtOOHRnU3cwVWJOakhGMFVmejhpN05wK1RBZkRFT1pUMWx3SllkSlJ2RjMvYTBRRmdHWTg3K1BJQmhvQklZd1YKU2RXZkRjVWVjcUVuWXkxMEF5MDhZeEZxUTlSNTM4djlpUUxYenZqSVdsYWdDSXBCejB2UDB3dFp4a0lGaFFJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFsanRtbzJwMklvVmhKTXZ4MEFNM08xUkxHMGVRbHY5QTNQWm14dTlWemJkaWZjUFYKTWFaeHp3blVtcjRFa2Y2RmY4WGI4Sk90ZWJUdWo4eDA4UWVaSFRnY3JJYitXdk9jVHJLTkFyeE1Ud0JHcnRRTwp0eWhxanJUSXE5SS9kZUNNeCt4RkNPSFVzdmNEa1FmRVVoZ21YWW5TVEdOU3poNmFmdlU5STFUVUNlWXRFeHlDCm5aaDlPd3hzcGFmcFhBRWdmdStTL0F6WFpLV255bjgyVHpGVTZnaFpFZnVDcndMd2JQRmlaS25ESm1mYlNWM1YKVGljTWdsSkNmZldsdk96OUN3eGtxQlRHNDZvUGZraFZVNElKS3pwekVXQXU2Q3lua09RalkvN2VDd1VFc1NWMQpwM0hZTXpNa05aRzlDNldSakd4VlF4NUo4djRuR1UzcXo2UTk4dz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    server: https://10.96.10.96:443/apis/authentication.k8s.io/v1/tokenreviews
  name: guard-server
contexts:
- context:
    cluster: guard-server
    user: appscode@Github
  name: webhook
current-context: webhook
kind: Config
preferences: {}
users:
- name: appscode@Github
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMyakNDQWNLZ0F3SUJBZ0lJQTZJZmR0dEIzTlV3RFFZSktvWklodmNOQVFFTEJRQXdEVEVMTUFrR0ExVUUKQXhNQ1kyRXdIaGNOTVRjd09ESTVNVFF5TXpNNFdoY05NVGd3T0RJNU1UUXlOREEwV2pBa01ROHdEUVlEVlFRSwpFd1pIYVhSb2RXSXhFVEFQQmdOVkJBTVRDR0Z3Y0hOamIyUmxNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DCkFROEFNSUlCQ2dLQ0FRRUE3NG5nc25IcldkeHovZUNWdEVwVUdVcmQ1dWkwSVhteGFkZlVuM3hpVTduZHdsYzgKc2loT1hLQ2pEaWM5cjZweW16MTJ5Ry93WCtZM1diOUU5a0VxQzg5L2oyR0ZrSFNVOU40VFlsQlZka0pSRUQ2dAoySjZpVnNPZE9BeE9MeDNidG5QR3RYNi9KRTZSTVFCLzhkcUYyenFUZm5ST3FFNFE0N01wL0NKYmdkcEpjbm9mCmdKVzN5Z2pJYk96WHhIZVNNcjlIclhveGRLbGFPVk1hU3BmY0RVREZrQXV1MEREZCszT3hwbG03TlNpS2JpbnUKM0d3VDc4YkRtNzZCTCtEOWhKeXB2OVZDTFkyMVB3WnB3ejJmaThYdzN2WDM3VVdPenA0OXlCMyttYnVjWkpGSgpCRTNORVJyZFVvWFhyeUJrdDhEdXp0SzhYM3dhOExGdmsvOWxWd0lEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFCkJBTUNCYUF3RXdZRFZSMGxCQXd3Q2dZSUt3WUJCUVVIQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFFN0YKZkdpOWpDSzc0eGZZdmY3b3Byb1NxSVpoYmdyODVkcENCZ2FVbEhlK0ZJYkpVeWhnL2c2OVNCOUlhWitpTWYxNgpNWkJVTWs4Tmg5dUZWYWVQVUdCWG0vTUtBeGx0V0J0UHh0VFFMV0cycU9LbXczK3Z5UUJ4R2JMaytycndXVS9YCjR5SzcxbWc5ZG5GVXdWOGI2UWtwM1IxMVdCZzlCeHExU1RQL210S2NyNVVDS3ZWVmlEYlpobzRkRy9ZUHdMbEUKS1N4UGtqS3NPVU54dDlHZkdNTHVRMVFQWXZDYTZjZ1MxN3BERWZacWVqdmthc0UxeTNSQWpPa2pubTJ2KzIzbgpBN2kxYWwwMDZjTEpQdTl3S3IyMzhJQ0Q3N2ZxQzhTaXJhN3V3NUgxbkVvMTh1am94NmI2UG5XazdmeTZabDRHCjBWMXBCSDJtL1JqK0RudGVTM0k9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBNzRuZ3NuSHJXZHh6L2VDVnRFcFVHVXJkNXVpMElYbXhhZGZVbjN4aVU3bmR3bGM4CnNpaE9YS0NqRGljOXI2cHltejEyeUcvd1grWTNXYjlFOWtFcUM4OS9qMkdGa0hTVTlONFRZbEJWZGtKUkVENnQKMko2aVZzT2RPQXhPTHgzYnRuUEd0WDYvSkU2Uk1RQi84ZHFGMnpxVGZuUk9xRTRRNDdNcC9DSmJnZHBKY25vZgpnSlczeWdqSWJPelh4SGVTTXI5SHJYb3hkS2xhT1ZNYVNwZmNEVURGa0F1dTBERGQrM094cGxtN05TaUtiaW51CjNHd1Q3OGJEbTc2QkwrRDloSnlwdjlWQ0xZMjFQd1pwd3oyZmk4WHczdlgzN1VXT3pwNDl5QjMrbWJ1Y1pKRkoKQkUzTkVScmRVb1hYcnlCa3Q4RHV6dEs4WDN3YThMRnZrLzlsVndJREFRQUJBb0lCQVFDam9LdTlPZFJyTGd5TwpBRHhEVEFMbXhCMlEvcVVOdVBOWU9mY2tldk12L21kZHVmbmNPV3hPR2UxSVhjWGxtYWx3SWl4aC94VlViUTZpClgrWGIwZWZHNlpkWmVtU2lxUUNYeEp1NUxPYzBRVmplbi9KaFp2dStDU0g4aDJ0aEJDUnlIZVEvVnJWN043QTIKcVFDOVZXamF1TWpJT09zQ1RWRjhPWWNVbE9PdGJ2MFFRSUFNRDdxak1jcTdIRlJMbjlHSXJjSGVZWTd2RTdONQpucFdOMWZOT25BZXNnaVNRcGYxdGNYcmJUckNkb1JweExlU2RFUngyVnBkdFE1V3hFTm9YYURJT3FJVXB2anJaCjIwSUZqazY3eVVOTlBCLzJkU3VmZzVaekxEYWNGUy9lenYvUDVJYnVEazBHZWQyZGhkT2t3K2duVFNabjg0Sk4KMDg4TlMvVUJBb0dCQVBGOG44dEVmTURDNjlTc1RGWWh4NUc2aE1PWThlS0kxTDJQeTVVaEN4WCtXUlpoQndaVQpWaGhVMVpOcTgvUXkwb201WEQyT1UwR0tGQjY3aE9kU1l6TUc3a0NqKzl3bHZDM0loRmxLd2R1ZC8relBWTFkvCngxRjFtVnA3aTllZXZiZmdZdUQxdDRvTDVxc2lveldGZ2g2UndSZzVWTXpGd0N2WXZwUk9CcGZUQW9HQkFQM3YKUjNyRit1cW96YU9IcldrUjFEL3E5MHlRd3ZpamlRWDhyNEg3QVZTVVpnQmhwbEpmUDRTVU84NWdUbmhBLzl4ZQp3R1NTQzFGWkpVeDRmOW41MFN0dFlkNFhIMTZiU1lyOEpWQWxVRWY5QWR1czBkc1pKdkZLSEM5S3VjNEFPeUFPClYvUGR0Rk1FV2J5WERSQXdBTTVwNzZsU1pZVXA5cDFLVTZ6SndHM3RBb0dCQUlmdzdRK0RmV3NTRDZwSVdDekEKbFZUL0Y4LzRZR3B6TnJlRHBFcE9NS3h2NDN6S29DYTdBVUJ2T1Uva2pISnl6YnlFSVYzeHFnS2lGVk43b29TSwpCNWZwRmVSRHEvdXhMbTdqaTBXczVOYVo2a0ZJTWRycXFteTc4OWxRNVZjN1lIZUxsSDRwTk9vOGF0ejZBY0NXCmFMcUd1Sm5IWkdwbUJCbHF5Vlk1V2xMTEFvR0FPWVlBelVFWURCeGRLUlJOSmlZUnpNRHZjSHJDa0F5THQ3MTgKREpmTnYxazJtaE9FMTlnWHpYSys4WXREZTE1T0Y1K25PYUVUeTBQRWZVUTJ3aXdqUkJFdFFHQkFqTy9rZ3dXSApkbFpkajFFeklJNVBvN0JZOEFQM3lvYkUvSE4wOFZnT2VJSGFuWXU0d0UzL2VaRkdQWHdsL0ZkY0JBUnpoMElWCkhtazluQ2tDZ1lFQXJyNGQ3WFBUNlBvTTA5TjFSYnVtQnROUW96cGxpb0wwb0tqZUxHL0RHcUd5Q3lpcDVsNnMKUW5vblpPcW9BVzdqZlRiMThvVVVzVkJTdDJvd20zempac05XZFdZcEptL1ZCSDQvY0dOeFJuSXN1OFNsdFdBZAp1YjVZQS9YcnEvUTRNcUJnZEVwYm9HTnlzUVlscE0rMmxHZWpiZ0pDUTI4b3dJYndLQ3JSY2t3PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
```


## Using Guard service name instead of IP in authentication webhook config file

So far, we have used a preset ClusterIP address in Guard service. This cluster ip was used in the authentication webhook config file and used by Kubernetes api server to connect to guard server.

It is possible to use Guard service name in authentication webhook config file so that Kubernetes api server connects to Guard using its domain name. This requires the following additional steps:

- Since Kubernetes api server pod uses `HostNetwok`, change the DNS policy for Kubernetes api server to [`ClusterFirstWithHostNet`](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pods-dns-policy). The actual process depends on your cluster provisioning process. Usually this involves updating the kube-apiserver manifest file in /etc/kubernetes/manifests folder in Master machines.

- When issuing server certificate for Guard server, provide the domain name so that it is included in CN/SANS for server cert.

```console
$ guard init server --domains=guard.<namespace>.svc
```

- Now, pass the Guard service name when generating webhook config.

```console
$ guard get webhook-config appscode -o github --addr=guard.<namespace>.svc:443
``
