This release adds CVE scanners for Docker Images and updates several dependencies to resolve CVE issues.

Additionally, support for various tracers has been added, patches to caching and JWT audiences have been made, and more configuration options have been added for various rules.

### Bug Fixes

- Add config schema for tracing for jaeger ([#830](https://github.com/ory/oathkeeper/issues/830)) ([59871fc](https://github.com/ory/oathkeeper/commit/59871fca6984d221051e837eb768894c4c48ee27))
- Add hiring notice to README ([#884](https://github.com/ory/oathkeeper/issues/884)) ([9dea379](https://github.com/ory/oathkeeper/commit/9dea379a12abed4ceb84067d054d28032a50c783))
- Add ory cli ([df8a19b](https://github.com/ory/oathkeeper/commit/df8a19bd9adad664beddb017073c77a9e82b37af))
- Allow forwarding query parameters to the session store ([#817](https://github.com/ory/oathkeeper/issues/817)) ([9375f92](https://github.com/ory/oathkeeper/commit/9375f92b5d647c8417389158bf66e060b4ab8ad6)), closes [#786](https://github.com/ory/oathkeeper/issues/786) [#786](https://github.com/ory/oathkeeper/issues/786)
- Building docker image for docker-compose ([#889](https://github.com/ory/oathkeeper/issues/889)) ([adf0d1b](https://github.com/ory/oathkeeper/commit/adf0d1baaf466cafdc72cba3818867545a91e0b1))
- Remote_json default configuration ([#880](https://github.com/ory/oathkeeper/issues/880)) ([18788d1](https://github.com/ory/oathkeeper/commit/18788d1393c041c97d89812366f899ed359c67cf)), closes [#797](https://github.com/ory/oathkeeper/issues/797)
- Use NYT capitalistaion for all Swagger headlines ([#859](https://github.com/ory/oathkeeper/issues/859)) ([8c2da46](https://github.com/ory/oathkeeper/commit/8c2da466edb0e72a4bcb4c854bf80b6a98e3ac7a)), closes [#503](https://github.com/ory/oathkeeper/issues/503):

  Capitalised all the Swagger headlines for files found in /api.

### Code Generation

- Pin v0.38.17-beta.1 release commit ([f16db10](https://github.com/ory/oathkeeper/commit/f16db102eae5fc8ebf94af4fc9bca6387d0d41fa))

### Documentation

- Update authz.md ([#879](https://github.com/ory/oathkeeper/issues/879)) ([b6b5824](https://github.com/ory/oathkeeper/commit/b6b58249aec358d903bee18acc23836fe77b3860))
- Use correct casing ([58b1d43](https://github.com/ory/oathkeeper/commit/58b1d43dd99ebceea22980d5debefdbcc0a4f3c7)), closes [#900](https://github.com/ory/oathkeeper/issues/900)
- Warn that gzip is unsupported ([#835](https://github.com/ory/oathkeeper/issues/835)) ([78e612e](https://github.com/ory/oathkeeper/commit/78e612eeeba20c3ce1f5ff32c8dde0a9b6534eb7)):

  Note to users that gzip responses are as of now unsupported for Cookie and Bearer authenticators.
  The result is that the `subject` and `extra` will not be filled in, and will fail silently.

### Features

- Add retry and timeout support in authorizers ([#883](https://github.com/ory/oathkeeper/issues/883)) ([ec926b0](https://github.com/ory/oathkeeper/commit/ec926b09908e51fe6f4819e281beaf639a22eb69)):

  Adds the ability to define HTTP timeouts for authorizers.

- Add support for X-Forwarded-Proto header ([#665](https://github.com/ory/oathkeeper/issues/665)) ([a8c9354](https://github.com/ory/oathkeeper/commit/a8c9354acd64b097492c9dae9df092fecb1b310e)), closes [#153](https://github.com/ory/oathkeeper/issues/153)
- Allow both string and []string in aud field ([#822](https://github.com/ory/oathkeeper/issues/822)) ([1897f31](https://github.com/ory/oathkeeper/commit/1897f318c522ce3d5698e5cca234ab170bf10596)), closes [#491](https://github.com/ory/oathkeeper/issues/491) [#601](https://github.com/ory/oathkeeper/issues/601) [#792](https://github.com/ory/oathkeeper/issues/792) [#810](https://github.com/ory/oathkeeper/issues/810)
- Introduce cve scanning ([#839](https://github.com/ory/oathkeeper/issues/839)) ([1432e2c](https://github.com/ory/oathkeeper/commit/1432e2cbbd53d86133307d23ec5b85dc032e00fd))
- **jwt:** Replace jwt module ([#818](https://github.com/ory/oathkeeper/issues/818)) ([301b673](https://github.com/ory/oathkeeper/commit/301b673483b7af59dd0f38148edd12da22c67a6c))
- Store oauth2 introspection result as bytes in cache ([#811](https://github.com/ory/oathkeeper/issues/811)) ([5645605](https://github.com/ory/oathkeeper/commit/56456056909d19c04353347e9543e9dce73edfca))
- Support Zipkin tracer ([#832](https://github.com/ory/oathkeeper/issues/832)) ([2f2552d](https://github.com/ory/oathkeeper/commit/2f2552dc2769673c0f397dfec6022eb9395476ee))

### Unclassified

- docs: declare s3, gs, and azblob access rule repositories in config schema (#829) ([e2433f6](https://github.com/ory/oathkeeper/commit/e2433f6318eb77cf4e870d26f90a0d44a8f93d2e)), closes [#829](https://github.com/ory/oathkeeper/issues/829)


## Changelog
* 1f1f03a autogen(docs): regenerate and update changelog
* 0725820 autogen(docs): regenerate and update changelog
* 6cb417c autogen(docs): regenerate and update changelog
* 83cb5c0 autogen(docs): regenerate and update changelog
* 0dcd1f5 autogen(docs): regenerate and update changelog
* 38dfbcc autogen(docs): regenerate and update changelog
* c89737b autogen(docs): regenerate and update changelog
* 08324dd autogen(docs): regenerate and update changelog
* 9636c96 autogen(docs): regenerate and update changelog
* 667aeed autogen(docs): regenerate and update changelog
* 057293f autogen(docs): regenerate and update changelog
* e807863 autogen(docs): regenerate and update changelog
* b131d94 autogen(docs): regenerate and update changelog
* 255ad15 autogen(docs): regenerate and update changelog
* 168086e autogen(docs): regenerate and update changelog
* 317f874 autogen(docs): regenerate and update changelog
* 133e8a5 autogen(docs): regenerate and update changelog
* be93f1e autogen(docs): regenerate and update changelog
* 8a51d52 autogen(docs): update milestone document
* 7504e1e autogen(docs): update milestone document
* e785140 autogen(docs): update milestone document
* 19f2c68 autogen(docs): update milestone document
* 511d4b7 autogen(docs): update milestone document
* 9910160 autogen(openapi): Regenerate swagger spec and internal client
* cf63dc5 autogen(openapi): Regenerate swagger spec and internal client
* 8db79c9 autogen: add v0.38.15-beta.1 to version.schema.json
* 737320f autogen: pin v0.38.16-beta.1 release commit
* f16db10 autogen: pin v0.38.17-beta.1 release commit
* 5cc648e chore(deps): bump github.com/tidwall/gjson from 1.6.7 to 1.9.3 (#873)
* 65e53b6 chore: bump alpine version in dockerfiles (#837)
* 9b41eed chore: remove old sdk generator (#842)
* e49dbbd chore: update docusaurus template
* 2d359d9 chore: update docusaurus template
* a686910 chore: update docusaurus template
* 3f4c2ed chore: update docusaurus template
* 23e624d chore: update docusaurus template (#820)
* 1f64342 chore: update docusaurus template (#821)
* 9ca90e3 chore: update docusaurus template (#840)
* 002a2a8 chore: update docusaurus template (#847)
* 14dd31a chore: update docusaurus template (#866)
* 1564e0c chore: update docusaurus template (#872)
* 3381b6c chore: update docusaurus template (#875)
* 2980573 chore: update docusaurus template (#891)
* 9f29fc4 chore: update repository templates
* da516f5 chore: update repository templates
* 1553c14 chore: update repository templates
* 9f6644a chore: update repository templates
* 62ebb22 chore: update repository templates
* bc70566 chore: update repository templates
* ee210a3 chore: update repository templates
* 9c80149 chore: update repository templates (#823)
* be72846 chore: update repository templates (#825)
* 80bc079 chore: update repository templates (#827)
* 1da447d chore: update repository templates (#857)
* 8f23209 chore: update repository templates (#858)
* 497cd3c chore: update repository templates (#863)
* 7cd7bca chore: update repository templates (#864)
* ade680b chore: update repository templates to 8191b78131173cce8788143f6ad95119d9b813c5
* b1e772e ci: bump goreleaser (#816)
* 38d0883 ci: bump orbs (#815)
* 30ff27f ci: resolve regression issues (#881)
* e2433f6 docs: declare s3, gs, and azblob access rule repositories in config schema (#829)
* b6b5824 docs: update authz.md (#879)
* 58b1d43 docs: use correct casing
* 78e612e docs: warn that gzip is unsupported (#835)
* 301b673 feat(jwt): replace jwt module (#818)
* ec926b0 feat: add retry and timeout support in authorizers (#883)
* a8c9354 feat: add support for X-Forwarded-Proto header (#665)
* 1897f31 feat: allow both string and []string in aud field (#822)
* 1432e2c feat: introduce cve scanning (#839)
* 5645605 feat: store oauth2 introspection result as bytes in cache (#811)
* 2f2552d feat: support Zipkin tracer (#832)
* 59871fc fix: add config schema for tracing for jaeger (#830)
* 9dea379 fix: add hiring notice to README (#884)
* df8a19b fix: add ory cli
* 9375f92 fix: allow forwarding query parameters to the session store (#817)
* adf0d1b fix: building docker image for docker-compose (#889)
* 18788d1 fix: remote_json default configuration (#880)
* 8c2da46 fix: use NYT capitalistaion for all Swagger headlines (#859)
