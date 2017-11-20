<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Quickstart](#quickstart)
  - [Installing ORY Oathkeeper](#installing-ory-oathkeeper)
  - [Running ORY Oathkeeper](#running-ory-oathkeeper)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Quickstart

This section covers how to get started using ORY Oathkeeper.

## Installing ORY Oathkeeper

The only way to install ORY Oathkeeper currently is through [Docker Hub]. ORY Oathkeeper is a private docker repository
and to gain access to it you have create an ORY Gatekeeper instance at the [ORY Admin Console](https://admin.ory.am).

Once your account is authorized to access the repository, log in to docker using `docker login`. Then you should
be able to fetch the ory/oathkeeper image and also be able to see it on [Docker Hub](https://hub.docker.com/r/oryd/oathkeeper/).

To run any oathkeeper command, do:

```
$ docker run oryd/oathkeeper:v<version> <command>
```

for example (this guide is written for ORY Oathkeeper 0.0.20:

```
$ docker run oryd/oathkeeper:v0.0.20 help
```

## Running ORY Oathkeeper

ORY Oathkeeper has two servers that run on separate ports:

* The management server: This server is responsible for exposing the management REST API.
* The proxy server: This server is responsible for evaluating access requests and forwarding them to the backend.

For detailed documentation on the two servers, run `oathkeeper help serve management` and `oathkeeper help serve proxy`.

ORY Oathkeeper supports two types of storage backends: In-memory and MySQL/PostgreSQL. The former is well suited
for testing while the latter should be used in production. For brevity, we will use the in-memory adapters for this
quickstart tutorial.

First, we need to set up an ORY Hydra instance as instructed [here](https://ory.gitbooks.io/hydra/content/install.html).
We also need to create an OAuth 2.0 client capable of making requests to ORY Hydra's Warden API. For this purpose we
create two files:

**client.json**
```json
{
  "id": "oathkeeper-client",
  "client_secret": "something-secure",
  "scope": "hydra.warden",
  "grant_types": ["client_credentials"],
  "response_types": ["token"]
}
```

**policy.json**
```json
{
  "id": "oathkeeper-policy",
  "subjects": [
    "oathkeeper-client"
  ],
  "effect": "allow",
  "resources": [
    "rn:hydra:warden:allowed",
    "rn:hydra:warden:token:allowed"
  ],
  "actions": [
    "decide"
  ]
}
```

and import them using ORY Hydra's CLI:

```
$ hydra clients import client.json
$ hydra policies create -f policy.json
```

Now we are all set to boot up ORY Oathkeeper. Because we are using the in-memory database backend, we have to use
`docker run oryd/oathkeeper:v0.0.20 serve all`, otherwise the proxy and management process would not be able to talk to each other:

```
# We assume that ORY Hydra is running on localhost:4444
$ export HYDRA_URL=http://localhost:4444/

# These are the values from the client.json file
$ export HYDRA_CLIENT_ID=oathkeeper-client
$ export HYDRA_CLIENT_SECRET=something-secure

# This needs to be a URL of your HTTP endpoint.
$ export BACKEND_URL=http://my-api-endpoint-servers.com/

# This should be a proper database URL in production, see `hydra help serve proxy`
$ export DATABASE_URL=memory

$ docker run oryd/oathkeeper:v0.0.20 serve all
```

Now, we have a proxy listening on port 4455 and the management REST API at port 4456.
