---
id: index
title: Overview
---

All SDKs use automated code generation provided by
[`openapi-generator`](https://github.com/OpenAPITools/openapi-generator).
Unfortunately, `openapi-generator` has serious breaking changes in the generated
code when upgrading versions. Therefore, we do not make backwards compatibility
promises with regards to the generated SDKs. We hope to improve this process in
the future.

Before you check out the SDKs, head over to the
[REST API](oathkeeper/sdk/api.md) documentation which includes code samples for
common programming languages for each REST endpoint.

We publish our SDKs for popular languages in their respective package
repositories:

- [Python](https://pypi.org/project/ory-oathkeeper-client/)
- [PHP](https://packagist.org/packages/ory/oathkeeper-client)
- [Go](https://github.com/ory/oathkeeper-client-go)
- [NodeJS](https://www.npmjs.com/package/@oryd/oathkeeper-client) (with
  TypeScript)
- [Java](https://search.maven.org/artifact/sh.ory.oathkeeper/oathkeeper-client)
- [Ruby](https://rubygems.org/gems/ory-oathkeeper-client)

Missing your programming language?
[Create an issue](https://github.com/ory/oathkeeper/issues) and help us build,
test and publish the SDK for your programming language!
