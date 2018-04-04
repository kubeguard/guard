[![Go Report Card](https://goreportcard.com/badge/github.com/appscode/guard)](https://goreportcard.com/report/github.com/appscode/guard)
[![Build Status](https://travis-ci.org/appscode/guard.svg?branch=master)](https://travis-ci.org/appscode/guard)
[![codecov](https://codecov.io/gh/appscode/guard/branch/master/graph/badge.svg)](https://codecov.io/gh/appscode/guard)
[![Docker Pulls](https://img.shields.io/docker/pulls/appscode/guard.svg)](https://hub.docker.com/r/appscode/guard/)
[![Slack](https://slack.appscode.com/badge.svg)](https://slack.appscode.com)
[![Twitter](https://img.shields.io/twitter/follow/appscodehq.svg?style=social&logo=twitter&label=Follow)](https://twitter.com/intent/follow?screen_name=AppsCodeHQ)

# Guard
 Guard by AppsCode is a [Kubernetes Webhook Authentication](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication) server. Using guard, you can log into your Kubernetes cluster using your Github or Google authentication token. Guard also sets authenticated user's groups to his Github teams or Google groups. This allows cluster administrator to setup RBAC rules based on membership in Github teams or Google groups.

## Supported Versions
Kubernetes 1.8+

## Installation
To install Guard, please follow the guide [here](https://appscode.com/products/guard/0.1.0/setup/install/).

## Using Guard
Want to learn how to use Guard? Please start [here](https://appscode.com/products/guard/0.1.0/).

## Contribution guidelines
Want to help improve Guard? Please start [here](https://appscode.com/products/guard/0.1.0/welcome/contributing/).

---

**Guard binaries collects anonymous usage statistics to help us learn how the software is being used and how we can improve it. To disable stats collection, run the operator with the flag** `--analytics=false`.

---

## Acknowledgement

- [apprenda-kismatic/kubernetes-ldap](https://github.com/apprenda-kismatic/kubernetes-ldap)
- [Nike-Inc/harbormaster](https://github.com/Nike-Inc/harbormaster)

## Support
We use Slack for public discussions. To chit chat with us or the rest of the community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8M8HANQ0/details/) channel `#guard`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

If you have found a bug with Searchlight or want to request for new features, please [file an issue](https://github.com/appscode/guard/issues/new).
