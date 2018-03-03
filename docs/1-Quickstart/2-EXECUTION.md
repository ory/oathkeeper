# Running ORY Oathkeeper

<!-- toc -->

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
  "scope": "hydra.warden hydra.keys.* hydra.introspect",
  "grant_types": ["client_credentials"],
  "response_types": ["token"]
}
```

**policy.json**
```json
{
  "subjects": ["oathkeeper-client"],
  "effect": "allow",
  "resources": [
    "rn:hydra:keys:oathkeeper:id-token<.*>",
    "rn:hydra:warden:<.*>",
    "rn:hydra:oauth2:tokens"
  ],
  "actions": [
    "decide",
    "get",
    "create",
    "introspect",
    "update",
    "delete"
  ]
}
```

and import them using ORY Hydra's CLI:

```
$ hydra clients import client.json
$ hydra policies import policy.json
```

Now we are all set to boot up ORY Oathkeeper. Because we are using the in-memory database backend, we have to use
`docker run oryd/oathkeeper:v0.0.20 serve all`, otherwise the proxy and management process would not be able to talk to each other:

```
# We assume that ORY Hydra is running on localhost:4444
$ export HYDRA_URL=http://localhost:4444/

# ORY Oathkeeper stores the public/private key in ORY Hydra using this set id
export HYDRA_JWK_SET_ID=oathkeeper:id-token

# These are the values from the client.json file
$ export HYDRA_CLIENT_ID=oathkeeper-client
$ export HYDRA_CLIENT_SECRET=something-secure

# This needs to be a URL of your HTTP endpoint.
$ export BACKEND_URL=http://my-api-endpoint-servers.com/

# This should be a proper database URL in production, see `hydra help serve proxy`
$ export DATABASE_URL=memory

# Each of environment variables needs to be injected into docker container properly '-e key=value'
$ docker run -e DATABASE_URL=$DATABASE_URL -e HYDRA_URL=$HYDRA_URL -e HYDRA_JWK_SET_ID=$HYDRA_JWK_SET_ID -e HYDRA_CLIENT_ID=$HYDRA_CLIENT_ID -e HYDRA_CLIENT_SECRET=$HYDRA_CLIENT_SECRET -e BACKEND_URL=$BACKEND_URL  oryd/oathkeeper:v0.0.20 serve all
```

Now, we have a proxy listening on port 4455 and the management REST API at port 4456.
