---
title: Welcome | Guard
description: Welcome to Guard
menu:
  product_guard_0.1.0-rc.5:
    identifier: readme-guard
    name: Readme
    parent: welcome
    weight: -1
product_name: guard
menu_name: product_guard_0.1.0-rc.5
section_menu_id: welcome
url: /products/guard/0.1.0-rc.5/welcome/
aliases:
  - /products/guard/0.1.0-rc.5/
  - /products/guard/0.1.0-rc.5/README/
---

# Guard

Guard by AppsCode is a [Kubernetes Webhook Authentication](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication) server. Using guard, you can log into your Kubernetes cluster using your Github or Google authentication token. Guard also sets authenticated user's groups to his Github teams or Google groups. This allows cluster administrator to setup RBAC rules based on membership in Github teams or Google groups.

From here you can learn all about Guard's architecture and how to deploy and use Guard.

- [Concepts](/docs/concepts/). Concepts explain some significant aspect of Guard. This is where you can learn about what Guard does and how it does it.

- [Setup](/docs/setup/). Setup contains instructions for installing
  the Guard in various cloud providers.

- [Guides](/docs/guides/). Guides show you how to perform tasks with Guard.

- [Reference](/docs/reference/). Detailed exhaustive lists of
command-line Options, configuration Options, API definitions, and procedures.

We're always looking for help improving our documentation, so please don't hesitate to [file an issue](https://github.com/appscode/guard/issues/new) if you see some problem. Or better yet, submit your own [contributions](/docs/CONTRIBUTING.md) to help
make our docs better.

---

**Guard binaries collects anonymous usage statistics to help us learn how the software is being used and how we can improve it. To disable stats collection, run the operator with the flag** `--analytics=false`.

---
