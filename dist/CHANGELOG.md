This release introduces caching capabilities for the OAuth2 Client Credentials authenticator as well as compatibility with Traefik!

### Bug Fixes

- Add pre-steps with packr2 ([#921](https://github.com/ory/oathkeeper/issues/921)) ([d53ef01](https://github.com/ory/oathkeeper/commit/d53ef0123830060cec73d425fc9b3f7e93ada66d)), closes [#920](https://github.com/ory/oathkeeper/issues/920)
- Bump goreleaser orb ([#919](https://github.com/ory/oathkeeper/issues/919)) ([f8dcda2](https://github.com/ory/oathkeeper/commit/f8dcda26cca0489248739cbcb4133b959d4991fe))
- Use all pre-hooks ([09be55f](https://github.com/ory/oathkeeper/commit/09be55feddffc8ed483258ce3e250fc57528054f))

### Code Generation

- Pin v0.38.20-beta.1 release commit ([410d69e](https://github.com/ory/oathkeeper/commit/410d69edfca4cc3a83c1d819d648709ba438e74a))

### Code Refactoring

- Move docs to ory/docs ([a0c6927](https://github.com/ory/oathkeeper/commit/a0c69275fb6e768cfd07e4d467155f4cf95ebbb8))

### Documentation

- Recover sidebar ([165224f](https://github.com/ory/oathkeeper/commit/165224fdf6636d55b9fb71c81da9b13426b201f6))

### Features

- Add post-release step ([e7fd550](https://github.com/ory/oathkeeper/commit/e7fd55030b9408e863f497deeb3e8f1bf66a9855))
- Introduce token caching for client credentials authentication ([#922](https://github.com/ory/oathkeeper/issues/922)) ([9a56154](https://github.com/ory/oathkeeper/commit/9a56154161429f9080ed6204e61aaf3a1ab731a1)), closes [#870](https://github.com/ory/oathkeeper/issues/870):

  Right now every request via Oathkeeper that uses client credentials
  authentication requests a new access token. This can introduce a lot
  of latency in the critical path of an application in case of a slow
  token endpoint.

  This change introduces a cache similar to the one that is used in the
  introspection authentication.

- Migrate to openapi 3.0 generation ([190d1a7](https://github.com/ory/oathkeeper/commit/190d1a7d1319f216ca3c9e9289d5282733ecc88c))
- Traefik decision api support ([#904](https://github.com/ory/oathkeeper/issues/904)) ([bfde9df](https://github.com/ory/oathkeeper/commit/bfde9dfc6ef71762ab25289a0afbe6793899f312)), closes [#521](https://github.com/ory/oathkeeper/issues/521) [#441](https://github.com/ory/oathkeeper/issues/441) [#487](https://github.com/ory/oathkeeper/issues/487) [#263](https://github.com/ory/oathkeeper/issues/263):

  Closes https://github.com/ory/oathkeeper/discussions/899


## Changelog
* 8579000 autogen(docs): generate and format documentation
* 71e69ef autogen(docs): regenerate and update changelog
* a3b5b28 autogen(docs): regenerate and update changelog
* 31fe9b7 autogen(docs): regenerate and update changelog
* cb01565 autogen(docs): regenerate and update changelog
* 3fea697 autogen(openapi): Regenerate openapi spec and internal client
* 84c15a6 autogen(openapi): Regenerate openapi spec and internal client
* 83d6728 autogen: add v0.38.19-beta.1 to version.schema.json
* 410d69e autogen: pin v0.38.20-beta.1 release commit
* 33b0c63 autogen: pin v0.38.20-beta.1.pre.0 release commit
* 06bc33f autogen: update release artifacts
* bd1b03a autogen: update release artifacts
* 2cd6282 chore: bump sprig version (#917)
* f8f82c4 chore: update repository templates
* 5d3e1bf chore: update repository templates
* 3c8b49e ci: add next cli docs generator
* 729fadc ci: remove docs/build from cci
* 962f57e ci: update cli location and fix generation script
* 165224f docs: recover sidebar
* bfde9df feat: Traefik decision api support (#904)
* e7fd550 feat: add post-release step
* 9a56154 feat: introduce token caching for client credentials authentication (#922)
* 190d1a7 feat: migrate to openapi 3.0 generation
* d53ef01 fix: add pre-steps with packr2 (#921)
* f8dcda2 fix: bump goreleaser orb (#919)
* 09be55f fix: use all pre-hooks
* a0c6927 refactor: move docs to ory/docs
