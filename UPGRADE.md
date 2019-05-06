# Upgrading

The intent of this document is to make migration of breaking changes as easy as possible. Please note that not all
breaking changes might be included here. Please check the [CHANGELOG.md](./CHANGELOG.md) for a full list of changes
before finalizing the upgrade process.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Most recent release](#most-recent-release)
  - [Refresh Configuration](#refresh-configuration)
- [v0.13.9+oryOS.9](#v0139oryos9)
  - [Refresh Configuration](#refresh-configuration-1)
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

## master

## v0.16.0+oryOS.12

### all env vars have been renamed but you can use older ones np

### everything is disabled by default you must enable it explicitly

### id token

jwks are now no longer fetched from hydra

instead you can use `oathkeeper credentials generate [--alg <RS256>] [--bits <2048|4096>] <transformer_id_token>`

```
jwk-keygen --use sig --alg RS256 --bits 4096
```

### renamed

cookies -> cookie
headers -> header
credentials issuer -> mutator

### new

* new unauthorized authorizer

### Rule changes

Credential Issuer -> (Request) Transformer

### RSA key

* should no be imported from file/env

* should still work with hydra though

### Serve Changes

* serve api -> stays the same
* serve proxy -> expose 2 ports, one proxy on api for health check, metrics and so on

### SQL Store Deprecation

SQL -> in memory / from disk

## v0.15.0+oryOS.11

### New Go SDK Generator

The ORY Oathkeeper Go SDK is no being generated using [`go-swagger`](https://github.com/go-swagger/go-swagger) instead of
[`swagger-codegen`](https://github.com/go-swagger/go-swagger). If you have questions regarding upgrading, please open an issue.

## v0.14.0+oryOS.10

### Changes to the ORY Keto Authorizer

As ORY Keto's API and scope have changed, the `keto_warden` authorizer has changed as well. The most important
change is that the identifier changed from `keto_warden` to `keto_engine_acp_ory`. This reflects the new ORY Keto concept
which supports different engines. The functionality of the authorizer itself remains the same. A new configuration
option called `flavor` was added, which sets what flavor (e.g. `regex`, `exact`, ...). Here's an exemplary diff
of a rule using `keto_warden`

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

As part of this change, environment variable `AUTHORIZER_KETO_WARDEN_KETO_URL` was renamed to `AUTHORIZER_KETO_URL`.

### Environment variables

- Environment variables `HTTP_TLS_xxx` are now called `HTTPS_TLS_xxx`.
- Environment variable `AUTHORIZER_KETO_WARDEN_KETO_URL` is now `AUTHORIZER_KETO_URL`.

## v0.13.9+oryOS.9

### Refresh Configuration

Environment variable `CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_REFRESH_INTERVAL` is now called
`CREDENTIALS_ISSUER_ID_TOKEN_JWK_REFRESH_INTERVAL`.

### Scope Matching

Previously, `fosite.WildcardScopeStrategy` was used to validate OAuth 2.0 Scope. This is now configurable
with environment variables `AUTHENTICATOR_JWT_SCOPE_STRATEGY` and `AUTHENTICATOR_OAUTH2_INTROSPECTION_SCOPE_STRATEGY`.
Supported strategies are `HIERARCHIC`, `EXACT`, `WILDCARD`, `NONE`.

As part of this change, the default strategy is no longer `WILDCARD` but instead `EXACT`.

### Configuration changes

To improve compatibility with ORY Hydra v1.0.0-beta.8, which introduces the public and admin endpoint, the following
environment variables have now been made optional:

- `CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_ID`
- `CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_SECRET`
- `CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_SCOPES`
- `AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_ID`
- `AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_SECRET`
- `AUTHENTICATOR_OAUTH2_INTROSPECTION_TOKEN_URL`
- `AUTHENTICATOR_OAUTH2_INTROSPECTION_SCOPE`

They are optional because ORY Hydra's administrative endpoints no longer require authorization as they now
run on a privileged port. If you are running ORY Hydra behind a firewall that requires OAuth 2.0 Access tokens,
or you are using another OAuth 2.0 Server that requires an access token, you can still use these settings.

And the following environment variables have changed:

- `CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_URL` is now `CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_ADMIN_URL` and
`CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_PUBLIC_URL` if ORY Hydra is protected with OAuth 2.0.
- `AUTHENTICATOR_OAUTH2_INTROSPECTION_INTROSPECT_URL` is now `AUTHENTICATOR_OAUTH2_INTROSPECTION_URL`.

### CORS is disabled by default

A new environment variable `CORS_ENABLED` was introduced. It sets whether CORS is enabled ("true") or not ("false")".
Default is disabled.

## v0.13.8+oryOS.8

### `noop` authenticator no longer bypasses authorizers/credentials issuers

The `noop` authenticator is now very similar to `anonymous` with the difference that no anonymous subject is being
set.

Previously, the `noop` authenticator bypassed the authorizer and credential issuers. This patch changes that.

## v0.13.2+oryOS.2

This release introduces serious breaking changes. If you are upgrading, you will - unfortunately - need to
re-create the database schema and migrate your rules manually. While this is frustrating, there are a ton of features
that are added with this release:

- ORY Oathkeeper is now a standalone project and is independent from ORY Hydra.
- Supports generic & extensible authentication strategies like
  * OAuth 2.0 Token Introspection
  * OAuth 2.0 Client Credentials
  * JSON Web Token (in the future)
  * SAML (in the future)
  * ...
- Supports generic & extensible authorization strategies like
  * ORY Keto Warden API
  * Allow all
  * Deny all
  * ... more to come
- Supports generic & extensible credential issuance strategies like
  * ID Token
  * None
  * ...
* Supports basic routing logic per rule

We recommend re-reading the user guide.

If you are upgrading a production deployment and have issues or questions, reach out to the [ORY Community](https://discord.gg/PAMQWkr) or to [mailto:hi@ory.sh](hi@ory.sh).

### Changes to the CLI

Apart from various environment variables which changed (use `oathkeeper help serve proxy` and `oathkeeper help serve api` for an
overview), the `oathkeeper serve all` command has been deprecated.

The proxy command no longer needs access to the database, but instead pulls the information from the API using the `OATHKEEPER_API_URL`
environment variable.

Most notably, the `BACKEND_URL` environment variable was deprecated. Instead, rules define their upstream server themselves,
allowing for simple routing using this software.

#### `migrate`

Command `migrate` is now called `migrate sql`.

### Not compatible with ORY Hydra < 1.0.0

This release is not compatible with ORY Hydra versions < 1.0.0. Instead, it relies on a combination of ORY Hydra
and ORY Keto to provide the same functionality as before.

## 0.11.12

This release adds no breaking changes but brings this version up to speed with the latest version of ORY Hydra
that Oathkeeper works with.
