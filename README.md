<h1 align="center"><img src="./docs/images/banner_oathkeeper.png" alt="ORY Oathkeeper - Cloud Native Identity & Access Proxy"></h1>

<h4 align="center">
    <a href="https://discord.gg/PAMQWkr">Chat</a> |
    <a href="https://community.ory.am/">Forums</a> |
    <a href="http://eepurl.com/di390P">Newsletter</a><br/><br/>
    <a href="https://www.ory.sh/docs/guides/master/oathkeeper/">Guide</a> |
    <a href="https://www.ory.sh/docs/api/oathkeeper?version=master">API Docs</a> |
    <a href="https://godoc.org/github.com/ory/oathkeeper">Code Docs</a><br/><br/>
    <a href="https://opencollective.com/ory-hydra">Support this project!</a>
</h4>

ORY Oathkeeper is an Identity & Access Proxy (IAP) that authorizes HTTP requests based on sets of rules. The BeyondCorp
Model is designed by [Google](https://cloud.google.com/beyondcorp/) and secures applications in Zero-Trust networks.
An Identity & Access Proxy is typically deployed in front of (think API Gateway) web-facing applications and is capable
of authenticating and optionally authorizing access requests.

ORY Oathkeeper is a reverse proxy which evaluates incoming HTTP requests based on a set of rules that are defined
by administartive users. ORY Oathkeeper is thus capable of:

* Identifying the user and providing the user session to API backends.
* Restricting access to certain resources based on a set of rules (Authorization).

This service is under active development and may introduce breaking changes in future releases.

[![CircleCI](https://circleci.com/gh/ory/oathkeeper.svg?style=shield&circle-token=eb458bf636326d41674141b6bbfa475a39c9db1e)](https://circleci.com/gh/ory/oathkeeper)
[![Coverage Status](https://coveralls.io/repos/github/ory/oathkeeper/badge.svg?branch=master)](https://coveralls.io/github/ory/oathkeeper?branch=master)
![Go Report Card](https://goreportcard.com/badge/github.com/ory/oathkeeper)

---

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Installation](#installation)
  - [Download binaries](#download-binaries)
  - [Using Docker](#using-docker)
  - [Building from source](#building-from-source)
- [Ecosystem](#ecosystem)
  - [ORY Security Console: Administrative User Interface](#ory-security-console-administrative-user-interface)
  - [ORY Hydra: OAuth2 & OpenID Connect Server](#ory-hydra-oauth2--openid-connect-server)
  - [ORY Keto: Access Control Policies as a Server](#ory-keto-access-control-policies-as-a-server)
- [Security](#security)
  - [Disclosing vulnerabilities](#disclosing-vulnerabilities)
- [Telemetry](#telemetry)
- [Documentation](#documentation)
  - [Guide](#guide)
  - [HTTP API documentation](#http-api-documentation)
  - [Upgrading and Changelog](#upgrading-and-changelog)
  - [Command line documentation](#command-line-documentation)
  - [Develop](#develop)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation

There are various ways of installing ORY Oathkeeper on your system.

### Download binaries

The client and server **binaries are downloadable at [releases](https://github.com/ory/oathkeeper/releases)**.
There is currently no installer available. You have to add the ORY Oathkeeper binary to the PATH environment variable yourself or put
the binary in a location that is already in your path (`/usr/bin`, ...).
If you do not understand what that all of this means, ask in our [chat channel](https://www.ory.sh/chat). We are happy to help.

### Using Docker

**Starting the host** is easiest with docker. The host process handles HTTP requests and is backed by a database.
Read how to install docker on [Linux](https://docs.docker.com/linux/), [OSX](https://docs.docker.com/mac/) or
[Windows](https://docs.docker.com/windows/). ORY Oathkeeper is available on [Docker Hub](https://hub.docker.com/r/oryd/oathkeeper/).

You can use ORY Oathkeeper without a database, but be aware that restarting, scaling
or stopping the container will **lose all data**:

```
$ docker run -e "DATABASE_URL=memory" -d --name my-oathkeeper -p 4455:4455 -p 4456:4456 oryd/oathkeeper serve api
ec91228cb105db315553499c81918258f52cee9636ea2a4821bdb8226872f54b
```

### Building from source

If you wish to compile ORY Oathkeeper yourself, you need to install and set up [Go 1.10+](https://golang.org/) and add `$GOPATH/bin`
to your `$PATH` as well as [golang/dep](http://github.com/golang/dep).

The following commands will check out the latest release tag of ORY Oathkeeper and compile it and set up flags so that `oathkeeper version`
works as expected. Please note that this will only work with a linux shell like bash or sh.

```
go get -d -u github.com/ory/oathkeeper
cd $(go env GOPATH)/src/github.com/ory/oathkeeper
OATHKEEPER_LATEST=$(git describe --abbrev=0 --tags)
git checkout $OATHKEEPER_LATEST
dep ensure -vendor-only
go install \
    -ldflags "-X github.com/ory/oathkeeper/cmd.Version=$OATHKEEPER_LATEST -X github.com/ory/oathkeeper/cmd.BuildTime=`TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ'` -X github.com/ory/oathkeeper/cmd.GitHash=`git rev-parse HEAD`" \
    github.com/ory/oathkeeper
git checkout master
oathkeeper help
```

## Ecosystem

<a href="https://console.ory.am/auth/login">
    <img align="right" width="30%" src="docs/images/sec-console.png" alt="ORY Security Console">
</a>

### ORY Security Console: Administrative User Interface

The [ORY Security Console](https://console.ory.am/auth/login) is a visual admin interface for managing ORY Hydra,
ORY Oathkeeper, and ORY Keto.

### ORY Hydra: OAuth2 & OpenID Connect Server

[ORY Hydra](https://github.com/ory/hydra) ORY Hydra is a hardened OAuth2 and OpenID Connect server optimized
for low-latency, high throughput, and low resource consumption. ORY Hydra is not an identity provider
(user sign up, user log in, password reset flow), but connects to your existing identity provider through a consent app.

### ORY Keto: Access Control Policies as a Server

[ORY Keto](https://github.com/ory/keto) is a policy decision point. It uses a set of access control policies, similar
to AWS IAM Policies, in order to determine whether a subject (user, application, service, car, ...) is authorized to
perform a certain action on a resource.

## Security

### Disclosing vulnerabilities

If you think you found a security vulnerability, please refrain from posting it publicly on the forums, the chat, or GitHub
and send us an email to [hi@ory.am](mailto:hi@ory.am) instead.

## Telemetry

Our services collect summarized, anonymized data which can optionally be turned off. Click
[here](https://www.ory.sh/docs/guides/latest/9-telemetry) to learn more.

## Documentation

### Guide

The Guide is available [here](https://www.ory.sh/docs/guides/master/oathkeeper/).

### HTTP API documentation

The HTTP API is documented [here](https://www.ory.sh/docs/api/oathkeeper?version=master).

### Upgrading and Changelog

New releases might introduce breaking changes. To help you identify and incorporate those changes, we document these
changes in [UPGRADE.md](./UPGRADE.md) and [CHANGELOG.md](./CHANGELOG.md).

### Command line documentation

Run `oathkeeper -h` or `oathkeeper help`.

### Develop

Developing with ORY Oathkeeper is as easy as:

```
go get -d -u github.com/ory/oathkeeper
cd $GOPATH/src/github.com/ory/oathkeeper
dep ensure
go test ./...
```

Then run it with in-memory database:

```
DATABASE_URL=memory go run main.go serve all
```
