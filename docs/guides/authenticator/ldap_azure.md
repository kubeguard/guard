---
title: Azure Active Directory | Guard
description: Authenticate into Kubernetes using Azure Active Directory
menu:
  product_guard_0.1.4:
    identifier: azure-ad-authenticator
    parent: authenticator-guides
    name: Azure AD
    weight: 40
product_name: guard
menu_name: product_guard_0.1.4
section_menu_id: guides
---

# Authenticate using secure LDAP of Azure Active Directory Domain Services

There is a nice documentation about how to enable secure LDAP for the managed domain using Azure portal [here](https://docs.microsoft.com/en-us/azure/active-directory-domain-services/active-directory-ds-admin-guide-configure-secure-ldap-enable-ldaps). If you configured DNS to access the managed domain, then use it as `SERVER_ADDRESS`. If not configured, then you can use **EXTERNAL IP ADDRESS FOR LDAPS ACCESS** as `SERVER_ADDRESS`. For LDAPS use `636` as server `PORT`. Procedure to find **EXTERNAL IP ADDRESS FOR LDAPS ACCESS** is given below:

1.  Write `domain services` in the Search resources search box. Select Azure AD Domain Services from the search result.

![azure-azure-ADDS](/docs/images/ldap-azure/azure-ADDS.png)

2.  Click the name of the managed domain(for example: appscode.com)

![azure-ADDS-home](/docs/images/ldap-azure/azure-ADDS-home.png)

3.  Click the **Properties** and find the **SECURE LDAP EXTERNAL IP ADDRESS**

![azure-ADDS-properties](/docs/images/ldap-azure/azure-ADDS-properties.png)

> **Note:** guard uses `SERVER_ADDRESS` as **Server Name** in TLS verification when `--ldap-skip-tls-verification` flag is set to `false`. So, please remember this fact when generating certificates.

### Guard Installation and PKI Initialization

Guide for Guard installation and PKI initialization are given [here](/docs/setup/install.md).
Create a client cert with `Organization` set to `Ldap`.For LDAP `COMMON_NAME` is optional. To ease this process, use the Guard cli to issue a client cert/key pair.

```console
    # If COMMON_NAME is not provided, then default COMMON_NAME `ldap` is used
    $ guard init client [COMMON_NAME] -o Ldap
```

### Deploy guard server

Now deploy Guard server so that your Kubernetes api server can access it. Use the command below to generate YAMLs for your particular setup. Then use `kubectl apply -f` to install Guard server.
```console
     # generate Kubernetes YAMLs for deploying guard server
     $  guard get installer \
            --auth-providers="ldap" \
            --ldap.server-address=[SERVER_ADDRESS] \
            --ldap.server-port=636 \
            --ldap.bind-dn=[BIND_DN] \
            --ldap.bind-password=[BIND_PASSWORD] \
            --ldap.user-search-dn=[USER_SEARCH_DN] \
            --ldap.user-search-filter=[USER_SEARCH_FILTER] \
            --ldap.user-attribute=[USER_ATTRIBUTE] \
            --ldap.group-search-dn=[GROUP_SEARCH_DN] \
            --ldap.group-search-filter=[GROUP_SEARCH_FILTER] \
            --ldap.group-name-attribute=[GROUP_NAME_ATTRIBUTE] \
            --ldap.group-member-attribute=[GROUP_MEMBER_ATTRIBUTE] \
            --ldap.skip-tls-verification=[true/false] \
            --ldap.start-tls=false\
            --ldap.is-secure-ldap=true\
            > installer.yaml

     $ kubectl apply -f installer.yaml
```

> **Note:** Azure managed domain LDAPS doesn't support start TLS

### Configure Kubernetes API Server
To use webhook authentication, you need to set `--authentication-token-webhook-config-file` flag of your Kubernetes api server to a [kubeconfig file](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication) describing how to access the Guard webhook service. You can use the following command to generate a sample `kubeconfig` file.

```console
# print auth token webhook config file. Change the server address to your guard server address.
$ guard get webhook-config [COMMON_NAME]-o Ldap --addr=10.96.10.96:443

apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWFDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFOTVFzd0NRWURWUVFERXdKallUQWUKRncweE56QTRNamt4TkRJek16aGFGdzB5TnpBNE1qY3hOREl6TXpoYU1BMHhDekFKQmdOVkJBTVRBbU5oTUlJQgpJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBclF3aGRweVE3KzBaVzBDMFpEblI1VS85CjdaVmgyODRtYjdYSmVVaHM0QmI0UHE2Mk1remZpc2lzdXZQZmJzK2Y3dW1oeXhyOGVLak10RlJyc3ZQakhUeDUKTG5rM0hvdE1wNGtCbGJjOTl4dExqN0VwazlXSVNNUW42OEo4K0NCU2N3aFZ1Zlg4NndrNFo3cTZuYXdEUTlRbApyMG1qNWZCVFZ1K3gzYWhXa1F4Rzczd3QxRUdZRkRFMFR1UlV6OXpDclk2bFZzdnZJcXlHL04wWHlUVHNMeFQxCmtOOHRnU3cwVWJOakhGMFVmejhpN05wK1RBZkRFT1pUMWx3SllkSlJ2RjMvYTBRRmdHWTg3K1BJQmhvQklZd1YKU2RXZkRjVWVjcUVuWXkxMEF5MDhZeEZxUTlSNTM4djlpUUxYenZqSVdsYWdDSXBCejB2UDB3dFp4a0lGaFFJRApBUUFCb3lNd0lUQU9CZ05WSFE4QkFmOEVCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBTkJna3Foa2lHCjl3MEJBUXNGQUFPQ0FRRUFsanRtbzJwMklvVmhKTXZ4MEFNM08xUkxHMGVRbHY5QTNQWm14dTlWemJkaWZjUFYKTWFaeHp3blVtcjRFa2Y2RmY4WGI4Sk90ZWJUdWo4eDA4UWVaSFRnY3JJYitXdk9jVHJLTkFyeE1Ud0JHcnRRTwp0eWhxanJUSXE5SS9kZUNNeCt4RkNPSFVzdmNEa1FmRVVoZ21YWW5TVEdOU3poNmFmdlU5STFUVUNlWXRFeHlDCm5aaDlPd3hzcGFmcFhBRWdmdStTL0F6WFpLV255bjgyVHpGVTZnaFpFZnVDcndMd2JQRmlaS25ESm1mYlNWM1YKVGljTWdsSkNmZldsdk96OUN3eGtxQlRHNDZvUGZraFZVNElKS3pwekVXQXU2Q3lua09RalkvN2VDd1VFc1NWMQpwM0hZTXpNa05aRzlDNldSakd4VlF4NUo4djRuR1UzcXo2UTk4dz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    server: https://10.96.10.96:443/apis/authentication.k8s.io/v1/tokenreviews
  name: guard-server
contexts:
- context:
    cluster: guard-server
    user: client@ldap
  name: webhook
current-context: webhook
kind: Config
preferences: {}
users:
- name: client@ldap
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMyakNDQWNLZ0F3SUJBZ0lJQTZJZmR0dEIzTlV3RFFZSktvWklodmNOQVFFTEJRQXdEVEVMTUFrR0ExVUUKQXhNQ1kyRXdIaGNOTVRjd09ESTVNVFF5TXpNNFdoY05NVGd3T0RJNU1UUXlOREEwV2pBa01ROHdEUVlEVlFRSwpFd1pIYVhSb2RXSXhFVEFQQmdOVkJBTVRDR0Z3Y0hOamIyUmxNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DCkFROEFNSUlCQ2dLQ0FRRUE3NG5nc25IcldkeHovZUNWdEVwVUdVcmQ1dWkwSVhteGFkZlVuM3hpVTduZHdsYzgKc2loT1hLQ2pEaWM5cjZweW16MTJ5Ry93WCtZM1diOUU5a0VxQzg5L2oyR0ZrSFNVOU40VFlsQlZka0pSRUQ2dAoySjZpVnNPZE9BeE9MeDNidG5QR3RYNi9KRTZSTVFCLzhkcUYyenFUZm5ST3FFNFE0N01wL0NKYmdkcEpjbm9mCmdKVzN5Z2pJYk96WHhIZVNNcjlIclhveGRLbGFPVk1hU3BmY0RVREZrQXV1MEREZCszT3hwbG03TlNpS2JpbnUKM0d3VDc4YkRtNzZCTCtEOWhKeXB2OVZDTFkyMVB3WnB3ejJmaThYdzN2WDM3VVdPenA0OXlCMyttYnVjWkpGSgpCRTNORVJyZFVvWFhyeUJrdDhEdXp0SzhYM3dhOExGdmsvOWxWd0lEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFCkJBTUNCYUF3RXdZRFZSMGxCQXd3Q2dZSUt3WUJCUVVIQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFFN0YKZkdpOWpDSzc0eGZZdmY3b3Byb1NxSVpoYmdyODVkcENCZ2FVbEhlK0ZJYkpVeWhnL2c2OVNCOUlhWitpTWYxNgpNWkJVTWs4Tmg5dUZWYWVQVUdCWG0vTUtBeGx0V0J0UHh0VFFMV0cycU9LbXczK3Z5UUJ4R2JMaytycndXVS9YCjR5SzcxbWc5ZG5GVXdWOGI2UWtwM1IxMVdCZzlCeHExU1RQL210S2NyNVVDS3ZWVmlEYlpobzRkRy9ZUHdMbEUKS1N4UGtqS3NPVU54dDlHZkdNTHVRMVFQWXZDYTZjZ1MxN3BERWZacWVqdmthc0UxeTNSQWpPa2pubTJ2KzIzbgpBN2kxYWwwMDZjTEpQdTl3S3IyMzhJQ0Q3N2ZxQzhTaXJhN3V3NUgxbkVvMTh1am94NmI2UG5XazdmeTZabDRHCjBWMXBCSDJtL1JqK0RudGVTM0k9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBNzRuZ3NuSHJXZHh6L2VDVnRFcFVHVXJkNXVpMElYbXhhZGZVbjN4aVU3bmR3bGM4CnNpaE9YS0NqRGljOXI2cHltejEyeUcvd1grWTNXYjlFOWtFcUM4OS9qMkdGa0hTVTlONFRZbEJWZGtKUkVENnQKMko2aVZzT2RPQXhPTHgzYnRuUEd0WDYvSkU2Uk1RQi84ZHFGMnpxVGZuUk9xRTRRNDdNcC9DSmJnZHBKY25vZgpnSlczeWdqSWJPelh4SGVTTXI5SHJYb3hkS2xhT1ZNYVNwZmNEVURGa0F1dTBERGQrM094cGxtN05TaUtiaW51CjNHd1Q3OGJEbTc2QkwrRDloSnlwdjlWQ0xZMjFQd1pwd3oyZmk4WHczdlgzN1VXT3pwNDl5QjMrbWJ1Y1pKRkoKQkUzTkVScmRVb1hYcnlCa3Q4RHV6dEs4WDN3YThMRnZrLzlsVndJREFRQUJBb0lCQVFDam9LdTlPZFJyTGd5TwpBRHhEVEFMbXhCMlEvcVVOdVBOWU9mY2tldk12L21kZHVmbmNPV3hPR2UxSVhjWGxtYWx3SWl4aC94VlViUTZpClgrWGIwZWZHNlpkWmVtU2lxUUNYeEp1NUxPYzBRVmplbi9KaFp2dStDU0g4aDJ0aEJDUnlIZVEvVnJWN043QTIKcVFDOVZXamF1TWpJT09zQ1RWRjhPWWNVbE9PdGJ2MFFRSUFNRDdxak1jcTdIRlJMbjlHSXJjSGVZWTd2RTdONQpucFdOMWZOT25BZXNnaVNRcGYxdGNYcmJUckNkb1JweExlU2RFUngyVnBkdFE1V3hFTm9YYURJT3FJVXB2anJaCjIwSUZqazY3eVVOTlBCLzJkU3VmZzVaekxEYWNGUy9lenYvUDVJYnVEazBHZWQyZGhkT2t3K2duVFNabjg0Sk4KMDg4TlMvVUJBb0dCQVBGOG44dEVmTURDNjlTc1RGWWh4NUc2aE1PWThlS0kxTDJQeTVVaEN4WCtXUlpoQndaVQpWaGhVMVpOcTgvUXkwb201WEQyT1UwR0tGQjY3aE9kU1l6TUc3a0NqKzl3bHZDM0loRmxLd2R1ZC8relBWTFkvCngxRjFtVnA3aTllZXZiZmdZdUQxdDRvTDVxc2lveldGZ2g2UndSZzVWTXpGd0N2WXZwUk9CcGZUQW9HQkFQM3YKUjNyRit1cW96YU9IcldrUjFEL3E5MHlRd3ZpamlRWDhyNEg3QVZTVVpnQmhwbEpmUDRTVU84NWdUbmhBLzl4ZQp3R1NTQzFGWkpVeDRmOW41MFN0dFlkNFhIMTZiU1lyOEpWQWxVRWY5QWR1czBkc1pKdkZLSEM5S3VjNEFPeUFPClYvUGR0Rk1FV2J5WERSQXdBTTVwNzZsU1pZVXA5cDFLVTZ6SndHM3RBb0dCQUlmdzdRK0RmV3NTRDZwSVdDekEKbFZUL0Y4LzRZR3B6TnJlRHBFcE9NS3h2NDN6S29DYTdBVUJ2T1Uva2pISnl6YnlFSVYzeHFnS2lGVk43b29TSwpCNWZwRmVSRHEvdXhMbTdqaTBXczVOYVo2a0ZJTWRycXFteTc4OWxRNVZjN1lIZUxsSDRwTk9vOGF0ejZBY0NXCmFMcUd1Sm5IWkdwbUJCbHF5Vlk1V2xMTEFvR0FPWVlBelVFWURCeGRLUlJOSmlZUnpNRHZjSHJDa0F5THQ3MTgKREpmTnYxazJtaE9FMTlnWHpYSys4WXREZTE1T0Y1K25PYUVUeTBQRWZVUTJ3aXdqUkJFdFFHQkFqTy9rZ3dXSApkbFpkajFFeklJNVBvN0JZOEFQM3lvYkUvSE4wOFZnT2VJSGFuWXU0d0UzL2VaRkdQWHdsL0ZkY0JBUnpoMElWCkhtazluQ2tDZ1lFQXJyNGQ3WFBUNlBvTTA5TjFSYnVtQnROUW96cGxpb0wwb0tqZUxHL0RHcUd5Q3lpcDVsNnMKUW5vblpPcW9BVzdqZlRiMThvVVVzVkJTdDJvd20zempac05XZFdZcEptL1ZCSDQvY0dOeFJuSXN1OFNsdFdBZAp1YjVZQS9YcnEvUTRNcUJnZEVwYm9HTnlzUVlscE0rMmxHZWpiZ0pDUTI4b3dJYndLQ3JSY2t3PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
```
### Configure Kubectl
```console
kubectl config set-credentials [USERNAME] --token=[TOKEN]
```

Or You can add user in ~/.kube/config file
```yaml
...
users:
- name: [USERNAME]
  user:
    token: [TOKEN]
```

### How to generate token

For LDAP, user is authenticated using user DN and password. User DN is collected using `[USERNAME]` and `[USER_ATTRIBUTE]`.
So, use base64 encoded string of `[USERNAME]:[PASSWORD]` as token.

|username |password |username:password     |token
|---------|---------|----------------------|------------------
|user12   |12345    |user12:12345          |dXNlcjEyOjEyMzQ1

