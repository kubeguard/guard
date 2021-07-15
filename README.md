<p align="center"><img src="/logo/LOGO_GUARD_Readme.jpg"></p>

[![Build Status](https://github.com/kubeguard/guard/workflows/CI/badge.svg)](https://github.com/kubeguard/guard/actions?workflow=CI)
[![codecov](https://codecov.io/gh/kubeguard/guard/branch/master/graph/badge.svg)](https://codecov.io/gh/kubeguard/guard)
[![Docker Pulls](https://img.shields.io/docker/pulls/appscode/guard.svg)](https://hub.docker.com/r/appscode/guard/)
[![Twitter](https://img.shields.io/twitter/follow/kubeguard.svg?style=social&logo=twitter&label=Follow)](https://twitter.com/intent/follow?screen_name=KubeGuard)

# Guard
Guard by AppsCode is a [Kubernetes Webhook Authentication](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication) server. Using guard, you can log into your Kubernetes cluster using various auth providers. Guard also configures groups of authenticated user appropriately. This allows cluster administrator to setup RBAC rules based on membership in groups. Guard supports following auth providers:

- [Static Token File](https://appscode.com/products/guard/latest/guides/authenticator/static_token_file/)
- [Github](https://appscode.com/products/guard/latest/guides/authenticator/github/)
- [Gitlab](https://appscode.com/products/guard/latest/guides/authenticator/gitlab/)
- [Google](https://appscode.com/products/guard/latest/guides/authenticator/google/)
- [Azure](https://appscode.com/products/guard/latest/guides/authenticator/azure/)
- [LDAP using Simple or Kerberos authentication](https://appscode.com/products/guard/latest/guides/authenticator/ldap/)
- [Azure Active Directory via LDAP](https://appscode.com/products/guard/latest/guides/authenticator/ldap_azure/)

## Supported Versions
Kubernetes 1.9+

## Installation
To install Guard, please follow the guide [here](https://appscode.com/products/guard/latest/setup/install/).

## Using Guard
Want to learn how to use Guard? Please start [here](https://appscode.com/products/guard/latest/).

## Contribution guidelines
Want to help improve Guard? Please start [here](https://appscode.com/products/guard/latest/welcome/contributing/).

## Acknowledgement

- [apprenda-kismatic/kubernetes-ldap](https://github.com/apprenda-kismatic/kubernetes-ldap)
- [Nike-Inc/harbormaster](https://github.com/Nike-Inc/harbormaster)

## Support
We use Slack for public discussions. To chit chat with us or the rest of the community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8M8HANQ0/details/) channel `#guard`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

If you have found a bug with Guard or want to request for new features, please [file an issue](https://github.com/kubeguard/guard/issues/new).

<p align="center"><img src="/logo/Separador.jpg"></p>
