# ORY Oathkeeper

## Introduction

Welcome to the ORY Oathkeeper documentation!

ORY Oathkeeper is a reverse proxy which evaluates incoming HTTP requests based on a set of rules. This capability
is referred to as an Identity and Access Proxy (IAP) in the context of [BeyondCorp and ZeroTrust](https://www.beyondcorp.com).

Rudimentally, ORY Oathkeeper inspects the `Authorization` header and the full request url (e.g. `https://mydomain.com/api/foo`)
of incoming HTTP requests, applies a given rule, and either grants access to the requested url or denies access.

Please keep in mind that ORY Oathkeeper is a supplement to ORY Hydra. It is thus imperative to be familiar with the core
concepts of ORY Hydra.

**OATHKEEPER DIAGRAM HERE**

## Concepts

### Terminology

* Access credentials: The credentials used to access an endpoint. This is typically an OAuth 2.0 access token supplied
    by ORY Hydra.
* Backend URL: The URL (backend) where requests will be forwarded to, if access is granted. Typically an API Gateway
    such as Mashape Kong.
* Scopes: An OAuth 2.0 Scope.

### Rules

ORY Oathkeeper has a configurable set of rules. Rules are applied to all incoming requests and based on the rule definition,
an action is taken. There are four types of rules:

1. **Passthrough**: Forwards the original request to the backend url without any modification to its headers.
2. **Anonymous**: Tries to extract user information from the given access credentials. If that fails, or no access
    credentials have been provided, the request is forwarded and the user is marked as "anonymous".
3. **NAME TO BE FOUND**: Requires valid access credentials and optionally checks for a set of OAuth 2.0 Scopes. If
    the supplied access credentials are invalid (expired, malformed, revoked) or do not fulfill the requested scopes,
    access is denied.
4. **NAME TO BE FOUND**: Requires valid access credentials as defined in 3. and additionally validates if the user
    is authorized to make the request, based on access control policies.

### Consumable Authorization

ORY Oathkeeper makes authentication and authorization data consumable to backends by providing an ID Token, as defined
by the OpenID Connect specification and [BeyondCorp](https://www.beyondcorp.com).

* Explain the payload of the ID token
* Explain how to validate the ID token

### API Docs

## Security

## Deployment

Best practices:

* Natively scalable because stateless
* Converts opaque tokens to transparent ones for internal ingestion
  * Explain that transparent tokens are bad practice when faced in the wild
* Small docker size, small footprint, resiliant