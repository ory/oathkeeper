<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Concepts](#concepts)
  - [Terminology](#terminology)
  - [Rules](#rules)
    - [Rules REST API](#rules-rest-api)
    - [Rules CLI API](#rules-cli-api)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Concepts

This section covers core concepts of ORY Oathekeeper.

## Terminology

* Access credentials: The credentials used to access an endpoint. This is typically an OAuth 2.0 access token supplied
    by ORY Hydra.
* Backend URL: The URL (backend) where requests will be forwarded to, if access is granted. Typically an API Gateway
    such as Mashape Kong.
* Scopes: An OAuth 2.0 Scope.
* Access Control Policies: These are JSON documents similar to AWS and Google Cloud Policies. Read more about them [here](https://ory.gitbooks.io/hydra/content/security.html#how-does-access-control-work-with-hydra).

## Rules

ORY Oathkeeper has a configurable set of rules. Rules are applied to all incoming requests and based on the rule definition,
an action is taken. There are four types of rules:

1. **Passthrough**: Forwards the original request to the backend url without any modification to its headers.
2. **Anonymous**: Tries to extract user information from the given access credentials. If that fails, or no access
    credentials have been provided, the request is forwarded and the user is marked as "anonymous".
3. **Basic Authorization**: Requires valid access credentials and optionally checks for a set of OAuth 2.0 Scopes. If
    the supplied access credentials are invalid (expired, malformed, revoked) or do not fulfill the requested scopes,
    access is denied.
4. **Policy Authorization**: Requires valid access credentials as defined in 3. and additionally validates if the user
    is authorized to make the request, based on access control policies.

The exact payloads are explained in detail in the REST API (see next section).

### Rules REST API

For more information on available fields and exemplary payloads of rules, as well as rule management using HTTP
please refer to the [REST API docs](https://oathkeeper.docs.apiary.io/#)

### Rules CLI API

Management of rules is not only possible through the REST API, but additionally using the ORY Oathkeeper CLI.
For help on how to manage the CLI, type `oathkeeper help rules`.
