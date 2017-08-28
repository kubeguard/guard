> New to Guard? Please start [here](/docs/tutorial.md).

# Installation Guide

Guard binary works as a cli and server. In cli mode, you can use guard to generate various configuration to easily deploy guard server.

## Install Guard as CLI
Download pre-built binaries from [appscode/guard Github releases](https://github.com/appscode/guard/releases) and put the binary to some directory in your `PATH`. To install on Linux 64-bit and MacOS 64-bit you can run the following commands:

```console
# Linux amd 64-bit:
wget -O guard https://github.com/appscode/guard/releases/download/0.1.0-alpha.0/guard-linux-amd64 \
  && chmod +x guard \
  && sudo mv guard /usr/local/bin/

# Mac 64-bit
wget -O guard https://github.com/appscode/guard/releases/download/0.1.0-alpha.0/guard-darwin-amd64 \
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
Wrote ca certificates in  /home/tamal/.guard/pki

# generate server certificate pair
$ guard init server
Wrote server certificates in  /home/tamal/.guard/pki

# generate client certificate pair for Github organization `appscode`
$ guard init client appscode -o github
Wrote client certificates in  /home/tamal/.guard/pki

$ guard init client appscode.com -o google
Wrote client certificates in  /home/tamal/.guard/pki

$ ls -l /home/tamal/.guard/pki
total 32
-rwxr-xr-- 1 tamal tamal 1054 Aug 28 07:42 appscode.com@google.crt
-rw------- 1 tamal tamal 1679 Aug 28 07:42 appscode.com@google.key
-rwxr-xr-- 1 tamal tamal 1050 Aug 28 07:12 appscode@github.crt
-rw------- 1 tamal tamal 1675 Aug 28 07:12 appscode@github.key
-rwxr-xr-- 1 tamal tamal 1005 Aug 28 07:12 ca.crt
-rw------- 1 tamal tamal 1675 Aug 28 07:12 ca.key
-rwxr-xr-- 1 tamal tamal 1046 Aug 28 07:12 server.crt
-rw------- 1 tamal tamal 1675 Aug 28 07:12 server.key
```

As you can see, Guard stores the generated certificates in `.guard` subdirectory of home directory of user executing these commands. Guard can use either a Github organization or a Google Apps domain (now G Suite) to authenticate users for a Kubernetes cluster. A Kubernetes cluster can use one of these organization to authenticate users. But you can configure a single Guard server to perform authentication for multiple clusters, where each cluster uses a different auth provider.

## Deploy Guard server
Now deploy Guard server so that your Kubernetes api server can access it. You can self-host Guard server in a Kubernetes cluster. Use the command below to generate YAMLs for your particular setup. Then use `kubectl apply -f` to install Guard server.

```console
# generate Kubernetes YAMLs for deploying guard server
$ guard get installer > docs/examples/installer.yaml
$ kubectl apply -f docs/examples/installer.yaml
```

## Configure Kubernetes API Server
To use webhook authentication, you need to set `--authentication-token-webhook-config-file` flag of your Kubernetes api server to a [kubeconfig file](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication) describing how to access the Guard webhook service. You can use the following command to generate a sample `kubeconfig` file.

```console
# print auth token webhook config file. Change the server address to your guard server address.
$ guard get webhook-config appscode -o github

apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE56QTRNamd4TkRFeU1UZGFGdzB5TnpBNE1qWXhOREV5TVRkYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdmNpdWowWEFCcmFxbTlEbEljZVY3RjNkCk9FNktRMm1JaEdWRmxsNGRydEdEMjhOUThMUnlXTFZhMUR4Q01Eb2FJRHBTMzc0QXRNeWFsdnUzWFBONUszYzEKdDJ3TE5RZnZjUEV1cjN6eHVDL3pKbXozUGRsOEpES3VpOGs4emFzVFBmaUNLUHh1S0Q4cytQOUtRS1QxcFhvSQpRaXRKZENPQ0ZCT0tWbzdsTlNBOEd3K0orenJhUW9WdEZrYkVOOG10UVM0UUthSmRTVjVsc3JldWt5YkNuNUs5ClJDaVJ4OVFZSEpRdUxXVjB5c0x2Qnloakl3Tjc3TTlHei9GbFlxdVFTdklOc2VvcWpXSkRPb01pUXB4N2V4TXoKTFgyRzB3TTR2MmNkSG5xUWdyeW4vY2ZkUkVpck90TDNFaVFja0ZmNTV4ZHRSOHBiSURUdEVlUUJFK0VsUlFJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFXNk5INXJ0UVdydnBqdU1DU2dHNitYUVljZWwvN0RZY3QwUHNmZmQ3bzNEM3huNEMKUnh4TFJSc09GdXdGb3p6dU5MQmlTVXlSdmNWcnJLN3o4UDMyaktPWlY3Tk83blMyY3NsOE1FVnJFVEtGWkxCaApYOVJ2aURSZW5wZkdNR2F6WFhKbE5xMTVaYXZ2WE5KTXBYbHdhYkhqZWt0WU1XbkNUcG5RaTNvR3MzYW0zQWRpCno3UTA1SGtFNnVIV0cwc3VIVHp4dmlrUzdhRUNJZzhqczBaK0Uxa1Myc2tLVzV2NzN6bVhFR0paakRWaUgwd3UKbDZLRVlPRDRHWmQ0cEdZTFd3VWdBVmZib05lS3E1dnczanFWcjNQS01rang5M0tOamhuOHpuQmVlOGh0cjBzQwpRN0tGekc1K2grMUlCOHMycWlVdXBRU0s3eU0xc1doZDgvQzh2QT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    server: http(s)://>guard-server-host:port>/apis/authentication.k8s.io/v1beta1/tokenreviews
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
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMyakNDQWNLZ0F3SUJBZ0lJUkJTSzdrcFI5YWN3RFFZSktvWklodmNOQVFFTEJRQXdEVEVMTUFrR0ExVUUKQXhNQ1kyRXdIaGNOTVRjd09ESTRNVFF4TWpFM1doY05NVGd3T0RJNE1UUXhNak16V2pBa01ROHdEUVlEVlFRSwpFd1pIYVhSb2RXSXhFVEFQQmdOVkJBTVRDR0Z3Y0hOamIyUmxNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DCkFROEFNSUlCQ2dLQ0FRRUF6Zmd2MlFWbWdtVGRXaVB1V1I2TWV0RVlJdmpIQXRrZGR6Qk5ITFdva1hxcG5EUHcKT3JDcnduWS9XdVZRMlZ5QnRPeHRxNzJQV2NkVWZUdk55MElaWEl0NVFXRnp4U3V4WHBLcGRoK1UvRXVML1dkdgp5ZEVGYy9KbFl2Qk4yeThYbUhkUHVrVzJmZy9BdXZUTExjQjB2V1NTY0RnT0FETTl5USs1ZFVEMWZObmVHaWJLCnl2ZGpWcVNsaUZ4MDhQenZzQ2hMNDdFN2Q0NFlKQnJDR2dVNG55b3BNbGZPanA3VXZxb25LWmVSZURYcGtJREkKcjVQaTNUQlRLc0RHZlBLNmZSV3pqSkdHVDEraVVuUXc2NGtQMUYraTVlQnorNHdBejFERzg2dElPY1l4Z3NJUgpiWW1oanRRdnVkTzZtMXVCeDdGaVU4SDB1Y052WkpYYXVwRitud0lEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFCkJBTUNCYUF3RXdZRFZSMGxCQXd3Q2dZSUt3WUJCUVVIQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFKL2wKWENPZjFtUi9SWjNsc3hlMmw5ZExzSXdrUzdGVmZLRFJYNGJ4N2tSWkNYZUpLVjdNcTBCN2pmNlFWVEFDMGdxNwpEWkk3NXJvR2JvMkFDNW4wZmJlMms5Z3ZnMExkU1lmWDdRdG0wcEtUZEJ5SmdBa2JEbWNYMERIVFdvaGZlaU1sCmRKb05CdENURlZITlZ4djNVVkUzUGVnVEp2NFpnNFV2OXZzY2plOG9IcEpqWUt5WGNieTdFWGlBTFRvakxkS2wKdVRxcmszeEwwQXlVdDFEd1phQXRHYjZiS1VWZkJnaFlBdWw3MjBzUG9hbkpUUG90WFZueGpEYTNWZ2JwWWpmeAovaEtCbGY0cnZqaXRKcS9Pa1RQYmNBeUFsYk1abUpQakg3VjdjcmJid3I2Q2pKQk5wK2NlbFNFcnJRUEhUVUc4CnVmOVYxSkM4SzZ4V3BwK09Oek09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBemZndjJRVm1nbVRkV2lQdVdSNk1ldEVZSXZqSEF0a2RkekJOSExXb2tYcXBuRFB3Ck9yQ3J3blkvV3VWUTJWeUJ0T3h0cTcyUFdjZFVmVHZOeTBJWlhJdDVRV0Z6eFN1eFhwS3BkaCtVL0V1TC9XZHYKeWRFRmMvSmxZdkJOMnk4WG1IZFB1a1cyZmcvQXV2VExMY0IwdldTU2NEZ09BRE05eVErNWRVRDFmTm5lR2liSwp5dmRqVnFTbGlGeDA4UHp2c0NoTDQ3RTdkNDRZSkJyQ0dnVTRueW9wTWxmT2pwN1V2cW9uS1plUmVEWHBrSURJCnI1UGkzVEJUS3NER2ZQSzZmUld6akpHR1QxK2lVblF3NjRrUDFGK2k1ZUJ6KzR3QXoxREc4NnRJT2NZeGdzSVIKYlltaGp0UXZ1ZE82bTF1Qng3RmlVOEgwdWNOdlpKWGF1cEYrbndJREFRQUJBb0lCQUJZczhubmJLdjQrL2RZRwpxRkxRQnkveDh6ZFFzWnlNNDY5QmRBdnpRd0xtd2Z3ZlcyMzJONGZvVTBubUJwNWdaMWFqWGV0dnJVOXROcHVGCkpOTU5lczJMOXJJalcrT09vTG9kOEVEWGhhTGtQMTQ1Rk1BVjBSQjZ1SSsrYjZONW5vQzcxOG1tVjgrYXdwVmUKWmNHM1J0cGRTYWJyWWhhRGJkT0g3ZC9CN3F4U1Z5Y0J4cDdtNHZnM05NWnNlWG1VakM5VDNpdEJ6Ty9KMVlDYgpac1pRTjZ4WGJRZVFqa3JlMER5a3ZrekJmQk1pai9tZWVpQTc2Z01idEEvT0pFM0NCL2NBZ3l3cVBoZ2ZIZWVHCjBMVTR5amEzUmF2OUllMzZjWnhPQVBaMDFrQ0JkaGRsVlRDaG0wQy9OcDltdFdMSDB0WGRnaERUbk5mSDFCc3YKV0Vwd3lJRUNnWUVBMndFVm9ZY2R6d1UrTTN3V09BQjcvdHd0cldxL0tnU24rTjlndEtNV2N5Y3RkaWpxVm9oQgpIS2hreHFwaU1CTXV1d1hPUklZcEQ2WC9kTmhrcWpaTGR6bmJxdVJpdFJ1SHoxSlpZbUVubFJ6M0JXWHM2TG1qCkk4Nk04Y2E0eXJMVlE2NWpYeHA3YkFLdzQ1TkJiY3lvaWJpUWhCZzlyemNWMGpwVDljcnZLMFVDZ1lFQThNTm8KNXhCR3BTOWNyNnEwaVZoOEZtdDh2ZG5qMXNMREdhcElBY3JaOGtxOXFsMndBNEt5cGtHZ3N0K1VRd3FGQ1lkRgpQeUVTTjYxbndYV3d0dFpqSXdIbk85eUZoUkpHZXhxaTdON0ovVmk3d2ZQekNmdk5BNmJOT0Vkd2sva0ZBZE5mCiswQTdveTBqQjdPNnd1S0pqTUFyQjRJaHE5aHNLcktNcHd4VWJwTUNnWUE5VDFscDVmU2ZYeDFodG14Vjh6VEQKVFlwd0VRRkJWeHBiSHRYbzIvdE44M3JUcUhLcUZPejlnOXJxandwNzRQTGxJcVB6SlFmYnZLSCthUklOWUxQUgp4ZDNNUXJHcmQvQ1dScnlGUVNPZXFBUXplNnhPSHFJZ1JSUEtIOUxkMUNER0dNenk4K3YzZWUxaFdIa3BydkRECjFXcUh3RzJNWHNSNkhTQWlJRlRDYlFLQmdBZTIrMEdNTC9kVEVURS8weEVqbUxaUE0yd1I4MDhLWnA0SDZzN0QKNVQveVRTbU1YdnQ5MEtPckxxOE1vditTOHJoZmNVU1lsckRhQ1owVlhGZy9mbVc4eGVBUkxPWWFzODkyQndwNApDUmpwSXZzUUNoV2p6K255Q2xsblVLQXROby9jYWhMdTkvbytsQVRIS1pEZEdYTTlKU1BVYzZmQ0E1VktxMThlCjhnV3BBb0dCQUs3Q0QreFovbzhKQlRWajdpQW9NRy9xV3EraHJ0K1RpU0ZzRW8vTEwydmhMWW1KeDMybGV0YmcKc05wbDZwWlUvUU56MW5QZlJmVU1nbkF2NnZxbW10ZUxDWUtUdmdwaW8zZUxwWjNzRFoxejVmSVBYb0RVcDBHOApPODQwOGJYVExvNW8vY29JbFBiUklFdS9xU2VYb3REZWFac0EwNjhidSszUC9pUEgrZFpXCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==
```

## Issue Token

To use Github authentication, you can use your personal access token with permissions to read `public_repo` and `read:org`. You can use the following command to issue a token:
```
$ guard get token -o github
```
![github-token](/docs/images/github-token.png)


To use Google authentication, you need a token with the following OAuth scopes:
 - https://www.googleapis.com/auth/userinfo.email
 - https://www.googleapis.com/auth/admin.directory.group.readonly

You can use the following command to issue a token:
```
$ guard get token -o google
```
This will run a local HTTP server to issue a token with appropriate OAuth2 scopes.


## Configure kubectl

```console
$ kubectl config set-cluster NAME [--server=server] [--certificate-authority=path/to/certificate/authority] [--insecure-skip-tls-verify=true]
$ kubectl config set-credentials NAME [--token=bearer_token] [--auth-provider=provider_name] [--auth-provider-arg=key=value]
$ kubectl config set-context NAME [--cluster=cluster_nickname] [--user=user_nickname] [--namespace=namespace]
$ kubectl config use-context CONTEXT_NAME
$ kubectl config view
```
