This release adds support for rewriting the HTTP method in certain authenticators.

### Bug Fixes

- Bump Ory CLI ([5c03d4f](https://github.com/ory/oathkeeper/commit/5c03d4f0b8e1868fe6b1a30396f8411093d9c797))
- Update cve scanners ([#905](https://github.com/ory/oathkeeper/issues/905)) ([57c38c0](https://github.com/ory/oathkeeper/commit/57c38c0d4e75658373daaf3f6a80e22efd4dc3d5))

### Code Generation

- Pin v0.38.19-beta.1 release commit ([dedb92c](https://github.com/ory/oathkeeper/commit/dedb92cd98311bcf94b39c846b3d769aa63476f2))

### Documentation

- Fix "decisions" typo in Introduction ([#907](https://github.com/ory/oathkeeper/issues/907)) ([db346d5](https://github.com/ory/oathkeeper/commit/db346d5e3cae966f609f6bae38958c5d00970abe))

### Features

- Allow overriding HTTP method for upstream calls ([69c64e7](https://github.com/ory/oathkeeper/commit/69c64e79eb7eb5ad415503c8f71a424f8da90f10)):

  This patch adds new configuration `force_method` to the bearer token and cookie session authenticators. It allows overriding the HTTP method for upstream calls.


## Changelog
* 5ee5b44 autogen(docs): generate and format documentation
* a6c6cf3 autogen(docs): generate and format documentation
* 2ff93eb autogen(docs): generate and format documentation
* bc655dd autogen(openapi): Regenerate swagger spec and internal client
* 4a87707 autogen: add v0.38.17-beta.1 to version.schema.json
* dedb92c autogen: pin v0.38.19-beta.1 release commit
* 6463019 autogen: update release artifacts
* db346d5 docs: fix "decisions" typo in Introduction (#907)
* 69c64e7 feat: allow overriding HTTP method for upstream calls
* 5c03d4f fix: bump Ory CLI
* 57c38c0 fix: update cve scanners (#905)
