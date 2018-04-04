---
title: Changelog | Guard
description: Changelog
menu:
  product_stash_0.1.0:
    identifier: changelog-guard
    name: Changelog
    parent: welcome
    weight: 10
product_name: guard
menu_name: product_guard_0.1.0
section_menu_id: welcome
url: /products/guard/0.1.0/welcome/changelog/
aliases:
  - /products/guard/0.1.0/CHANGELOG/
---

# Change Log

## [0.1.0](https://github.com/appscode/guard/tree/0.1.0) (2018-04-04)
[Full Changelog](https://github.com/appscode/guard/compare/0.1.0-rc.5...0.1.0)

**Closed issues:**

- Refactor token command [\#92](https://github.com/appscode/guard/issues/92)
- Enable code coverage tests [\#84](https://github.com/appscode/guard/issues/84)
- Check NTP perodically [\#83](https://github.com/appscode/guard/issues/83)
- Gitlab [\#53](https://github.com/appscode/guard/issues/53)
- Expose metrics via Prometheus [\#52](https://github.com/appscode/guard/issues/52)
- Run guard in its own namespace [\#29](https://github.com/appscode/guard/issues/29)
- Support basic auth / token auth [\#28](https://github.com/appscode/guard/issues/28)
- Run guard on master by default [\#19](https://github.com/appscode/guard/issues/19)
- Installer.yaml should use /healthz checks. [\#17](https://github.com/appscode/guard/issues/17)
- Support LDAP [\#11](https://github.com/appscode/guard/issues/11)
- Support Active Directory / Office365 [\#10](https://github.com/appscode/guard/issues/10)
- Test the big switch statement in server handler [\#96](https://github.com/appscode/guard/issues/96)
- Fix installer [\#91](https://github.com/appscode/guard/issues/91)
- Kerberos [\#58](https://github.com/appscode/guard/issues/58)
- Is the google auth actually working/tested? [\#36](https://github.com/appscode/guard/issues/36)
- Document how to fix bad certificate issue [\#21](https://github.com/appscode/guard/issues/21)
- Fuzz server [\#14](https://github.com/appscode/guard/issues/14)

**Merged pull requests:**

- Update readme. [\#107](https://github.com/appscode/guard/pull/107) ([tamalsaha](https://github.com/tamalsaha))
- Prepare docs for 0.1.0 [\#106](https://github.com/appscode/guard/pull/106) ([tamalsaha](https://github.com/tamalsaha))
- Reorder auth providers [\#105](https://github.com/appscode/guard/pull/105) ([tamalsaha](https://github.com/tamalsaha))
- Update docs [\#104](https://github.com/appscode/guard/pull/104) ([nightfury1204](https://github.com/nightfury1204))
- Add e2e tests [\#103](https://github.com/appscode/guard/pull/103) ([tamalsaha](https://github.com/tamalsaha))
- Use a global var for pki dir [\#102](https://github.com/appscode/guard/pull/102) ([tamalsaha](https://github.com/tamalsaha))
- Fix installer issues [\#101](https://github.com/appscode/guard/pull/101) ([tamalsaha](https://github.com/tamalsaha))
- Refactor commands [\#99](https://github.com/appscode/guard/pull/99) ([tamalsaha](https://github.com/tamalsaha))
- Add kerberos authentication for LDAP [\#98](https://github.com/appscode/guard/pull/98) ([nightfury1204](https://github.com/nightfury1204))
- Refactor token command [\#93](https://github.com/appscode/guard/pull/93) ([tamalsaha](https://github.com/tamalsaha))
- Add coverage tracking using codecov.io [\#90](https://github.com/appscode/guard/pull/90) ([tamalsaha](https://github.com/tamalsaha))
- Add prometheus metrics [\#89](https://github.com/appscode/guard/pull/89) ([tamalsaha](https://github.com/tamalsaha))
- concourse-ci pipeline [\#87](https://github.com/appscode/guard/pull/87) ([tahsinrahman](https://github.com/tahsinrahman))
- Update docs [\#85](https://github.com/appscode/guard/pull/85) ([nightfury1204](https://github.com/nightfury1204))
- Reorg codebase [\#82](https://github.com/appscode/guard/pull/82) ([tamalsaha](https://github.com/tamalsaha))
- Use github.com/json-iterator/go [\#81](https://github.com/appscode/guard/pull/81) ([tamalsaha](https://github.com/tamalsaha))
- Simplify use cluster command [\#80](https://github.com/appscode/guard/pull/80) ([tamalsaha](https://github.com/tamalsaha))
- Ensure max clock skew is no more than 5 sec [\#79](https://github.com/appscode/guard/pull/79) ([tamalsaha](https://github.com/tamalsaha))
- Add travis.yml [\#78](https://github.com/appscode/guard/pull/78) ([tamalsaha](https://github.com/tamalsaha))
- Add test for google [\#77](https://github.com/appscode/guard/pull/77) ([nightfury1204](https://github.com/nightfury1204))
- Validate google IDToken [\#74](https://github.com/appscode/guard/pull/74) ([nightfury1204](https://github.com/nightfury1204))
- Print id\_token & refresh\_token for Google [\#73](https://github.com/appscode/guard/pull/73) ([tamalsaha](https://github.com/tamalsaha))
- Add test for LDAP [\#70](https://github.com/appscode/guard/pull/70) ([nightfury1204](https://github.com/nightfury1204))
- Bug fixes and add CA cert flag for LDAP [\#69](https://github.com/appscode/guard/pull/69) ([nightfury1204](https://github.com/nightfury1204))
- Add test for azure [\#68](https://github.com/appscode/guard/pull/68) ([nightfury1204](https://github.com/nightfury1204))
- Add test for token auth [\#67](https://github.com/appscode/guard/pull/67) ([nightfury1204](https://github.com/nightfury1204))
- Add test for gitlab [\#66](https://github.com/appscode/guard/pull/66) ([nightfury1204](https://github.com/nightfury1204))
- Add test for github [\#65](https://github.com/appscode/guard/pull/65) ([nightfury1204](https://github.com/nightfury1204))
- Add docs to configure guard for Azure ADDS LDAPS [\#64](https://github.com/appscode/guard/pull/64) ([nightfury1204](https://github.com/nightfury1204))
- Use authentication/v1 api [\#63](https://github.com/appscode/guard/pull/63) ([tamalsaha](https://github.com/tamalsaha))
- Various fixes to installer [\#62](https://github.com/appscode/guard/pull/62) ([tamalsaha](https://github.com/tamalsaha))
- Fix Google groups detection [\#61](https://github.com/appscode/guard/pull/61) ([tamalsaha](https://github.com/tamalsaha))
- Various fixes [\#60](https://github.com/appscode/guard/pull/60) ([tamalsaha](https://github.com/tamalsaha))
- Add support for LDAP [\#59](https://github.com/appscode/guard/pull/59) ([nightfury1204](https://github.com/nightfury1204))
- Add support for Azure [\#57](https://github.com/appscode/guard/pull/57) ([nightfury1204](https://github.com/nightfury1204))
- Add support for static token file authentication [\#56](https://github.com/appscode/guard/pull/56) ([nightfury1204](https://github.com/nightfury1204))
- Update client-go to v6.0.0 [\#55](https://github.com/appscode/guard/pull/55) ([tamalsaha](https://github.com/tamalsaha))
- Gitlab [\#54](https://github.com/appscode/guard/pull/54) ([nightfury1204](https://github.com/nightfury1204))
- Document how to use kube-dns to connect api server to guard server [\#47](https://github.com/appscode/guard/pull/47) ([tamalsaha](https://github.com/tamalsaha))

## [0.1.0-rc.5](https://github.com/appscode/guard/tree/0.1.0-rc.5) (2018-01-04)
[Full Changelog](https://github.com/appscode/guard/compare/0.1.0-rc.4...0.1.0-rc.5)

**Closed issues:**

- kops [\#34](https://github.com/appscode/guard/issues/34)
- Document how to use with kops [\#18](https://github.com/appscode/guard/issues/18)

**Merged pull requests:**

- Prepare docs for 0.1.0-rc.5 [\#51](https://github.com/appscode/guard/pull/51) ([tamalsaha](https://github.com/tamalsaha))
- Fix client id generation [\#49](https://github.com/appscode/guard/pull/49) ([tamalsaha](https://github.com/tamalsaha))
- Reorganize & write front matter for 0.1.0-rc.5 [\#48](https://github.com/appscode/guard/pull/48) ([sajibcse68](https://github.com/sajibcse68))
- Use clientcmd apis to generate webhook config [\#46](https://github.com/appscode/guard/pull/46) ([tamalsaha](https://github.com/tamalsaha))
- Use client scheme to convert to YAML [\#45](https://github.com/appscode/guard/pull/45) ([tamalsaha](https://github.com/tamalsaha))
- Use cert store from kutil [\#44](https://github.com/appscode/guard/pull/44) ([tamalsaha](https://github.com/tamalsaha))
- Add kops documentation [\#43](https://github.com/appscode/guard/pull/43) ([tsupertramp](https://github.com/tsupertramp))
- Format error messages [\#42](https://github.com/appscode/guard/pull/42) ([tamalsaha](https://github.com/tamalsaha))
- Generate RBAC roles in installer [\#41](https://github.com/appscode/guard/pull/41) ([tamalsaha](https://github.com/tamalsaha))
- Simplify ClientID generation for analytics [\#40](https://github.com/appscode/guard/pull/40) ([tamalsaha](https://github.com/tamalsaha))
- Correctly set analytics clientID [\#39](https://github.com/appscode/guard/pull/39) ([tamalsaha](https://github.com/tamalsaha))
- Update appscode.com api pkg paths [\#38](https://github.com/appscode/guard/pull/38) ([tamalsaha](https://github.com/tamalsaha))
- Add front mater for docs 0.1.0-rc.4 [\#35](https://github.com/appscode/guard/pull/35) ([sajibcse68](https://github.com/sajibcse68))
- Add front matter for guard cli [\#33](https://github.com/appscode/guard/pull/33) ([tamalsaha](https://github.com/tamalsaha))
- Remove expiration time for appscode token by using validation [\#32](https://github.com/appscode/guard/pull/32) ([sanjid133](https://github.com/sanjid133))
- Cleanup dependencies [\#31](https://github.com/appscode/guard/pull/31) ([tamalsaha](https://github.com/tamalsaha))
- Add appscode authenticator [\#30](https://github.com/appscode/guard/pull/30) ([sanjid133](https://github.com/sanjid133))
- Use client-go 5.x [\#27](https://github.com/appscode/guard/pull/27) ([tamalsaha](https://github.com/tamalsaha))

## [0.1.0-rc.4](https://github.com/appscode/guard/tree/0.1.0-rc.4) (2017-09-25)
[Full Changelog](https://github.com/appscode/guard/compare/0.1.0-rc.3...0.1.0-rc.4)

**Merged pull requests:**

- Add release.sh script [\#26](https://github.com/appscode/guard/pull/26) ([tamalsaha](https://github.com/tamalsaha))
- Add --pki-dir flag [\#25](https://github.com/appscode/guard/pull/25) ([tamalsaha](https://github.com/tamalsaha))
- Revendor dependencies. [\#24](https://github.com/appscode/guard/pull/24) ([tamalsaha](https://github.com/tamalsaha))
- Fix docs of Developer-guide [\#23](https://github.com/appscode/guard/pull/23) ([the-redback](https://github.com/the-redback))

## [0.1.0-rc.3](https://github.com/appscode/guard/tree/0.1.0-rc.3) (2017-09-07)
[Full Changelog](https://github.com/appscode/guard/compare/0.1.0-rc.2...0.1.0-rc.3)

## [0.1.0-rc.2](https://github.com/appscode/guard/tree/0.1.0-rc.2) (2017-09-01)
[Full Changelog](https://github.com/appscode/guard/compare/0.1.0-rc.1...0.1.0-rc.2)

**Merged pull requests:**

- Make sure user of member of Github org or GSuite domain [\#22](https://github.com/appscode/guard/pull/22) ([tamalsaha](https://github.com/tamalsaha))

## [0.1.0-rc.1](https://github.com/appscode/guard/tree/0.1.0-rc.1) (2017-08-30)
[Full Changelog](https://github.com/appscode/guard/compare/0.1.0-rc.0...0.1.0-rc.1)

**Merged pull requests:**

- Improve logging for Guard server [\#20](https://github.com/appscode/guard/pull/20) ([tamalsaha](https://github.com/tamalsaha))

## [0.1.0-rc.0](https://github.com/appscode/guard/tree/0.1.0-rc.0) (2017-08-29)
[Full Changelog](https://github.com/appscode/guard/compare/0.1.0-alpha.0...0.1.0-rc.0)

**Merged pull requests:**

- Document ClusterIP choice [\#16](https://github.com/appscode/guard/pull/16) ([tamalsaha](https://github.com/tamalsaha))
- Various fixes  [\#15](https://github.com/appscode/guard/pull/15) ([tamalsaha](https://github.com/tamalsaha))
- Refactor handlers [\#12](https://github.com/appscode/guard/pull/12) ([tamalsaha](https://github.com/tamalsaha))

## [0.1.0-alpha.0](https://github.com/appscode/guard/tree/0.1.0-alpha.0) (2017-08-28)
**Closed issues:**

- Gtihub Teams [\#2](https://github.com/appscode/guard/issues/2)
- Retrieve all Google groups for a member [\#1](https://github.com/appscode/guard/issues/1)

**Merged pull requests:**

- Add tutorial [\#9](https://github.com/appscode/guard/pull/9) ([tamalsaha](https://github.com/tamalsaha))
- Add kubectl commands. [\#8](https://github.com/appscode/guard/pull/8) ([tamalsaha](https://github.com/tamalsaha))
- Add docs. [\#7](https://github.com/appscode/guard/pull/7) ([tamalsaha](https://github.com/tamalsaha))
- Add `get` commands [\#6](https://github.com/appscode/guard/pull/6) ([tamalsaha](https://github.com/tamalsaha))
- Revise docs [\#5](https://github.com/appscode/guard/pull/5) ([tamalsaha](https://github.com/tamalsaha))
- Implement authN webhook for Google and Github [\#4](https://github.com/appscode/guard/pull/4) ([tamalsaha](https://github.com/tamalsaha))
- Implement init commands [\#3](https://github.com/appscode/guard/pull/3) ([tamalsaha](https://github.com/tamalsaha))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*