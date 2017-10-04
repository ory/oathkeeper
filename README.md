# ORY Oathkeeper

This is a simple reverse proxy that checks the HTTP Authorization for validity against a set of rules. This service
uses Hydra to validate access tokens and policies.

## Writing Rules

Check out [director/director.go](director/director.go) for some exemplary rules and rule documentation.

## Running

This service has a couple of environment variables:

* `BACKEND_URL` is the URL where requests should be forwarded to. If a path (and query) is used, they will be prepended to the request. (default: `http://localhost:7000`)
* `HYDRA_CLIENT` the client id used to access Hydra.
* `HYDRA_SECRET` the client secret used to access Hydra.
* `HYDRA_HOST` the URL of the Hydra instance.
* `PORT` the port to listen on. (default: `3000`)
* `HOST` the host to listen on.

You can run this sever using `go run main.go` or `go install . && firewall-reverse-proxy` or `docker build . && docker run <image>`

## Generate the mock

```
mockgen -package evaluator -destination evaluator/hydra_sdk_mock.go github.com/ory/hydra/sdk/go/hydra SDK
```