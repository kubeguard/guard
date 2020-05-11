---
title: Changelog | Guard
description: Changelog
menu:
  product_guard_{{ .version }}:
    identifier: changelog-guard
    name: Changelog
    parent: welcome
    weight: 10
product_name: guard
menu_name: product_guard_{{ .version }}
section_menu_id: welcome
url: /products/guard/{{ .version }}/welcome/changelog/
aliases:
  - /products/guard/{{ .version }}/CHANGELOG/
---

# Change Log

## [v0.5.0](https://github.com/appscode/guard/tree/v0.5.0) (2020-05-11)
[Full Changelog](https://github.com/appscode/guard/compare/v0.5.0-rc.1...v0.5.0)

**Merged pull requests:**

- Use Go 1.14.2 [\#251](https://github.com/appscode/guard/pull/251) ([tamalsaha](https://github.com/tamalsaha))
- Azure: switch to new graph endpoint for US Government cloud [\#250](https://github.com/appscode/guard/pull/250) ([karataliu](https://github.com/karataliu))

## [v0.5.0-rc.1](https://github.com/appscode/guard/tree/v0.5.0-rc.1) (2020-02-16)
[Full Changelog](https://github.com/appscode/guard/compare/v0.5.0-rc.0...v0.5.0-rc.1)

**Closed issues:**

- azure auth provider should validate token audience [\#244](https://github.com/appscode/guard/issues/244)
- Enable guard to use AKS On-Behalf-Of flow to obtain user's group memberships [\#235](https://github.com/appscode/guard/issues/235)
- Don't query for groups if AAD token already contains groups [\#171](https://github.com/appscode/guard/issues/171)
- Azure AAD [\#152](https://github.com/appscode/guard/issues/152)
- Enable guard to use OAuth 2.0 On-Behalf-Of flow to obtain user's group memberships [\#131](https://github.com/appscode/guard/issues/131)

**Merged pull requests:**

- Add blank line before package delcaration [\#247](https://github.com/appscode/guard/pull/247) ([tamalsaha](https://github.com/tamalsaha))
- Add option to verify client ID [\#246](https://github.com/appscode/guard/pull/246) ([weinong](https://github.com/weinong))
- Add flag to call Graph api only when overage indicator is present [\#245](https://github.com/appscode/guard/pull/245) ([weinong](https://github.com/weinong))

## [v0.5.0-rc.0](https://github.com/appscode/guard/tree/v0.5.0-rc.0) (2020-01-26)
[Full Changelog](https://github.com/appscode/guard/compare/0.4.0...v0.5.0-rc.0)

**Closed issues:**

- Use github.com/glauth/glauth LDAP server for testing [\#220](https://github.com/appscode/guard/issues/220)
- Unable to build docker image from src [\#217](https://github.com/appscode/guard/issues/217)

**Merged pull requests:**

- Update dependencies [\#240](https://github.com/appscode/guard/pull/240) ([tamalsaha](https://github.com/tamalsaha))
- Update azure doc with new feature [\#239](https://github.com/appscode/guard/pull/239) ([weinong](https://github.com/weinong))
- Implement On-Behalf-Of \(OBO\) flow [\#236](https://github.com/appscode/guard/pull/236) ([weinong](https://github.com/weinong))
- Update client-go to kubernetes-1.16.3 [\#234](https://github.com/appscode/guard/pull/234) ([tamalsaha](https://github.com/tamalsaha))
- fix small errors [\#233](https://github.com/appscode/guard/pull/233) ([torubylist](https://github.com/torubylist))
- Fix Linter Issues [\#231](https://github.com/appscode/guard/pull/231) ([faem](https://github.com/faem))
- Various Makefile improvements [\#230](https://github.com/appscode/guard/pull/230) ([tamalsaha](https://github.com/tamalsaha))
- Templatize front matter [\#229](https://github.com/appscode/guard/pull/229) ([tamalsaha](https://github.com/tamalsaha))
- migrate from apps/v1beta1 to apps/v1 [\#225](https://github.com/appscode/guard/pull/225) ([qw1mb0](https://github.com/qw1mb0))
- Fix LDAP test [\#223](https://github.com/appscode/guard/pull/223) ([nightfury1204](https://github.com/nightfury1204))
- Add Makefile [\#222](https://github.com/appscode/guard/pull/222) ([tamalsaha](https://github.com/tamalsaha))
- Use absolute path as aliases for reference docs [\#221](https://github.com/appscode/guard/pull/221) ([tamalsaha](https://github.com/tamalsaha))
- Update to k8s 1.14.0 client libraries using go.mod [\#219](https://github.com/appscode/guard/pull/219) ([tamalsaha](https://github.com/tamalsaha))
- Update Kubernetes client libraries to 1.13.5 [\#218](https://github.com/appscode/guard/pull/218) ([tamalsaha](https://github.com/tamalsaha))
- fix typo [\#216](https://github.com/appscode/guard/pull/216) ([zabio3](https://github.com/zabio3))
- typo in error command the it should be -o [\#214](https://github.com/appscode/guard/pull/214) ([kanolato](https://github.com/kanolato))

## [0.4.0](https://github.com/appscode/guard/tree/0.4.0) (2019-02-04)
[Full Changelog](https://github.com/appscode/guard/compare/0.3.0...0.4.0)

**Closed issues:**

- Support Azure Active Directory in sovereign cloud  [\#208](https://github.com/appscode/guard/issues/208)
- Dependabot couldn't find a Gopkg.toml for this project [\#205](https://github.com/appscode/guard/issues/205)

**Merged pull requests:**

- Prepare docs for 0.4.0 release [\#213](https://github.com/appscode/guard/pull/213) ([tamalsaha](https://github.com/tamalsaha))
- Document how to use Azure sovereign cloud instances [\#212](https://github.com/appscode/guard/pull/212) ([tamalsaha](https://github.com/tamalsaha))
- Revendor dependencies [\#211](https://github.com/appscode/guard/pull/211) ([tamalsaha](https://github.com/tamalsaha))
- Add a little bit more guidance for users. [\#210](https://github.com/appscode/guard/pull/210) ([coderanger](https://github.com/coderanger))
- Support aad sovereign clouds [\#209](https://github.com/appscode/guard/pull/209) ([karataliu](https://github.com/karataliu))

## [0.3.0](https://github.com/appscode/guard/tree/0.3.0) (2018-12-03)
[Full Changelog](https://github.com/appscode/guard/compare/0.2.1...0.3.0)

**Fixed bugs:**

- Fix panic if there is no $HOME/.kube/config [\#179](https://github.com/appscode/guard/pull/179) ([nightfury1204](https://github.com/nightfury1204))

**Closed issues:**

- Github connector docs asks for unnecessary public\_repo permissions [\#195](https://github.com/appscode/guard/issues/195)
- Guard should not log password/secrets [\#194](https://github.com/appscode/guard/issues/194)
- Issues with LDAP and guard get installer [\#193](https://github.com/appscode/guard/issues/193)
- Dependabot couldn't find a Gopkg.toml for this project [\#191](https://github.com/appscode/guard/issues/191)
- Guard crashes multiple times a day [\#190](https://github.com/appscode/guard/issues/190)
- Missing binaries for 0.2.1 guard-darwin [\#189](https://github.com/appscode/guard/issues/189)
- Guard pod keeps restarting ~ every 8 hours [\#187](https://github.com/appscode/guard/issues/187)
- Logo Proposal for Guard [\#184](https://github.com/appscode/guard/issues/184)
- Using Gitlab's full\_path instead of name while using authentication by Groups [\#181](https://github.com/appscode/guard/issues/181)
- panic: assignment to entry in nil map [\#178](https://github.com/appscode/guard/issues/178)

**Merged pull requests:**

- Use docker build --pull [\#204](https://github.com/appscode/guard/pull/204) ([tamalsaha](https://github.com/tamalsaha))
- Prepare docs for 0.3.0 release [\#203](https://github.com/appscode/guard/pull/203) ([tamalsaha](https://github.com/tamalsaha))
- Fix Gitlab tests [\#202](https://github.com/appscode/guard/pull/202) ([tamalsaha](https://github.com/tamalsaha))
- Support group id or full\_path for Gitlab connector [\#201](https://github.com/appscode/guard/pull/201) ([tamalsaha](https://github.com/tamalsaha))
- Document Github connector permissions to read:org [\#200](https://github.com/appscode/guard/pull/200) ([tamalsaha](https://github.com/tamalsaha))
- Redact password/secrets when dumping flags. [\#199](https://github.com/appscode/guard/pull/199) ([tamalsaha](https://github.com/tamalsaha))
- Revendor and run gofmt [\#198](https://github.com/appscode/guard/pull/198) ([tamalsaha](https://github.com/tamalsaha))
- Set periodic analytics [\#197](https://github.com/appscode/guard/pull/197) ([tamalsaha](https://github.com/tamalsaha))
- Update Kubernetes client libraries to 1.12.0 [\#196](https://github.com/appscode/guard/pull/196) ([tamalsaha](https://github.com/tamalsaha))
- Typos [\#192](https://github.com/appscode/guard/pull/192) ([tmatias](https://github.com/tmatias))
- Use kubernetes-1.11.3 [\#188](https://github.com/appscode/guard/pull/188) ([tamalsaha](https://github.com/tamalsaha))
- Add readme images to logo folder [\#186](https://github.com/appscode/guard/pull/186) ([tamalsaha](https://github.com/tamalsaha))
- Add Logos for Guard [\#185](https://github.com/appscode/guard/pull/185) ([area55git](https://github.com/area55git))
- Cleanup error handling for Azure provider [\#182](https://github.com/appscode/guard/pull/182) ([tamalsaha](https://github.com/tamalsaha))
- Log userinfo at glog.V\(10\) [\#180](https://github.com/appscode/guard/pull/180) ([tamalsaha](https://github.com/tamalsaha))

## [0.2.1](https://github.com/appscode/guard/tree/0.2.1) (2018-07-10)
[Full Changelog](https://github.com/appscode/guard/compare/0.2.0...0.2.1)

**Closed issues:**

- Support B2B auth for Azure provider by supporting both `oid` or `upn` claims in the token [\#170](https://github.com/appscode/guard/issues/170)
- Update kops docs [\#142](https://github.com/appscode/guard/issues/142)

**Merged pull requests:**

- Fix hugo frontmatter [\#177](https://github.com/appscode/guard/pull/177) ([tamalsaha](https://github.com/tamalsaha))
- Prepare 0.2.1 release [\#176](https://github.com/appscode/guard/pull/176) ([tamalsaha](https://github.com/tamalsaha))
- Format shell scripts [\#175](https://github.com/appscode/guard/pull/175) ([tamalsaha](https://github.com/tamalsaha))
- Use client-go v8.0.0 [\#173](https://github.com/appscode/guard/pull/173) ([tamalsaha](https://github.com/tamalsaha))
- Enable B2B auth for Azure provider by supporting either `oid` or `upn` claim in the token [\#172](https://github.com/appscode/guard/pull/172) ([amanohar](https://github.com/amanohar))
- Add missing image in azure [\#169](https://github.com/appscode/guard/pull/169) ([nightfury1204](https://github.com/nightfury1204))

## [0.2.0](https://github.com/appscode/guard/tree/0.2.0) (2018-06-22)
[Full Changelog](https://github.com/appscode/guard/compare/0.1.4...0.2.0)

**Closed issues:**

- Use a GUARD\_PKI\_DIR env variable [\#158](https://github.com/appscode/guard/issues/158)
- Azure AAD auth provider is using AAD group's displayName instead of unique objectId for auth [\#153](https://github.com/appscode/guard/issues/153)

**Merged pull requests:**

- Prepare 0.2.0 release [\#168](https://github.com/appscode/guard/pull/168) ([tamalsaha](https://github.com/tamalsaha))
- Fix flaky LDAP tests [\#167](https://github.com/appscode/guard/pull/167) ([tamalsaha](https://github.com/tamalsaha))
- Fix bad pointer in frame ldap.handleBind [\#166](https://github.com/appscode/guard/pull/166) ([tamalsaha](https://github.com/tamalsaha))
- Document --azure.use-group-uid flag [\#165](https://github.com/appscode/guard/pull/165) ([tamalsaha](https://github.com/tamalsaha))
- Add GUARD\_DATA\_DIR env variable [\#164](https://github.com/appscode/guard/pull/164) ([tamalsaha](https://github.com/tamalsaha))
- Various fixed based on goreportcard [\#163](https://github.com/appscode/guard/pull/163) ([tamalsaha](https://github.com/tamalsaha))
- Fix test command in make.py [\#162](https://github.com/appscode/guard/pull/162) ([tamalsaha](https://github.com/tamalsaha))
- Skip e2e tests from travis. [\#161](https://github.com/appscode/guard/pull/161) ([tamalsaha](https://github.com/tamalsaha))
- Fix formatting error [\#160](https://github.com/appscode/guard/pull/160) ([tamalsaha](https://github.com/tamalsaha))
- Fix test [\#159](https://github.com/appscode/guard/pull/159) ([nightfury1204](https://github.com/nightfury1204))
- Vendor aws sdk [\#157](https://github.com/appscode/guard/pull/157) ([tamalsaha](https://github.com/tamalsaha))
- User auth info added for AWS EKS [\#150](https://github.com/appscode/guard/pull/150) ([sanjid133](https://github.com/sanjid133))

## [0.1.4](https://github.com/appscode/guard/tree/0.1.4) (2018-06-20)
[Full Changelog](https://github.com/appscode/guard/compare/0.1.3...0.1.4)

**Closed issues:**

- NTP sync causing periodic crashes [\#143](https://github.com/appscode/guard/issues/143)
- Hardcoded Oauth client/secret for google? [\#138](https://github.com/appscode/guard/issues/138)
- Add paging to get around directoryObjects.getByIds limit of 1000 [\#132](https://github.com/appscode/guard/issues/132)

**Merged pull requests:**

- Prepare docs for 0.1.4 release [\#155](https://github.com/appscode/guard/pull/155) ([tamalsaha](https://github.com/tamalsaha))
- Allow Azure AAD auth provider to use AAD group ids instead of display name for authn/authz [\#154](https://github.com/appscode/guard/pull/154) ([amanohar](https://github.com/amanohar))

## [0.1.3](https://github.com/appscode/guard/tree/0.1.3) (2018-06-06)
[Full Changelog](https://github.com/appscode/guard/compare/0.1.2...0.1.3)

**Merged pull requests:**

- Prepare docs for 0.1.3 [\#147](https://github.com/appscode/guard/pull/147) ([tamalsaha](https://github.com/tamalsaha))
- Support ntp server as flag [\#145](https://github.com/appscode/guard/pull/145) ([tamalsaha](https://github.com/tamalsaha))
- Increase NTP clock skew to 2 min and check every 10 min [\#144](https://github.com/appscode/guard/pull/144) ([tamalsaha](https://github.com/tamalsaha))
- Fix typo [\#141](https://github.com/appscode/guard/pull/141) ([ryuheechul](https://github.com/ryuheechul))
- Fix a typo in kubectl invocation [\#137](https://github.com/appscode/guard/pull/137) ([farcaller](https://github.com/farcaller))

## [0.1.2](https://github.com/appscode/guard/tree/0.1.2) (2018-05-04)
[Full Changelog](https://github.com/appscode/guard/compare/0.1.1...0.1.2)

**Merged pull requests:**

- Update docs for 0.1.2 release [\#136](https://github.com/appscode/guard/pull/136) ([tamalsaha](https://github.com/tamalsaha))
- Update client-go to 7.0.0 \(Kubernetes 1.10\) [\#135](https://github.com/appscode/guard/pull/135) ([tamalsaha](https://github.com/tamalsaha))
- Add paging to get around directoryObjects.getByIds limit of 1000 [\#133](https://github.com/appscode/guard/pull/133) ([amanohar](https://github.com/amanohar))

## [0.1.1](https://github.com/appscode/guard/tree/0.1.1) (2018-04-20)
[Full Changelog](https://github.com/appscode/guard/compare/0.1.0...0.1.1)

**Fixed bugs:**

- Error: guard does not provide service for github [\#127](https://github.com/appscode/guard/issues/127)

**Closed issues:**

- Test the big switch statement in server handler [\#96](https://github.com/appscode/guard/issues/96)

**Merged pull requests:**

- Prepare docs for 0.1.1 release [\#130](https://github.com/appscode/guard/pull/130) ([tamalsaha](https://github.com/tamalsaha))
- Revendor dependencies [\#129](https://github.com/appscode/guard/pull/129) ([tamalsaha](https://github.com/tamalsaha))
- Add auth provider case insensitive check [\#128](https://github.com/appscode/guard/pull/128) ([nightfury1204](https://github.com/nightfury1204))
- Improve azure auth provider docs [\#126](https://github.com/appscode/guard/pull/126) ([tamalsaha](https://github.com/tamalsaha))
- Revert "Use image with tag 0.1.0 not canary" [\#125](https://github.com/appscode/guard/pull/125) ([tamalsaha](https://github.com/tamalsaha))
- Use image with tag 0.1.0 not canary [\#124](https://github.com/appscode/guard/pull/124) ([potsbo](https://github.com/potsbo))
- Guard Installation guide for Kubernetes Clusters built with Kubespray [\#123](https://github.com/appscode/guard/pull/123) ([vikas027](https://github.com/vikas027))
- Update docs to indicate node selector bug in Kubespray [\#122](https://github.com/appscode/guard/pull/122) ([tamalsaha](https://github.com/tamalsaha))
- fix error glog undefined [\#119](https://github.com/appscode/guard/pull/119) ([nightfury1204](https://github.com/nightfury1204))
- Generate flag methods using go-enum [\#118](https://github.com/appscode/guard/pull/118) ([tamalsaha](https://github.com/tamalsaha))
- Use github.com/golang/glog library [\#117](https://github.com/appscode/guard/pull/117) ([tamalsaha](https://github.com/tamalsaha))
- Add simple Grafana Dashboard for Guard [\#116](https://github.com/appscode/guard/pull/116) ([alexanderdavidsen](https://github.com/alexanderdavidsen))
- Documentation: how to create serviceMonitors for prometheus-operator [\#115](https://github.com/appscode/guard/pull/115) ([alexanderdavidsen](https://github.com/alexanderdavidsen))
- Add test for auth provider client [\#114](https://github.com/appscode/guard/pull/114) ([nightfury1204](https://github.com/nightfury1204))
- Fix metrics spelling [\#113](https://github.com/appscode/guard/pull/113) ([tamalsaha](https://github.com/tamalsaha))
- Update cli reference docs [\#112](https://github.com/appscode/guard/pull/112) ([tamalsaha](https://github.com/tamalsaha))
- Update GitLab documentation to clarify the usage of base-url [\#111](https://github.com/appscode/guard/pull/111) ([alexanderdavidsen](https://github.com/alexanderdavidsen))

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
- Support Azure Active Directory / Office365 [\#10](https://github.com/appscode/guard/issues/10)
- Fix installer [\#91](https://github.com/appscode/guard/issues/91)
- Kerberos [\#58](https://github.com/appscode/guard/issues/58)
- Is the google auth actually working/tested? [\#36](https://github.com/appscode/guard/issues/36)
- Document how to fix bad certificate issue [\#21](https://github.com/appscode/guard/issues/21)
- Fuzz server [\#14](https://github.com/appscode/guard/issues/14)

**Merged pull requests:**

- Support lowercase match for LDAP auth choice [\#110](https://github.com/appscode/guard/pull/110) ([tamalsaha](https://github.com/tamalsaha))
- Update changelog [\#108](https://github.com/appscode/guard/pull/108) ([tamalsaha](https://github.com/tamalsaha))
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
- Add kops documentation [\#43](https://github.com/appscode/guard/pull/43) ([thomaspeitz](https://github.com/thomaspeitz))
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