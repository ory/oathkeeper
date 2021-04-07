# Upgrading

The intent of this document is to make migration of breaking changes as easy as
possible. Please note that not all breaking changes might be included here.
Please check the [CHANGELOG.md](./CHANGELOG.md) for a full list of changes
before finalizing the upgrade process.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [v0.37](#v0370)
- [v0.36](#v0360)
- [v0.35.0-beta.1](#v0350-beta1)
- [v0.34.0-beta.1+oryOS.14](#v0340-beta1oryos14)
- [v0.33.0-beta.1+oryOS.13](#v0330-beta1oryos13)
- [v0.32.0-beta.1+oryOS.12](#v0320-beta1oryos12)
- [v0.19.0-beta.1+oryOS.12](#v0190-beta1oryos12)
  - [Config Changes](#config-changes)
  - [Hydrator Mutator](#hydrator-mutator)
- [v0.18.0-beta.1+oryOS.12](#v0180-beta1oryos12)
  - [Access Rule Mutators](#access-rule-mutators)
    - [`id_token` mutator now renders go templates](#id_token-mutator-now-renders-go-templates)
- [v0.17.0-beta.1+oryOS.12](#v0170-beta1oryos12)
- [v0.16.0-beta.1+oryOS.12](#v0160-beta1oryos12)
  - [Access Rule Changes](#access-rule-changes)
  - [Mutators (formerly credentials issuers)](#mutators-formerly-credentials-issuers)
    - [`id_token` works stand-alone](#id_token-works-stand-alone)
    - [`headers` -> `header`](#headers---header)
    - [`cookies` -> `cookie`](#cookies---cookie)
- [v0.15.0+oryOS.11](#v0150oryos11)
  - [New Go SDK Generator](#new-go-sdk-generator)
- [v0.14.0+oryOS.10](#v0140oryos10)
  - [Changes to the ORY Keto Authorizer](#changes-to-the-ory-keto-authorizer)
  - [Environment variables](#environment-variables)
- [v0.13.9+oryOS.9](#v0139oryos9)
  - [Refresh Configuration](#refresh-configuration)
  - [Scope Matching](#scope-matching)
  - [Configuration changes](#configuration-changes)
  - [CORS is disabled by default](#cors-is-disabled-by-default)
- [v0.13.8+oryOS.8](#v0138oryos8)
  - [`noop` authenticator no longer bypasses authorizers/credentials issuers](#noop-authenticator-no-longer-bypasses-authorizerscredentials-issuers)
- [v0.13.2+oryOS.2](#v0132oryos2)
  - [Changes to the CLI](#changes-to-the-cli)
    - [`migrate`](#migrate)
  - [Not compatible with ORY Hydra < 1.0.0](#not-compatible-with-ory-hydra--100)
- [0.11.12](#01112)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## v0.37

BREAKING CHANGES:

- This patch allows using the regex capture groups from the URL matcher to be
  used in several places, including the ID Token generator and elsewhere. To get
  this working, existing `keto_engine_acp_ory` authorizers are no longer able to
  use regex substition in the form of `my:action:$1` but instead must use the
  new format which is `{{ printIndex .MatchContext.RegexpCaptureGroups 0}}`
  (notice that the index changed by _-1_). A rule migrator exists which makes
  old rules compatible with the new format, if a version string is given. More
  details on the rule migration can be found here:
  https://github.com/ory/oathkeeper/commit/fd16ceb230a1b14ebb01a147d2d70acce77f9fbd#diff-6177fb19f1b7d7bc392f5062b838df15

## v0.36

The access rule matcher now supports the
[Glob patterns](https://github.com/gobwas/glob)!

## v0.35.0-beta.1

This release focuses on a rework of the SDK pipeline. First of all, we have
introduced new SDKs for all popular programming languages and published them on
their respective package repositories:

- [Python](https://pypi.org/project/ory-oathkeeper-client/)
- [PHP](https://packagist.org/packages/ory/oathkeeper-client)
- [Go](https://github.com/ory/oathkeeper-client-go)
- [NodeJS](https://www.npmjs.com/package/@oryd/oathkeeper-client) (with
  TypeScript)
- [Java](https://search.maven.org/artifact/sh.ory.oathkeeper/oathkeeper-client)
- [Ruby](https://rubygems.org/gems/ory-oathkeeper-client)

The SDKs hosted in this repository (under ./sdk/...) have been completely
removed. Please use only the SDKs from the above sources from now on as it will
also remove several issues that were caused by the previous SDK pipeline.

Unfortunately, there were breaking changes introduced by the new SDK generation:

- Several structs and fields have been renamed in the Go SDK. However, nothing
  else changed so upgrading should be a matter of half an hour if you made
  extensive use of the SDK, or several minutes if just one or two methods are
  being used.
- All other SDKs changed to `openapi-generator`, which is a better maintained
  generator that creates better code than the one previously used. This
  manifests in TypeScript definitions for the NodeJS SDK and several other
  goodies. We do not have a proper migration path for those, unfortunately.

If you have issues with upgrading the SDK, please let us know in an issue on
this repository!

## v0.34.0-beta.1+oryOS.14

A new feature was introduced which allows you to handle errors in different ways
(e.g. redirect to login, return JSON error, ...) depending on the error and HTTP
request type. For more information, head over to:
https://www.ory.sh/docs/next/oathkeeper/pipeline/error

One backwards incompatible change has been made, as JSON errors returned by ORY
Oathkeeper now include only minimal information about the error source. You can
change this behavior by setting `verbose` to `true`, see:
https://www.ory.sh/docs/next/oathkeeper/pipeline/error#json

No other backwards incompatible changes have been made!

## v0.33.0-beta.1+oryOS.13

The "mutator" hydrator config has changed:

- `config.retry.delay_in_milliseconds: 100` (int) is now
  `config.retry.max_delay: 100ms` (duration)
- `config.retry.max_retries: 3` (int) is now `config.retry.give_up_after: 1s`
  (duration)

A new feature introduce in this release allows to keep using existing access
rules by setting `"version": "v0.32.0-beta.1"` in the existing rules. ORY
Oathkeeper will migrate the old config to the new config.

This access rule definition will properly be migrated:

```
{
+  "version": "v0.32.0-beta.1",
  "mutators": [
    {
      "handler": "hydrator",
      "config": {
        "retry": {
          "delay_in_milliseconds": 500,
          "max_retries": 5
        }
      }
    }
  ]
}
```

We encourage you to tag all your access rules with the version.

## v0.32.0-beta.1+oryOS.12

An issue with the release pipeline has been resolved, which required several
version increases. No functionality has changed in a backwards incompatible way.

## v0.19.0-beta.1+oryOS.12

### Config Changes

This release homogenizes all configuration settings. Previously all handlers
(mutators, authenticators, and authorizers) had two different types of config:
global and per access rule.

With this release, all handlers have the same configuration for global and per
access rule. For example, the `id_token` handler requires the `issuer_url`.
Previously, this value was only configurable in the global config. Now, it can
be set on a per rule basis as well as globally. The global config will always be
used as a fallback when no access rule specific configuration is set.

For this to work, the ORY Oathkeeper global configuration file
(`~/.oathkeeper.yaml`) has changed when it comes to mutators, authenticaotrs,
and authorizers. Instead of defining the config at the same level as the
`enabled` flag, it is now nested in a subkey "config":

```
authorizers:
  jwt:
    enabled: true
-    jwks_urls:
-    - foo
-    - bar
+    config
+      jwks_urls:
+      - foo
+      - bar
```

### Hydrator Mutator

The Hydrator mutator has two configuration keys `api.retry.number` and
`api.retry.delayInMilliseconds`. These have been renamed for consistency reasons
to: `api.retry.number_of_retries` and `api.retry.delay_in_milliseconds`.

## v0.18.0-beta.1+oryOS.12

### Access Rule Mutators

1. ORY Oathkeeper now supports multiple mutators. Mutations are performed in the
   provided order and must all succeed in order for the HTTP request to be
   forwarded.
2. The `mutator` property was renamed to `mutators` to reflect its true nature
   (see previous item).

If you have existing rules, please update them as follows:

```patch
[
  {
    "id": "jwt-rule",
    "upstream": {
      "url": "http://127.0.0.1:6662"
    },
    "match": {
      "url": "http://127.0.0.1:<6660|6661>/jwt",
      "methods": [
        "GET"
      ]
    },
    "authenticators": [
      {
        "handler": "jwt"
      }
    ],
    "authorizer": {
      "handler": "allow"
    },
-   "mutator": {
-     "handler": "id_token"
-   }
+   "mutators": [
+     {
+       "handler": "id_token"
+     }
+   ]
  }
]
```

#### `id_token` mutator now renders go templates

The `id_token` mutator is now capable of rendering custom claims using Go
[text/template](https://golang.org/pkg/text/template/) receiving the
`AuthenticationSession` struct as its parameters.

To enable this change, the `aud` config was removed and the `claims` config was
introduced. The `claims` field is a raw string representing a Go template.

To upgrade existing rules, apply patches similar to this one:

deprecated config:

```patch
{
  "handler": "id_token",
  "config": {
-    "aud": ["https://my-backend-service/some/endpoint"]
+    "claims": "{\"aud\": [\"https://my-backend-service/some/endpoint\"]}"
  }
}
```

## v0.17.0-beta.1+oryOS.12

ORY Oathkeeper now watches configuration files and access rules repositories on
the local disk for changes. This does currently not work for remote sources
(http/https). Additionally, access rules can now be written in YAML (expected
file extensions are `yaml` and `yml`).

## v0.16.0-beta.1+oryOS.12

ORY Oathkeeper was changed according to discussion
[177](https://github.com/ory/oathkeeper/issues/177). Several issues have been
resolved that could not be resolved before due to design decisions. We strongly
encourage you to re-read the
[documentation](https://www.ory.sh/oathkeeper/docs/) but to give you a short
overview of the most important changes:

1. Commands `oathkeeper serve api` and `oathkeeper serve proxy` have been
   deprecated of `oathkeeper serve` which exposes two ports (reverse proxy,
   API).
1. ORY Oathkeeper can now be configured from a file and configuration keys where
   updated. Where appropriate, environment variables from previous versions
   still work. Please check out
   [./docs/config.yml](internal/config/.oathkeeper.yaml) for a fully annotated
   configuration file as several environment variables changed, for example (not
   exclusive): `HTTPS_TLS_CERT_PATH`, `HTTPS_TLS_KEY_PATH`, `HTTPS_TLS_CERT`,
   `HTTPS_TLS_KEY`.
1. The Judge API (`/judge`) was renamed to Access Control Decision API
   (`/decisions`)
1. The need for a database was completely removed. Also, ORY Oathkeeper no
   longer runs as two separate processes but instead as one process that opens
   two ports (one proxy, one API).
1. For consistency, JWT claims `scope`, `scp`, `scopes` will always be
   transformed to `scp` (string[]) in the `jwt` authenticator.
1. ORY Oathkeeper no longer requires a database. Instead, cryptographic keys,
   access rules, and other configuration items are loaded from the file system,
   environment variables, or HTTP(s) locations.
1. Credential Issuers are now called `mutators` as they mutate the HTTP Request
   (Headers) for upstream services.
1. All authentication, authorization and mutation handlers are disabled by
   default and must be enabled and configured explicitly.

### Access Rule Changes

As already noted, `credentials_issuer` was renamed to `mutator`. If you have
existing rules, please update them as follows:

```
[
  {
    "id": "jwt-rule",
    "upstream": {
      "url": "http://127.0.0.1:6662"
    },
    "match": {
      "url": "http://127.0.0.1:6660/jwt",
      "methods": [
        "GET"
      ]
    },
    "authenticators": [
      {
        "handler": "jwt"
      }
    ],
    "authorizer": {
      "handler": "allow"
    },
-   "credentials_issuer": {
+   "mutator": {
      "handler": "id_token"
    }
  }
]
```

### Mutators (formerly credentials issuers)

#### `id_token` works stand-alone

The ID Token Mutator has completely been reworked. It no longer requires ORY
Hydra for RS256 algorithms but instead loads the required cryptographic keys
from the file system, environment variables, or a remote HTTP/HTTPS location.

To make development easy, ORY Oathkeeper ships a CLI command that allows you to
quickly create such a cryptographic key:

```shell
$ oathkeeper credentials generate --alg <RS256|ES256|HS256|RS512|...>
```

#### `headers` -> `header`

The ID of the Header Mutator has been updated from `headers` to `header`. Please
apply a patch similar to the listed one to your access rules:

```
[
  {
    "id": "jwt-rule",
    "upstream": {
      "url": "http://127.0.0.1:6662"
    },
    "match": {
      "url": "http://127.0.0.1:6660/jwt",
      "methods": [
        "GET"
      ]
    },
    "authenticators": [
      {
        "handler": "jwt"
      }
    ],
    "authorizer": {
      "handler": "allow"
    },
   "mutator": {
-      "handler": "headers"
+      "handler": "header"
    }
  }
]
```

#### `cookies` -> `cookie`

The ID of the Cookie Mutator has been updated from `cookies` to `cookie`. Please
apply a patch similar to the listed one to your access rules:

```
[
  {
    "id": "jwt-rule",
    "upstream": {
      "url": "http://127.0.0.1:6662"
    },
    "match": {
      "url": "http://127.0.0.1:6660/jwt",
      "methods": [
        "GET"
      ]
    },
    "authenticators": [
      {
        "handler": "jwt"
      }
    ],
    "authorizer": {
      "handler": "allow"
    },
   "mutator": {
-      "handler": "cookies"
+      "handler": "cookie"
    }
  }
]
```

## v0.15.0+oryOS.11

### New Go SDK Generator

The ORY Oathkeeper Go SDK is no being generated using
[`go-swagger`](https://github.com/go-swagger/go-swagger) instead of
[`swagger-codegen`](https://github.com/go-swagger/go-swagger). If you have
questions regarding upgrading, please open an issue.

## v0.14.0+oryOS.10

### Changes to the ORY Keto Authorizer

As ORY Keto's API and scope have changed, the `keto_warden` authorizer has
changed as well. The most important change is that the identifier changed from
`keto_warden` to `keto_engine_acp_ory`. This reflects the new ORY Keto concept
which supports different engines. The functionality of the authorizer itself
remains the same. A new configuration option called `flavor` was added, which
sets what flavor (e.g. `regex`, `exact`, ...). Here's an exemplary diff of a
rule using `keto_warden`

```
{
  "id": "...",
  "upstream": ...,
  "match": ...,
  "authenticators": ...,
  "authorizer": {
-    "handler": "keto_warden",
+    "handler": "keto_engine_acp_ory",
    "config": {
      "required_action": "...",
      "required_resource": ...",
      "subject": ...",
+      "flavor": "exact" (optional, defaults to `regex`)
    }
  },
  "credentials_issuer": ...
}
```

As part of this change, environment variable `AUTHORIZER_KETO_WARDEN_KETO_URL`
was renamed to `AUTHORIZER_KETO_URL`.

### Environment variables

- Environment variables `HTTP_TLS_xxx` are now called `HTTPS_TLS_xxx`.
- Environment variable `AUTHORIZER_KETO_WARDEN_KETO_URL` is now
  `AUTHORIZER_KETO_URL`.

## v0.13.9+oryOS.9

### Refresh Configuration

Environment variable `CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_REFRESH_INTERVAL` is now
called `CREDENTIALS_ISSUER_ID_TOKEN_JWK_REFRESH_INTERVAL`.

### Scope Matching

Previously, `fosite.WildcardScopeStrategy` was used to validate OAuth 2.0 Scope.
This is now configurable with environment variables
`AUTHENTICATOR_JWT_SCOPE_STRATEGY` and
`AUTHENTICATOR_OAUTH2_INTROSPECTION_SCOPE_STRATEGY`. Supported strategies are
`HIERARCHIC`, `EXACT`, `WILDCARD`, `NONE`.

As part of this change, the default strategy is no longer `WILDCARD` but instead
`EXACT`.

### Configuration changes

To improve compatibility with ORY Hydra v1.0.0-beta.8, which introduces the
public and admin endpoint, the following environment variables have now been
made optional:

- `CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_ID`
- `CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_SECRET`
- `CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_SCOPES`
- `AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_ID`
- `AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_SECRET`
- `AUTHENTICATOR_OAUTH2_INTROSPECTION_TOKEN_URL`
- `AUTHENTICATOR_OAUTH2_INTROSPECTION_SCOPE`

They are optional because ORY Hydra's administrative endpoints no longer require
authorization as they now run on a privileged port. If you are running ORY Hydra
behind a firewall that requires OAuth 2.0 Access tokens, or you are using
another OAuth 2.0 Server that requires an access token, you can still use these
settings.

And the following environment variables have changed:

- `CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_URL` is now
  `CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_ADMIN_URL` and
  `CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_PUBLIC_URL` if ORY Hydra is protected with
  OAuth 2.0.
- `AUTHENTICATOR_OAUTH2_INTROSPECTION_INTROSPECT_URL` is now
  `AUTHENTICATOR_OAUTH2_INTROSPECTION_URL`.

### CORS is disabled by default

A new environment variable `CORS_ENABLED` was introduced. It sets whether CORS
is enabled ("true") or not ("false")". Default is disabled.

## v0.13.8+oryOS.8

### `noop` authenticator no longer bypasses authorizers/credentials issuers

The `noop` authenticator is now very similar to `anonymous` with the difference
that no anonymous subject is being set.

Previously, the `noop` authenticator bypassed the authorizer and credential
issuers. This patch changes that.

## v0.13.2+oryOS.2

This release introduces serious breaking changes. If you are upgrading, you
will - unfortunately - need to re-create the database schema and migrate your
rules manually. While this is frustrating, there are a ton of features that are
added with this release:

- ORY Oathkeeper is now a standalone project and is independent from ORY Hydra.
- Supports generic & extensible authentication strategies like
  - OAuth 2.0 Token Introspection
  - OAuth 2.0 Client Credentials
  - JSON Web Token (in the future)
  - SAML (in the future)
  - ...
- Supports generic & extensible authorization strategies like
  - ORY Keto Warden API
  - Allow all
  - Deny all
  - ... more to come
- Supports generic & extensible credential issuance strategies like
  - ID Token
  - None
  - ...

* Supports basic routing logic per rule

We recommend re-reading the user guide.

If you are upgrading a production deployment and have issues or questions, reach
out to the [ORY Community](https://discord.gg/PAMQWkr) or to
[mailto:hi@ory.sh](hi@ory.sh).

### Changes to the CLI

Apart from various environment variables which changed (use
`oathkeeper help serve proxy` and `oathkeeper help serve api` for an overview),
the `oathkeeper serve all` command has been deprecated.

The proxy command no longer needs access to the database, but instead pulls the
information from the API using the `OATHKEEPER_API_URL` environment variable.

Most notably, the `BACKEND_URL` environment variable was deprecated. Instead,
rules define their upstream server themselves, allowing for simple routing using
this software.

#### `migrate`

Command `migrate` is now called `migrate sql`.

### Not compatible with ORY Hydra < 1.0.0

This release is not compatible with ORY Hydra versions < 1.0.0. Instead, it
relies on a combination of ORY Hydra and ORY Keto to provide the same
functionality as before.

## 0.11.12

This release adds no breaking changes but brings this version up to speed with
the latest version of ORY Hydra that Oathkeeper works with.
