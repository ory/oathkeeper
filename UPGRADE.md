# Upgrading

The intent of this document is to make migration of breaking changes as easy as possible. Please note that not all
breaking changes might be included here. Please check the [CHANGELOG.md](./CHANGELOG.md) for a full list of changes
before finalizing the upgrade process.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [0.1.0](#010)
  - [Changes to the CLI](#changes-to-the-cli)
  - [Not compatible with ORY Hydra < 1.0.0](#not-compatible-with-ory-hydra--100)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## 0.1.0

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

The following will list breaking changes so you can get an idea of what changed, but the list might be incomplete.
We recommend re-reading the user guide.

If you are upgrading a production deployment and have issues or questions, reach out to the [ORY Community](https://discord.gg/PAMQWkr).

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
