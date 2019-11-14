<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Change Log](#change-log)
  - [Unreleased](#unreleased)
  - [v0.32.1-beta.1 (2019-10-30)](#v0321-beta1-2019-10-30)
  - [v0.32.0-beta.1 (2019-10-20)](#v0320-beta1-2019-10-20)
  - [v0.31.0-beta.1 (2019-10-20)](#v0310-beta1-2019-10-20)
  - [v0.19.0-beta.1 (2019-09-23)](#v0190-beta1-2019-09-23)
  - [v0.18.0-beta.1 (2019-08-22)](#v0180-beta1-2019-08-22)
  - [v0.17.4-beta.1 (2019-08-09)](#v0174-beta1-2019-08-09)
  - [v0.17.3-beta.1 (2019-08-03)](#v0173-beta1-2019-08-03)
  - [v0.17.2-beta.1 (2019-08-02)](#v0172-beta1-2019-08-02)
  - [v0.17.1-beta.1 (2019-07-23)](#v0171-beta1-2019-07-23)
  - [v0.17.0-beta.1 (2019-07-18)](#v0170-beta1-2019-07-18)
  - [v0.16.0-beta.5 (2019-06-28)](#v0160-beta5-2019-06-28)
  - [v0.16.0-beta.4 (2019-05-28)](#v0160-beta4-2019-05-28)
  - [v0.16.0-beta.3 (2019-05-19)](#v0160-beta3-2019-05-19)
  - [v0.15.2 (2019-05-04)](#v0152-2019-05-04)
  - [v0.15.1 (2019-04-29)](#v0151-2019-04-29)
  - [v0.15.0 (2019-04-29)](#v0150-2019-04-29)
  - [v0.14.2+oryOS.10 (2018-12-13)](#v0142oryos10-2018-12-13)
  - [v0.14.1+oryOS.10 (2018-12-13)](#v0141oryos10-2018-12-13)
  - [v0.14.0+oryOS.10 (2018-12-13)](#v0140oryos10-2018-12-13)
  - [v0.13.9+oryOS.9 (2018-11-14)](#v0139oryos9-2018-11-14)
  - [v0.13.8+oryOS.8 (2018-11-14)](#v0138oryos8-2018-11-14)
  - [v0.13.7+oryOS.7 (2018-11-14)](#v0137oryos7-2018-11-14)
  - [v0.13.6+oryOS.6 (2018-11-14)](#v0136oryos6-2018-11-14)
  - [v0.13.5+oryOS.5 (2018-11-14)](#v0135oryos5-2018-11-14)
  - [v0.13.4+oryOS.4 (2018-11-14)](#v0134oryos4-2018-11-14)
  - [v0.13.3+oryOS.3 (2018-11-14)](#v0133oryos3-2018-11-14)
  - [v0.13.2+oryOS.2 (2018-11-14)](#v0132oryos2-2018-11-14)
  - [v0.13.1+oryOS.1 (2018-11-14)](#v0131oryos1-2018-11-14)
  - [v0.11.12 (2018-05-07)](#v01112-2018-05-07)
  - [v0.0.29 (2017-12-19)](#v0029-2017-12-19)
  - [v0.0.28 (2017-12-19)](#v0028-2017-12-19)
  - [v0.0.27 (2017-12-12)](#v0027-2017-12-12)
  - [v0.0.26 (2017-12-11)](#v0026-2017-12-11)
  - [v0.0.25 (2017-11-28)](#v0025-2017-11-28)
  - [v0.0.24 (2017-11-26)](#v0024-2017-11-26)
  - [v0.0.23 (2017-11-24)](#v0023-2017-11-24)
  - [v0.0.22 (2017-11-20)](#v0022-2017-11-20)
  - [v0.0.21 (2017-11-19)](#v0021-2017-11-19)
  - [v0.0.20 (2017-11-18)](#v0020-2017-11-18)
  - [v0.0.19 (2017-11-13)](#v0019-2017-11-13)
  - [v0.0.18 (2017-11-13)](#v0018-2017-11-13)
  - [v0.0.17 (2017-11-12)](#v0017-2017-11-12)
  - [v0.0.16 (2017-11-12)](#v0016-2017-11-12)
  - [v0.0.15 (2017-11-09)](#v0015-2017-11-09)
  - [v0.0.14 (2017-11-07)](#v0014-2017-11-07)
  - [v0.0.13 (2017-11-07)](#v0013-2017-11-07)
  - [v0.0.12 (2017-11-07)](#v0012-2017-11-07)
  - [v0.0.11 (2017-11-06)](#v0011-2017-11-06)
  - [v0.0.10 (2017-11-06)](#v0010-2017-11-06)
  - [v0.0.9 (2017-11-06)](#v009-2017-11-06)
  - [v0.0.8 (2017-11-06)](#v008-2017-11-06)
  - [v0.0.7 (2017-11-06)](#v007-2017-11-06)
  - [v0.0.6 (2017-11-03)](#v006-2017-11-03)
  - [v0.0.5 (2017-11-01)](#v005-2017-11-01)
  - [v0.0.4 (2017-10-21)](#v004-2017-10-21)
  - [v0.0.3 (2017-10-18)](#v003-2017-10-18)
  - [v0.0.2 (2017-10-12)](#v002-2017-10-12)
  - [v0.0.1 (2017-10-10)](#v001-2017-10-10)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Change Log

## [Unreleased](https://github.com/ory/oathkeeper/tree/HEAD)

[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.32.1-beta.1...HEAD)

**Implemented enhancements:**

- Extend JWT authenticator per rule config with global config [\#255](https://github.com/ory/oathkeeper/issues/255)

**Fixed bugs:**

- Missing Content-type [\#289](https://github.com/ory/oathkeeper/issues/289)
- Should not fatal when rule configmap changes [\#229](https://github.com/ory/oathkeeper/issues/229)

## [v0.32.1-beta.1](https://github.com/ory/oathkeeper/tree/v0.32.1-beta.1) (2019-10-30)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.32.0-beta.1...v0.32.1-beta.1)

**Implemented enhancements:**

- Remove the need for outbound internet connection from Oathkeeper [\#234](https://github.com/ory/oathkeeper/issues/234)

**Fixed bugs:**

- vendor: Update ory/x/viperx dependency [\#285](https://github.com/ory/oathkeeper/pull/285) ([aeneasr](https://github.com/aeneasr))

**Closed issues:**

- \[Helm chart\] Quick changes [\#278](https://github.com/ory/oathkeeper/issues/278)
- Env vars for jwks\_url in id\_token mutator not working in versions \>v0.19.0-beta.1 [\#276](https://github.com/ory/oathkeeper/issues/276)
- missing release assets in release v0.19.2-beta.1+oryOS.12 [\#275](https://github.com/ory/oathkeeper/issues/275)
- Env vars for client ID/secret in oauth2\_introspection don't work anymore in v0.19.0-beta.1 [\#270](https://github.com/ory/oathkeeper/issues/270)

**Merged pull requests:**

- authz: Add Content-Type header in the call to Keto [\#290](https://github.com/ory/oathkeeper/pull/290) ([Sbou](https://github.com/Sbou))
- Auto-kill test runner after 10 retries [\#286](https://github.com/ory/oathkeeper/pull/286) ([aeneasr](https://github.com/aeneasr))
- Dereference config schema and resolve issues [\#282](https://github.com/ory/oathkeeper/pull/282) ([aeneasr](https://github.com/aeneasr))

## [v0.32.0-beta.1](https://github.com/ory/oathkeeper/tree/v0.32.0-beta.1) (2019-10-20)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.31.0-beta.1...v0.32.0-beta.1)

## [v0.31.0-beta.1](https://github.com/ory/oathkeeper/tree/v0.31.0-beta.1) (2019-10-20)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.19.0-beta.1...v0.31.0-beta.1)

**Implemented enhancements:**

- Version access rules [\#266](https://github.com/ory/oathkeeper/issues/266)
- rule: Add migration capabilities [\#268](https://github.com/ory/oathkeeper/pull/268) ([aeneasr](https://github.com/aeneasr))

**Fixed bugs:**

- Client Credentials Authenticators not compatible with Hydra? [\#260](https://github.com/ory/oathkeeper/issues/260)
- "jwt" authenticator returns 403 instead of 401 [\#256](https://github.com/ory/oathkeeper/issues/256)

**Closed issues:**

- Access-rules conversion error. [\#274](https://github.com/ory/oathkeeper/issues/274)
- The configuration is invalid and could not be loaded. [\#273](https://github.com/ory/oathkeeper/issues/273)
- Update mutators in documentation [\#261](https://github.com/ory/oathkeeper/issues/261)
- Support fully both Oauth & JWT authenticator in access rule  [\#257](https://github.com/ory/oathkeeper/issues/257)

**Merged pull requests:**

- Support alternative token location [\#271](https://github.com/ory/oathkeeper/pull/271) ([kubadz](https://github.com/kubadz))
- authn: Force auth style in oauth2 client credentials authn [\#267](https://github.com/ory/oathkeeper/pull/267) ([aeneasr](https://github.com/aeneasr))
- fix \#256: change error code from 403 to 401 [\#259](https://github.com/ory/oathkeeper/pull/259) ([ngrigoriev](https://github.com/ngrigoriev))

## [v0.19.0-beta.1](https://github.com/ory/oathkeeper/tree/v0.19.0-beta.1) (2019-09-23)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.18.0-beta.1...v0.19.0-beta.1)

**Closed issues:**

- Keto engine doesn't build correctly the payload to call keto for URL with query parameters [\#250](https://github.com/ory/oathkeeper/issues/250)
- Mutator: unrecognized by oathkeeper \(v0.17.5\) [\#248](https://github.com/ory/oathkeeper/issues/248)
- Mutator issuing  JWT with custom claims [\#228](https://github.com/ory/oathkeeper/issues/228)

**Merged pull requests:**

- Resolve broken tests [\#262](https://github.com/ory/oathkeeper/pull/262) ([aeneasr](https://github.com/aeneasr))
- Homogenize configuration management [\#258](https://github.com/ory/oathkeeper/pull/258) ([aeneasr](https://github.com/aeneasr))
- Fix \#250: Ignore query parameters to build payload for Keto engine [\#251](https://github.com/ory/oathkeeper/pull/251) ([GuillaumeSmaha](https://github.com/GuillaumeSmaha))

## [v0.18.0-beta.1](https://github.com/ory/oathkeeper/tree/v0.18.0-beta.1) (2019-08-22)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.17.4-beta.1...v0.18.0-beta.1)

**Merged pull requests:**

- ID Token Custom Claims [\#246](https://github.com/ory/oathkeeper/pull/246) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#245](https://github.com/ory/oathkeeper/pull/245) ([aeneasr](https://github.com/aeneasr))
- Add mutator for modifying authenticationSession with external API [\#240](https://github.com/ory/oathkeeper/pull/240) ([kubadz](https://github.com/kubadz))
- docs: Updates issue and pull request templates [\#239](https://github.com/ory/oathkeeper/pull/239) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#238](https://github.com/ory/oathkeeper/pull/238) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#237](https://github.com/ory/oathkeeper/pull/237) ([aeneasr](https://github.com/aeneasr))
- doc: Add adopters placeholder [\#236](https://github.com/ory/oathkeeper/pull/236) ([aeneasr](https://github.com/aeneasr))

## [v0.17.4-beta.1](https://github.com/ory/oathkeeper/tree/v0.17.4-beta.1) (2019-08-09)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.17.3-beta.1...v0.17.4-beta.1)

**Merged pull requests:**

- Add sprig template library [\#235](https://github.com/ory/oathkeeper/pull/235) ([hypnoglow](https://github.com/hypnoglow))
- support multiple mutators [\#233](https://github.com/ory/oathkeeper/pull/233) ([jakkab](https://github.com/jakkab))
- docs: Updates issue and pull request templates [\#232](https://github.com/ory/oathkeeper/pull/232) ([aeneasr](https://github.com/aeneasr))

## [v0.17.3-beta.1](https://github.com/ory/oathkeeper/tree/v0.17.3-beta.1) (2019-08-03)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.17.2-beta.1...v0.17.3-beta.1)

**Fixed bugs:**

- rule: Resolve k8s configmap reload issue [\#231](https://github.com/ory/oathkeeper/pull/231) ([aeneasr](https://github.com/aeneasr))

## [v0.17.2-beta.1](https://github.com/ory/oathkeeper/tree/v0.17.2-beta.1) (2019-08-02)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.17.1-beta.1...v0.17.2-beta.1)

**Closed issues:**

- Panic on rolling update in Kubernetes [\#224](https://github.com/ory/oathkeeper/issues/224)
- Helm chart for oathkeeper [\#186](https://github.com/ory/oathkeeper/issues/186)

**Merged pull requests:**

- rules: Support kubernetes configmap reloading [\#230](https://github.com/ory/oathkeeper/pull/230) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#226](https://github.com/ory/oathkeeper/pull/226) ([aeneasr](https://github.com/aeneasr))

## [v0.17.1-beta.1](https://github.com/ory/oathkeeper/tree/v0.17.1-beta.1) (2019-07-23)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.17.0-beta.1...v0.17.1-beta.1)

**Merged pull requests:**

- Fix panic on send on closed channel [\#225](https://github.com/ory/oathkeeper/pull/225) ([hypnoglow](https://github.com/hypnoglow))

## [v0.17.0-beta.1](https://github.com/ory/oathkeeper/tree/v0.17.0-beta.1) (2019-07-18)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.16.0-beta.5...v0.17.0-beta.1)

**Implemented enhancements:**

- Add file watcher for access rules [\#216](https://github.com/ory/oathkeeper/issues/216)
- Add file watcher for config file [\#215](https://github.com/ory/oathkeeper/issues/215)

**Merged pull requests:**

- ci: Automate schema confiugration sync [\#222](https://github.com/ory/oathkeeper/pull/222) ([aeneasr](https://github.com/aeneasr))
- Validate Configuration with JSON Schema [\#220](https://github.com/ory/oathkeeper/pull/220) ([aeneasr](https://github.com/aeneasr))
- cmd: Do not fatal when immutable value is changed [\#218](https://github.com/ory/oathkeeper/pull/218) ([aeneasr](https://github.com/aeneasr))
- Watch configuration and access rule changes [\#217](https://github.com/ory/oathkeeper/pull/217) ([aeneasr](https://github.com/aeneasr))
- Add support for rules in YAML format [\#213](https://github.com/ory/oathkeeper/pull/213) ([hypnoglow](https://github.com/hypnoglow))

## [v0.16.0-beta.5](https://github.com/ory/oathkeeper/tree/v0.16.0-beta.5) (2019-06-28)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.16.0-beta.4...v0.16.0-beta.5)

**Closed issues:**

- Unable to build docker image on linux [\#207](https://github.com/ory/oathkeeper/issues/207)
- Always return 404 when used with Ambassador Auth Service [\#199](https://github.com/ory/oathkeeper/issues/199)

**Merged pull requests:**

- Add description into the name of subtest [\#212](https://github.com/ory/oathkeeper/pull/212) ([minchao](https://github.com/minchao))
- Add cookie session authenticator [\#211](https://github.com/ory/oathkeeper/pull/211) ([alexdavid](https://github.com/alexdavid))
- ci: Update golangci install script [\#210](https://github.com/ory/oathkeeper/pull/210) ([aeneasr](https://github.com/aeneasr))
- docker: Use non-root user in image [\#209](https://github.com/ory/oathkeeper/pull/209) ([aeneasr](https://github.com/aeneasr))
- Remove binary license [\#208](https://github.com/ory/oathkeeper/pull/208) ([aeneasr](https://github.com/aeneasr))
- Update config.yaml [\#204](https://github.com/ory/oathkeeper/pull/204) ([haf](https://github.com/haf))

## [v0.16.0-beta.4](https://github.com/ory/oathkeeper/tree/v0.16.0-beta.4) (2019-05-28)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.16.0-beta.3...v0.16.0-beta.4)

**Merged pull requests:**

- server: Properly declare negroni middleware [\#200](https://github.com/ory/oathkeeper/pull/200) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#198](https://github.com/ory/oathkeeper/pull/198) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#197](https://github.com/ory/oathkeeper/pull/197) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#196](https://github.com/ory/oathkeeper/pull/196) ([aeneasr](https://github.com/aeneasr))

## [v0.16.0-beta.3](https://github.com/ory/oathkeeper/tree/v0.16.0-beta.3) (2019-05-19)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.15.2...v0.16.0-beta.3)

**Implemented enhancements:**

- Clean up environment variables and throw errors on misconfiguration [\#140](https://github.com/ory/oathkeeper/issues/140)
- Missing serve all, both proxy/api using 4455 [\#122](https://github.com/ory/oathkeeper/issues/122)

**Closed issues:**

- json: cannot unmarshal string into Go value [\#183](https://github.com/ory/oathkeeper/issues/183)
- Oathkeeper \(v0.14.2\_oryOS.10\) returning empty reply on slow/long distance database calls [\#178](https://github.com/ory/oathkeeper/issues/178)
- Moving forward with ORY Oathkeeper [\#177](https://github.com/ory/oathkeeper/issues/177)
- Replace ORY Hydra JWK fetcher with local strategy and storage [\#174](https://github.com/ory/oathkeeper/issues/174)
- Support multiple JWKS URL in oathkeeper config rather than environment variable [\#168](https://github.com/ory/oathkeeper/issues/168)
- Move to new configuration management [\#164](https://github.com/ory/oathkeeper/issues/164)
- Do not disable filters, instead show decent error messages on misconfiguration [\#141](https://github.com/ory/oathkeeper/issues/141)
- make id\_token credential issuer optional [\#136](https://github.com/ory/oathkeeper/issues/136)

**Merged pull requests:**

- ci: Rename job release-docs to docs [\#193](https://github.com/ory/oathkeeper/pull/193) ([aeneasr](https://github.com/aeneasr))
- ci: Resolve goreleaser issues [\#192](https://github.com/ory/oathkeeper/pull/192) ([aeneasr](https://github.com/aeneasr))
- ci: Update release pipeline [\#191](https://github.com/ory/oathkeeper/pull/191) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#189](https://github.com/ory/oathkeeper/pull/189) ([aeneasr](https://github.com/aeneasr))
- install.sh: fix install script [\#187](https://github.com/ory/oathkeeper/pull/187) ([mkontani](https://github.com/mkontani))
- Reduce deployment complexity and refactor internals [\#185](https://github.com/ory/oathkeeper/pull/185) ([aeneasr](https://github.com/aeneasr))

## [v0.15.2](https://github.com/ory/oathkeeper/tree/v0.15.2) (2019-05-04)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.15.1...v0.15.2)

**Fixed bugs:**

- Credential issuer config is base64 encoded [\#182](https://github.com/ory/oathkeeper/issues/182)

**Merged pull requests:**

- Fix json encode of config for "credentials\_issuer" and "authorizer" during import [\#184](https://github.com/ory/oathkeeper/pull/184) ([stszap](https://github.com/stszap))

## [v0.15.1](https://github.com/ory/oathkeeper/tree/v0.15.1) (2019-04-29)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.15.0...v0.15.1)

**Merged pull requests:**

- vendor: Add go.sum [\#180](https://github.com/ory/oathkeeper/pull/180) ([aeneasr](https://github.com/aeneasr))

## [v0.15.0](https://github.com/ory/oathkeeper/tree/v0.15.0) (2019-04-29)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.14.2+oryOS.10...v0.15.0)

**Implemented enhancements:**

- Adopt new Keto SDK [\#172](https://github.com/ory/oathkeeper/issues/172)

**Fixed bugs:**

- Adopt new Keto SDK [\#172](https://github.com/ory/oathkeeper/issues/172)

**Closed issues:**

- Forward all auth\* headers in judge mode [\#166](https://github.com/ory/oathkeeper/issues/166)
- Move to go-swagger client [\#165](https://github.com/ory/oathkeeper/issues/165)
- Unable to install oathkeeper CLI [\#161](https://github.com/ory/oathkeeper/issues/161)
- Using Oathkeeper - External Consumer App [\#158](https://github.com/ory/oathkeeper/issues/158)
- Allow multiple rules for one URL [\#157](https://github.com/ory/oathkeeper/issues/157)
- CORS Not working as expected [\#151](https://github.com/ory/oathkeeper/issues/151)
- keto\_engine\_acp\_ory not working with oryOS10 [\#150](https://github.com/ory/oathkeeper/issues/150)
- Update README building-from-source part with the gomodule way [\#149](https://github.com/ory/oathkeeper/issues/149)
- required\_scope of authenticator validate only scope claim and not scp claim [\#138](https://github.com/ory/oathkeeper/issues/138)

**Merged pull requests:**

- docker: Remove full tag from build pipeline [\#179](https://github.com/ory/oathkeeper/pull/179) ([aeneasr](https://github.com/aeneasr))
-  sdk: Remove sdk dependencies to keto/hydra [\#173](https://github.com/ory/oathkeeper/pull/173) ([aeneasr](https://github.com/aeneasr))
- ci: Adopt new release pipeline [\#171](https://github.com/ory/oathkeeper/pull/171) ([aeneasr](https://github.com/aeneasr))
- sdk: Move to go-swagger SDK code generation [\#170](https://github.com/ory/oathkeeper/pull/170) ([aeneasr](https://github.com/aeneasr))
- judge: Set request headers for credential issuers [\#169](https://github.com/ory/oathkeeper/pull/169) ([aeneasr](https://github.com/aeneasr))
- Update dependencies [\#163](https://github.com/ory/oathkeeper/pull/163) ([aeneasr](https://github.com/aeneasr))
- proxy: Use scp,scope,scopes in jwt authenticator [\#162](https://github.com/ory/oathkeeper/pull/162) ([aeneasr](https://github.com/aeneasr))
- ci: Resolve CI build issue [\#160](https://github.com/ory/oathkeeper/pull/160) ([aeneasr](https://github.com/aeneasr))
- Ensure rule matcher is locked before updating [\#159](https://github.com/ory/oathkeeper/pull/159) ([jtescher](https://github.com/jtescher))
- proxy: improve debugability of JWT authenticator [\#156](https://github.com/ory/oathkeeper/pull/156) ([aeneasr](https://github.com/aeneasr))
- issue \#149 - Update README building-from-source part with the gomodul… [\#152](https://github.com/ory/oathkeeper/pull/152) ([pink-lucifer](https://github.com/pink-lucifer))

## [v0.14.2+oryOS.10](https://github.com/ory/oathkeeper/tree/v0.14.2+oryOS.10) (2018-12-13)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.14.1+oryOS.10...v0.14.2+oryOS.10)

**Merged pull requests:**

- ci: Fix docker push arguments in publish task [\#148](https://github.com/ory/oathkeeper/pull/148) ([aeneasr](https://github.com/aeneasr))

## [v0.14.1+oryOS.10](https://github.com/ory/oathkeeper/tree/v0.14.1+oryOS.10) (2018-12-13)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.14.0+oryOS.10...v0.14.1+oryOS.10)

**Merged pull requests:**

- ci: Fix docker release task [\#147](https://github.com/ory/oathkeeper/pull/147) ([aeneasr](https://github.com/aeneasr))

## [v0.14.0+oryOS.10](https://github.com/ory/oathkeeper/tree/v0.14.0+oryOS.10) (2018-12-13)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.13.9+oryOS.9...v0.14.0+oryOS.10)

**Closed issues:**

- Moving forward with this project's versioning [\#130](https://github.com/ory/oathkeeper/issues/130)
- Add OPA authorizer [\#98](https://github.com/ory/oathkeeper/issues/98)

**Merged pull requests:**

- vendor: Update keto to latest [\#146](https://github.com/ory/oathkeeper/pull/146) ([aeneasr](https://github.com/aeneasr))
- proxy: Update to recent keto changes  [\#145](https://github.com/ory/oathkeeper/pull/145) ([aeneasr](https://github.com/aeneasr))
- docs: Update documentation links [\#144](https://github.com/ory/oathkeeper/pull/144) ([aeneasr](https://github.com/aeneasr))
- docs: Align changelog, upgrade with new versions [\#143](https://github.com/ory/oathkeeper/pull/143) ([aeneasr](https://github.com/aeneasr))
- docs: Fix proxy help command description [\#142](https://github.com/ory/oathkeeper/pull/142) ([aeneasr](https://github.com/aeneasr))
- Ignore query parameters when matching url in rules. [\#139](https://github.com/ory/oathkeeper/pull/139) ([stszap](https://github.com/stszap))
- Support "scope" claim as a string in jwt authenticator [\#137](https://github.com/ory/oathkeeper/pull/137) ([stszap](https://github.com/stszap))

## [v0.13.9+oryOS.9](https://github.com/ory/oathkeeper/tree/v0.13.9+oryOS.9) (2018-11-14)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.13.8+oryOS.8...v0.13.9+oryOS.9)

## [v0.13.8+oryOS.8](https://github.com/ory/oathkeeper/tree/v0.13.8+oryOS.8) (2018-11-14)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.13.7+oryOS.7...v0.13.8+oryOS.8)

## [v0.13.7+oryOS.7](https://github.com/ory/oathkeeper/tree/v0.13.7+oryOS.7) (2018-11-14)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.13.6+oryOS.6...v0.13.7+oryOS.7)

## [v0.13.6+oryOS.6](https://github.com/ory/oathkeeper/tree/v0.13.6+oryOS.6) (2018-11-14)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.13.5+oryOS.5...v0.13.6+oryOS.6)

## [v0.13.5+oryOS.5](https://github.com/ory/oathkeeper/tree/v0.13.5+oryOS.5) (2018-11-14)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.13.4+oryOS.4...v0.13.5+oryOS.5)

## [v0.13.4+oryOS.4](https://github.com/ory/oathkeeper/tree/v0.13.4+oryOS.4) (2018-11-14)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.13.3+oryOS.3...v0.13.4+oryOS.4)

## [v0.13.3+oryOS.3](https://github.com/ory/oathkeeper/tree/v0.13.3+oryOS.3) (2018-11-14)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.13.2+oryOS.2...v0.13.3+oryOS.3)

## [v0.13.2+oryOS.2](https://github.com/ory/oathkeeper/tree/v0.13.2+oryOS.2) (2018-11-14)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.13.1+oryOS.1...v0.13.2+oryOS.2)

## [v0.13.1+oryOS.1](https://github.com/ory/oathkeeper/tree/v0.13.1+oryOS.1) (2018-11-14)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.11.12...v0.13.1+oryOS.1)

**Implemented enhancements:**

- Add JWT authenticator [\#112](https://github.com/ory/oathkeeper/issues/112)
- cmd: Should not fatal if ORY Hydra SDK is unable to start [\#71](https://github.com/ory/oathkeeper/issues/71)
- Slow POST through proxy causes timeout after 5 seconds [\#64](https://github.com/ory/oathkeeper/issues/64)
- proxy: Add JWT authenticator [\#109](https://github.com/ory/oathkeeper/pull/109) ([aeneasr](https://github.com/aeneasr))
- cmd: Disable cors per default [\#107](https://github.com/ory/oathkeeper/pull/107) ([aeneasr](https://github.com/aeneasr))
- Resolve various issues [\#93](https://github.com/ory/oathkeeper/pull/93) ([aeneasr](https://github.com/aeneasr))
- rule: Adds validator for rules [\#77](https://github.com/ory/oathkeeper/pull/77) ([aeneasr](https://github.com/aeneasr))

**Fixed bugs:**

- oathkeeper beta8 builds on older hydra SDK [\#101](https://github.com/ory/oathkeeper/issues/101)
- Invalid Url Validator [\#92](https://github.com/ory/oathkeeper/issues/92)
- Resolve stack overflow in key & rule refresher [\#80](https://github.com/ory/oathkeeper/issues/80)
- Deletion of conflicting rule doesn't solve the route conflict [\#73](https://github.com/ory/oathkeeper/issues/73)
- proxy: Improve compatibility with ORY Hydra 1.0.0-beta.8 [\#108](https://github.com/ory/oathkeeper/pull/108) ([aeneasr](https://github.com/aeneasr))
- cmd: Disable cors per default [\#107](https://github.com/ory/oathkeeper/pull/107) ([aeneasr](https://github.com/aeneasr))
- Resolve various issues [\#93](https://github.com/ory/oathkeeper/pull/93) ([aeneasr](https://github.com/aeneasr))
- rules: Properly handle conflicts on PUT and POST [\#76](https://github.com/ory/oathkeeper/pull/76) ([aeneasr](https://github.com/aeneasr))
- rules: Resolves an issue with cached matchers [\#75](https://github.com/ory/oathkeeper/pull/75) ([aeneasr](https://github.com/aeneasr))

**Closed issues:**

- Keto Warden Authorizer: Make Subject configurable. [\#128](https://github.com/ory/oathkeeper/issues/128)
- Inconsistent Environment Variable Docs [\#121](https://github.com/ory/oathkeeper/issues/121)
- --config flag doesn't work [\#110](https://github.com/ory/oathkeeper/issues/110)
- `noop` authenticator should not bypass allow/deny authorizers [\#97](https://github.com/ory/oathkeeper/issues/97)
- \[Proposal/Discussion\] New Credentials Issuers [\#96](https://github.com/ory/oathkeeper/issues/96)
- Build and upload binaries upon release [\#89](https://github.com/ory/oathkeeper/issues/89)
- Feature request: vault authenticator [\#88](https://github.com/ory/oathkeeper/issues/88)
- kid does not match .well-known/jwks.json [\#83](https://github.com/ory/oathkeeper/issues/83)
- MySQL not supported [\#82](https://github.com/ory/oathkeeper/issues/82)
- Make Oathkeeper work without Hydra \(Fix JWK Manager\) [\#65](https://github.com/ory/oathkeeper/issues/65)
- Expected at least one private key [\#61](https://github.com/ory/oathkeeper/issues/61)
- Disallow unknown JSON fields [\#45](https://github.com/ory/oathkeeper/issues/45)
- Write AWS Lambda function for oathkeeper [\#44](https://github.com/ory/oathkeeper/issues/44)
- Add endpoint for answering access requests directly [\#42](https://github.com/ory/oathkeeper/issues/42)
- Add input validator to rules [\#41](https://github.com/ory/oathkeeper/issues/41)
- PUT rules/unknownId does not error [\#38](https://github.com/ory/oathkeeper/issues/38)

**Merged pull requests:**

- docs: Improve some docs and update SDK [\#135](https://github.com/ory/oathkeeper/pull/135) ([aeneasr](https://github.com/aeneasr))
- Add environment parameters \(and description\) to configure proxy server timeout settings [\#132](https://github.com/ory/oathkeeper/pull/132) ([7phs](https://github.com/7phs))
- Make subject configurable using go template [\#129](https://github.com/ory/oathkeeper/pull/129) ([lsjostro](https://github.com/lsjostro))
- docs: Updates issue and pull request templates [\#127](https://github.com/ory/oathkeeper/pull/127) ([aeneasr](https://github.com/aeneasr))
- docs: Updates issue and pull request templates [\#126](https://github.com/ory/oathkeeper/pull/126) ([aeneasr](https://github.com/aeneasr))
- cmd: TLS environment variables [\#124](https://github.com/ory/oathkeeper/pull/124) ([fredbi](https://github.com/fredbi))
- docs: Fix typo in README. [\#118](https://github.com/ory/oathkeeper/pull/118) ([ddunkin](https://github.com/ddunkin))
- cmd: Properly document JWT refresh [\#117](https://github.com/ory/oathkeeper/pull/117) ([aeneasr](https://github.com/aeneasr))
- cmd: Enables TLS option on serve api [\#116](https://github.com/ory/oathkeeper/pull/116) ([fredbi](https://github.com/fredbi))
- Prepare beta.9 release [\#115](https://github.com/ory/oathkeeper/pull/115) ([aeneasr](https://github.com/aeneasr))
- Aligned TLS options with hydra: allow cert&key to be specified with file [\#114](https://github.com/ory/oathkeeper/pull/114) ([fredbi](https://github.com/fredbi))
- Improve integration tests [\#113](https://github.com/ory/oathkeeper/pull/113) ([aeneasr](https://github.com/aeneasr))
- cmd: Remove config flag [\#111](https://github.com/ory/oathkeeper/pull/111) ([aeneasr](https://github.com/aeneasr))
- \(fix\) Typo in checkResponse function print message [\#106](https://github.com/ory/oathkeeper/pull/106) ([devprincess](https://github.com/devprincess))
- proxy: add cookies ci to handler factory [\#103](https://github.com/ory/oathkeeper/pull/103) ([zikes](https://github.com/zikes))
- proxy: add cookies credentials issuer [\#102](https://github.com/ory/oathkeeper/pull/102) ([zikes](https://github.com/zikes))
- Headers Credentials Issuer [\#100](https://github.com/ory/oathkeeper/pull/100) ([zikes](https://github.com/zikes))
- Resolve various issues [\#99](https://github.com/ory/oathkeeper/pull/99) ([aeneasr](https://github.com/aeneasr))
- Node sdk [\#94](https://github.com/ory/oathkeeper/pull/94) ([aeneasr](https://github.com/aeneasr))
- judge: Add endpoint for answering access requests directly [\#91](https://github.com/ory/oathkeeper/pull/91) ([aeneasr](https://github.com/aeneasr))
- health: Introduce health and version endpoint [\#90](https://github.com/ory/oathkeeper/pull/90) ([aeneasr](https://github.com/aeneasr))
- docs: fix broken link [\#87](https://github.com/ory/oathkeeper/pull/87) ([orisano](https://github.com/orisano))
- README: grammatical fix in stability sentence [\#86](https://github.com/ory/oathkeeper/pull/86) ([philips](https://github.com/philips))
- rsakey: Resolve HS256 kid mismatch [\#85](https://github.com/ory/oathkeeper/pull/85) ([aeneasr](https://github.com/aeneasr))
- cmd: Allows connectivity to MySQL  [\#84](https://github.com/ory/oathkeeper/pull/84) ([aeneasr](https://github.com/aeneasr))
- cmd: Resolves recursive stack overflow [\#81](https://github.com/ory/oathkeeper/pull/81) ([aeneasr](https://github.com/aeneasr))
- docs: Adds link to examples repository [\#79](https://github.com/ory/oathkeeper/pull/79) ([aeneasr](https://github.com/aeneasr))
- docs: Adds gh templates & code of conduct [\#78](https://github.com/ory/oathkeeper/pull/78) ([aeneasr](https://github.com/aeneasr))
- ci: Prevent pushes from forks to coveralls [\#74](https://github.com/ory/oathkeeper/pull/74) ([aeneasr](https://github.com/aeneasr))
- Reduces setup complexity [\#72](https://github.com/ory/oathkeeper/pull/72) ([aeneasr](https://github.com/aeneasr))
- proxy: Resolves potential panic in request handler [\#70](https://github.com/ory/oathkeeper/pull/70) ([aeneasr](https://github.com/aeneasr))
- Minor improvements [\#69](https://github.com/ory/oathkeeper/pull/69) ([aeneasr](https://github.com/aeneasr))
- rsakey: Resolves issues with broken tests [\#68](https://github.com/ory/oathkeeper/pull/68) ([aeneasr](https://github.com/aeneasr))
- cmd: Improves cors parsing [\#67](https://github.com/ory/oathkeeper/pull/67) ([aeneasr](https://github.com/aeneasr))
- cmd: Doesn't fatal if no ORY Hydra is unresponsive. [\#66](https://github.com/ory/oathkeeper/pull/66) ([aeneasr](https://github.com/aeneasr))
- Keto [\#60](https://github.com/ory/oathkeeper/pull/60) ([aeneasr](https://github.com/aeneasr))

## [v0.11.12](https://github.com/ory/oathkeeper/tree/v0.11.12) (2018-05-07)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.29...v0.11.12)

**Closed issues:**

- Unable to refresh RSA keys for JWK signing [\#53](https://github.com/ory/oathkeeper/issues/53)
- Add well known endpoint to swagger docs [\#47](https://github.com/ory/oathkeeper/issues/47)

**Merged pull requests:**

- Update README.md [\#58](https://github.com/ory/oathkeeper/pull/58) ([aeneasr](https://github.com/aeneasr))
- docs: Moves documentation to new repository [\#57](https://github.com/ory/oathkeeper/pull/57) ([aeneasr](https://github.com/aeneasr))
- Update 2-EXECUTION.md [\#56](https://github.com/ory/oathkeeper/pull/56) ([maryoush](https://github.com/maryoush))
- Update 2-EXECUTION.md [\#55](https://github.com/ory/oathkeeper/pull/55) ([taland](https://github.com/taland))
- Improve tests [\#54](https://github.com/ory/oathkeeper/pull/54) ([aeneasr](https://github.com/aeneasr))
- cmd: correct logging typo [\#52](https://github.com/ory/oathkeeper/pull/52) ([euank](https://github.com/euank))
- docs: Adds license note to all source files [\#51](https://github.com/ory/oathkeeper/pull/51) ([aeneasr](https://github.com/aeneasr))
- ci: Resolves issue with pushing docs [\#50](https://github.com/ory/oathkeeper/pull/50) ([aeneasr](https://github.com/aeneasr))
- docs: Adds automatic summary generation [\#49](https://github.com/ory/oathkeeper/pull/49) ([aeneasr](https://github.com/aeneasr))

## [v0.0.29](https://github.com/ory/oathkeeper/tree/v0.0.29) (2017-12-19)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.28...v0.0.29)

**Merged pull requests:**

- Adds use field to well known [\#48](https://github.com/ory/oathkeeper/pull/48) ([aeneasr](https://github.com/aeneasr))

## [v0.0.28](https://github.com/ory/oathkeeper/tree/v0.0.28) (2017-12-19)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.27...v0.0.28)

**Closed issues:**

- Make key discovery easier with well-known feature [\#43](https://github.com/ory/oathkeeper/issues/43)

**Merged pull requests:**

- Replaces key discovery with well-known feature [\#46](https://github.com/ory/oathkeeper/pull/46) ([aeneasr](https://github.com/aeneasr))

## [v0.0.27](https://github.com/ory/oathkeeper/tree/v0.0.27) (2017-12-12)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.26...v0.0.27)

**Merged pull requests:**

- Adds cors capabilities to management server [\#40](https://github.com/ory/oathkeeper/pull/40) ([aeneasr](https://github.com/aeneasr))

## [v0.0.26](https://github.com/ory/oathkeeper/tree/v0.0.26) (2017-12-11)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.25...v0.0.26)

**Merged pull requests:**

- Fixes broken image link in docs [\#39](https://github.com/ory/oathkeeper/pull/39) ([aeneasr](https://github.com/aeneasr))

## [v0.0.25](https://github.com/ory/oathkeeper/tree/v0.0.25) (2017-11-28)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.24...v0.0.25)

**Merged pull requests:**

- Add extra data from token introspection to session [\#37](https://github.com/ory/oathkeeper/pull/37) ([aeneasr](https://github.com/aeneasr))

## [v0.0.24](https://github.com/ory/oathkeeper/tree/v0.0.24) (2017-11-26)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.23...v0.0.24)

**Closed issues:**

- Document HYDRA\_JWK\_SET\_ID [\#34](https://github.com/ory/oathkeeper/issues/34)
- Investigate if the issuer should be oathkeeper or hydra [\#27](https://github.com/ory/oathkeeper/issues/27)

**Merged pull requests:**

- Telemetry [\#36](https://github.com/ory/oathkeeper/pull/36) ([aeneasr](https://github.com/aeneasr))

## [v0.0.23](https://github.com/ory/oathkeeper/tree/v0.0.23) (2017-11-24)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.22...v0.0.23)

**Closed issues:**

- Rename basicAuthorizationModeEnabled to something that does not clash with HTTP basic authorization [\#29](https://github.com/ory/oathkeeper/issues/29)
- Rename bypass values for better clarity [\#13](https://github.com/ory/oathkeeper/issues/13)

**Merged pull requests:**

- Print formatted output string in rule management CLI [\#35](https://github.com/ory/oathkeeper/pull/35) ([aeneasr](https://github.com/aeneasr))
- docs: Add JWK set docs [\#33](https://github.com/ory/oathkeeper/pull/33) ([aeneasr](https://github.com/aeneasr))
- Update docs and add tests [\#32](https://github.com/ory/oathkeeper/pull/32) ([aeneasr](https://github.com/aeneasr))

## [v0.0.22](https://github.com/ory/oathkeeper/tree/v0.0.22) (2017-11-20)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.21...v0.0.22)

**Merged pull requests:**

- Renames bypass values for better clarity [\#31](https://github.com/ory/oathkeeper/pull/31) ([aeneasr](https://github.com/aeneasr))

## [v0.0.21](https://github.com/ory/oathkeeper/tree/v0.0.21) (2017-11-19)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.20...v0.0.21)

**Merged pull requests:**

- Request hydra.keys scope and fix panic [\#30](https://github.com/ory/oathkeeper/pull/30) ([aeneasr](https://github.com/aeneasr))

## [v0.0.20](https://github.com/ory/oathkeeper/tree/v0.0.20) (2017-11-18)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.19...v0.0.20)

**Merged pull requests:**

- docs: Improve swagger documentation [\#28](https://github.com/ory/oathkeeper/pull/28) ([aeneasr](https://github.com/aeneasr))
- cmd: Add rules management capabilities to the cli [\#26](https://github.com/ory/oathkeeper/pull/26) ([aeneasr](https://github.com/aeneasr))
- unstaged [\#25](https://github.com/ory/oathkeeper/pull/25) ([aeneasr](https://github.com/aeneasr))

## [v0.0.19](https://github.com/ory/oathkeeper/tree/v0.0.19) (2017-11-13)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.18...v0.0.19)

**Closed issues:**

- evaluator: token\[:5\] will cause panic [\#22](https://github.com/ory/oathkeeper/issues/22)

**Merged pull requests:**

- evaluator: Use full request URL [\#24](https://github.com/ory/oathkeeper/pull/24) ([aeneasr](https://github.com/aeneasr))

## [v0.0.18](https://github.com/ory/oathkeeper/tree/v0.0.18) (2017-11-13)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.17...v0.0.18)

**Merged pull requests:**

- evaluator: Resolve potential panic in token id generation [\#23](https://github.com/ory/oathkeeper/pull/23) ([aeneasr](https://github.com/aeneasr))

## [v0.0.17](https://github.com/ory/oathkeeper/tree/v0.0.17) (2017-11-12)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.16...v0.0.17)

**Merged pull requests:**

- Introduces surrogate\_id to SQLManager [\#21](https://github.com/ory/oathkeeper/pull/21) ([aeneasr](https://github.com/aeneasr))

## [v0.0.16](https://github.com/ory/oathkeeper/tree/v0.0.16) (2017-11-12)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.15...v0.0.16)

**Merged pull requests:**

- Replace MatchesPath with MatchesURL [\#20](https://github.com/ory/oathkeeper/pull/20) ([aeneasr](https://github.com/aeneasr))

## [v0.0.15](https://github.com/ory/oathkeeper/tree/v0.0.15) (2017-11-09)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.14...v0.0.15)

**Merged pull requests:**

- Add HTTPS capabilities and document proxy/management commands [\#19](https://github.com/ory/oathkeeper/pull/19) ([aeneasr](https://github.com/aeneasr))

## [v0.0.14](https://github.com/ory/oathkeeper/tree/v0.0.14) (2017-11-07)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.13...v0.0.14)

**Merged pull requests:**

- Make refresh\_delay configurable and skip it on boot [\#18](https://github.com/ory/oathkeeper/pull/18) ([aeneasr](https://github.com/aeneasr))

## [v0.0.13](https://github.com/ory/oathkeeper/tree/v0.0.13) (2017-11-07)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.12...v0.0.13)

**Merged pull requests:**

- Store rules path match in plaintext [\#17](https://github.com/ory/oathkeeper/pull/17) ([aeneasr](https://github.com/aeneasr))

## [v0.0.12](https://github.com/ory/oathkeeper/tree/v0.0.12) (2017-11-07)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.11...v0.0.12)

**Merged pull requests:**

- Use ladon regex compiler for matches [\#16](https://github.com/ory/oathkeeper/pull/16) ([aeneasr](https://github.com/aeneasr))

## [v0.0.11](https://github.com/ory/oathkeeper/tree/v0.0.11) (2017-11-06)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.10...v0.0.11)

## [v0.0.10](https://github.com/ory/oathkeeper/tree/v0.0.10) (2017-11-06)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.9...v0.0.10)

## [v0.0.9](https://github.com/ory/oathkeeper/tree/v0.0.9) (2017-11-06)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.8...v0.0.9)

## [v0.0.8](https://github.com/ory/oathkeeper/tree/v0.0.8) (2017-11-06)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.7...v0.0.8)

**Merged pull requests:**

- Make oathkeeper binary executable\# [\#15](https://github.com/ory/oathkeeper/pull/15) ([aeneasr](https://github.com/aeneasr))

## [v0.0.7](https://github.com/ory/oathkeeper/tree/v0.0.7) (2017-11-06)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.6...v0.0.7)

**Merged pull requests:**

- Build oathekeeper docker image statically [\#14](https://github.com/ory/oathkeeper/pull/14) ([aeneasr](https://github.com/aeneasr))

## [v0.0.6](https://github.com/ory/oathkeeper/tree/v0.0.6) (2017-11-03)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.5...v0.0.6)

**Merged pull requests:**

- Added serve all command [\#12](https://github.com/ory/oathkeeper/pull/12) ([aeneasr](https://github.com/aeneasr))

## [v0.0.5](https://github.com/ory/oathkeeper/tree/v0.0.5) (2017-11-01)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.4...v0.0.5)

**Merged pull requests:**

- Add cors handling to proxy [\#11](https://github.com/ory/oathkeeper/pull/11) ([aeneasr](https://github.com/aeneasr))
- Remove goveralls from circle build [\#10](https://github.com/ory/oathkeeper/pull/10) ([aeneasr](https://github.com/aeneasr))
- Use circle ci build status badge [\#9](https://github.com/ory/oathkeeper/pull/9) ([aeneasr](https://github.com/aeneasr))
- Switch from glide to golang/dep for vendoring [\#8](https://github.com/ory/oathkeeper/pull/8) ([aeneasr](https://github.com/aeneasr))
- Resolve tests by replacing nil slice [\#7](https://github.com/ory/oathkeeper/pull/7) ([aeneasr](https://github.com/aeneasr))
- Add circleci configuration file [\#5](https://github.com/ory/oathkeeper/pull/5) ([aeneasr](https://github.com/aeneasr))

## [v0.0.4](https://github.com/ory/oathkeeper/tree/v0.0.4) (2017-10-21)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.3...v0.0.4)

**Merged pull requests:**

- Return arrays instead of null on rule creation [\#6](https://github.com/ory/oathkeeper/pull/6) ([aeneasr](https://github.com/aeneasr))

## [v0.0.3](https://github.com/ory/oathkeeper/tree/v0.0.3) (2017-10-18)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.2...v0.0.3)

**Merged pull requests:**

- Fix unauthorized [\#4](https://github.com/ory/oathkeeper/pull/4) ([aeneasr](https://github.com/aeneasr))

## [v0.0.2](https://github.com/ory/oathkeeper/tree/v0.0.2) (2017-10-12)
[Full Changelog](https://github.com/ory/oathkeeper/compare/v0.0.1...v0.0.2)

**Merged pull requests:**

- Skip acp checks [\#3](https://github.com/ory/oathkeeper/pull/3) ([aeneasr](https://github.com/aeneasr))

## [v0.0.1](https://github.com/ory/oathkeeper/tree/v0.0.1) (2017-10-10)
**Merged pull requests:**

- travis: add goveralls report submission [\#2](https://github.com/ory/oathkeeper/pull/2) ([aeneasr](https://github.com/aeneasr))
- Prototype [\#1](https://github.com/ory/oathkeeper/pull/1) ([aeneasr](https://github.com/aeneasr))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*