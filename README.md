# ORY Oathkeeper

<h4 align="center">
    <a href="https://gitter.im/ory-am/hydra">Chat</a> |
    <a href="https://community.ory.am/">Forums</a> |
    <a href="http://eepurl.com/bKT3N9">Newsletter</a><br/><br/>
    <a href="http://docs.oathkeeper.apiary.io/">Developer Guide</a> |
    <a href="http://docs.oathkeeper.apiary.io/">API Docs</a> |
    <a href="https://patreon.com/user?u=4298803">Support us on patreon!</a>
</h4>

ORY Oathkeeper is a reverse proxy that checks the HTTP Authorization for validity against a set of rules. This service
uses Hydra to validate access tokens and policies. This service is under **active development** with **regular breaking changes**.

[![CircleCI](https://circleci.com/gh/ory/oathkeeper.svg?style=svg&circle-token=eb458bf636326d41674141b6bbfa475a39c9db1e)](https://circleci.com/gh/ory/oathkeeper)

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
