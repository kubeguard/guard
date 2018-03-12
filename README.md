[![Build Status](https://travis-ci.org/appscode/guard.svg?branch=master)](https://travis-ci.org/appscode/guard)
[![Generic Badge](http://159.65.228.198:8080/api/v1/teams/main/pipelines/guard/jobs/test-guard/badge)](http://159.65.228.198:8080/teams/main/pipelines/guard)
[![Generic badge](https://img.shields.io/badge/<SUBJECT>-<STATUS>-<COLOR>.svg)](https://google.com)
[![Go Report Card](https://goreportcard.com/badge/appscode/guard "Go Report Card")](https://goreportcard.com/report/appscode/guard)
[![GoDoc](https://godoc.org/github.com/appscode/guard?status.svg "GoDoc")](https://godoc.org/github.com/appscode/guard)

# Guard
 Guard by AppsCode is a [Kubernetes Webhook Authentication](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication) server. Using guard, you can log into your Kubernetes cluster using your Github or Google authentication token. Guard also sets authenticated user's groups to his Github teams or Google groups. This allows cluster administrator to setup RBAC rules based on membership in Github teams or Google groups.

## Supported Versions
Kubernetes 1.8+

## Installation
To install Guard, please follow the guide [here](https://appscode.com/products/guard/0.1.0-rc.5/setup/install/).

## Using Guard
Want to learn how to use Guard? Please start [here](https://appscode.com/products/guard/0.1.0-rc.5/).

## Contribution guidelines
Want to help improve Guard? Please start [here](https://appscode.com/products/guard/0.1.0-rc.5/welcome/contributing/).

---

**Guard binaries collects anonymous usage statistics to help us learn how the software is being used and how we can improve it. To disable stats collection, run the operator with the flag** `--analytics=false`.

---

## Acknowledgement

- [apprenda-kismatic/kubernetes-ldap](https://github.com/apprenda-kismatic/kubernetes-ldap)
- [Nike-Inc/harbormaster](https://github.com/Nike-Inc/harbormaster)

## Support
We use Slack for public discussions. To chit chat with us or the rest of the community, join us in the [AppsCode Slack team](https://appscode.slack.com/messages/C8M8HANQ0/details/) channel `#guard`. To sign up, use our [Slack inviter](https://slack.appscode.com/).

If you have found a bug with Searchlight or want to request for new features, please [file an issue](https://github.com/appscode/guard/issues/new).
