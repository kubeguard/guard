<p align="center"><img src="/logo/LOGO_GUARD_Readme.jpg"></p>

[![Go Report Card](https://goreportcard.com/badge/github.com/appscode/guard)](https://goreportcard.com/report/github.com/appscode/guard)
[![Build Status](https://travis-ci.org/appscode/guard.svg?branch=master)](https://travis-ci.org/appscode/guard)
[![codecov](https://codecov.io/gh/appscode/guard/branch/master/graph/badge.svg)](https://codecov.io/gh/appscode/guard)
[![Docker Pulls](https://img.shields.io/docker/pulls/appscode/guard.svg)](https://hub.docker.com/r/appscode/guard/)
[![Slack](https://slack.appscode.com/badge.svg)](https://slack.appscode.com)
[![Twitter](https://img.shields.io/twitter/follow/appscodehq.svg?style=social&logo=twitter&label=Follow)](https://twitter.com/intent/follow?screen_name=AppsCodeHQ)

# Guard
Guard by AppsCode is a [Kubernetes Webhook Authentication](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication) server. Using guard, you can log into your Kubernetes cluster using various auth providers. Guard also configures groups of authenticated user appropriately. This allows cluster administrator to setup RBAC rules based on membership in groups. Guard supports following auth providers:

- [Static Token File](https://appscode.com/products/guard/0.3.0/guides/authenticator/static_token_file/)
- [Github](https://appscode.com/products/guard/0.3.0/guides/authenticator/github/)
- [Gitlab](https://appscode.com/products/guard/0.3.0/guides/authenticator/gitlab/)
- [Google](https://appscode.com/products/guard/0.3.0/guides/authenticator/google/)
- [Azure](https://appscode.com/products/guard/0.3.0/guides/authenticator/azure/)
- [LDAP using Simple or Kerberos authentication](https://appscode.com/products/guard/0.3.0/guides/authenticator/ldap/)
- [Azure Active Directory via LDAP](https://appscode.com/products/guard/0.3.0/guides/authenticator/ldap_azure/)

## Supported Versions
Kubernetes 1.9+

## Installation
To install Guard, please follow the guide [here](https://appscode.com/products/guard/0.3.0/setup/install/).

## Using Guard
Want to learn how to use Guard? Please start [here](https://appscode.com/products/guard/0.3.0/).

## Contribution guidelines
Want to help improve Guard? Please start [here](https://appscode.com/products/guard/0.3.0/welcome/contributing/).

---

**Guard binaries collects anonymous usage statistics to help us learn how the software is being used and how we can improve it. To disable stats collection, run the operator with the flag** `--analytics=false`.

---

## Acknowledgement

- [apprenda-kismatic/kubernetes-ldap](https://github.com/apprenda-kismatic/kubernetes-ldap)
- [Nike-Inc/harbormaster](https://github.com/Nike-Inc/harbormaster)

## Support
We use Slack for public discussions. To chit chat with us or the rest of the community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8M8HANQ0/details/) channel `#guard`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

If you have found a bug with Guard or want to request for new features, please [file an issue](https://github.com/appscode/guard/issues/new).

<p align="center"><img src="/logo/Separador.jpg"></p>
