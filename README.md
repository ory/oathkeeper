<h1 align="center"><img src="./docs/images/banner_oathkeeper.png" alt="ORY Oathkeeper - Cloud Native Identity & Access Proxy"></h1>

<h4 align="center">
    <a href="https://discord.gg/PAMQWkr">Chat</a> |
    <a href="https://community.ory.am/">Forums</a> |
    <a href="http://eepurl.com/di390P">Newsletter</a><br/><br/>
    <a href="https://www.ory.sh/docs/guides/master/oathkeeper/">Guide</a> |
    <a href="https://www.ory.sh/docs/api/oathkeeper?version=master">API Docs</a> |
    <a href="https://godoc.org/github.com/ory/oathkeeper">Code Docs</a><br/><br/>
    <a href="https://opencollective.com/ory">Support this project!</a>
</h4>

ORY Oathkeeper is an Identity & Access Proxy (IAP) and Access Control Decision
API that authorizes HTTP requests based on sets of Access Rules. The BeyondCorp
Model is designed by [Google](https://cloud.google.com/beyondcorp/) and secures
applications in Zero-Trust networks.

An Identity & Access Proxy is typically deployed in front of (think API Gateway)
web-facing applications and is capable of authenticating and optionally
authorizing access requests. The Access Control Decision API can be deployed
alongside an existing API Gateway or reverse proxy. ORY Oathkeeper's Access
Control Decision API works with:

- [Ambassador](https://github.com/datawire/ambassador) via
  [auth service](https://www.getambassador.io/reference/services/auth-service).
- [Envoy](https://www.envoyproxy.io) via the
  [External Authorization HTTP Filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/ext_authz_filter#config-http-filters-ext-authz)
- AWS API Gateway via
  [Custom Authorizers](https://aws.amazon.com/de/blogs/compute/introducing-custom-authorizers-in-amazon-api-gateway/)
- [Nginx](https://www.nginx.com) via
  [Authentication Based on Subrequest Result](https://docs.nginx.com/nginx/admin-guide/security-controls/configuring-subrequest-authentication/)

among others.

This service is stable, but under active development and may introduce breaking
changes in future releases. Any breaking change will have extensive
documentation and upgrade instructions.

[![CircleCI](https://circleci.com/gh/ory/oathkeeper.svg?style=shield&circle-token=eb458bf636326d41674141b6bbfa475a39c9db1e)](https://circleci.com/gh/ory/oathkeeper)
[![Coverage Status](https://coveralls.io/repos/github/ory/oathkeeper/badge.svg?branch=master)](https://coveralls.io/github/ory/oathkeeper?branch=master)
![Go Report Card](https://goreportcard.com/badge/github.com/ory/oathkeeper)

---

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Installation](#installation)
- [Who's using it?](#whos-using-it)
- [Ecosystem](#ecosystem)
  - [ORY Security Console: Administrative User Interface](#ory-security-console-administrative-user-interface)
  - [ORY Hydra: OAuth2 & OpenID Connect Server](#ory-hydra-oauth2--openid-connect-server)
  - [ORY Keto: Access Control Policies as a Server](#ory-keto-access-control-policies-as-a-server)
  - [Examples](#examples)
- [Security](#security)
  - [Disclosing vulnerabilities](#disclosing-vulnerabilities)
- [Telemetry](#telemetry)
- [Documentation](#documentation)
  - [Guide](#guide)
  - [HTTP API documentation](#http-api-documentation)
  - [Upgrading and Changelog](#upgrading-and-changelog)
  - [Command line documentation](#command-line-documentation)
  - [Develop](#develop)
- [Backers](#backers)
- [Sponsors](#sponsors)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation

Head over to the
[ORY Developer Documentation](https://www.ory.sh/docs/oathkeeper/install) to
learn how to install ORY Oathkeeper on Linux, macOS, Windows, and Docker and how
to build ORY Oathkeeper from source.

## Who's using it?

<!--BEGIN ADOPTERS-->

The ORY community stands on the shoulders of individuals, companies, and
maintainers. We thank everyone involved - from submitting bug reports and
feature requests, to contributing patches, to sponsoring our work. Our community
is 1000+ strong and growing rapidly. The ORY stack protects 1.200.000.000+ API
requests every month with over 15.000+ active service nodes. Our small but
expert team would have never been able to achieve this without each and everyone
of you.

The following list represents companies that have accompanied us along the way
and that have made outstanding contributions to our ecosystem. _If you think
that your company deserves a spot here, reach out to
<a href="mailto:hi@ory.sh">hi@ory.sh</a>now_!

**Please consider giving back by becoming a sponsor of our open source work on
<a href="https://www.patreon.com/_ory">Patreon</a> or
<a href="https://opencollective.com/ory">Open Collective</a>.**

<table>
    <thead>
        <tr>
            <th>Type</th>
            <th>Name</th>
            <th>Logo</th>
            <th>Website</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>Sponsor</td>
            <td>Raspberry PI Foundation</td>
            <td align="center"><img height="32px" src="./docs/adopters/raspi.svg" alt="Raspberry PI Foundation"></td>
            <td><a href="https://www.raspberrypi.org/">raspberrypi.org</a></td>
        </tr>
        <tr>
            <td>Contributor</td>
            <td>Kyma Project</a>
            <td align="center"><img height="32px" src="./docs/adopters/kyma.svg" alt="Kyma Project"></td>
            <td><a href="https://kyma-project.io">kyma-project.io</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>ThoughtWorks</td>
            <td align="center"><img height="32px" src="./docs/adopters/tw.svg" alt="ThoughtWorks"></td>
            <td><a href="https://www.thoughtworks.com/">thoughtworks.com</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Tulip</td>
            <td align="center"><img height="32px" src="./docs/adopters/tulip.svg" alt="Tulip Retail"></td>
            <td><a href="https://tulip.com/">tulip.com</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Cashdeck / All My Funds</td>
            <td align="center"><img height="32px" src="./docs/adopters/allmyfunds.svg" alt="All My Funds"></td>
            <td><a href="https://cashdeck.com.au/">cashdeck.com.au</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>3 Rein</td>
            <td align="center"><img height="32px" src="./docs/adopters/3R-horiz.svg" alt="3REIN"></td>
            <td><a href="https://3rein.com/">3rein.com</a> <em>(avaiable soon)</em></td>
        </tr>
        <tr>
            <td>Contributor</td>
            <td>Hootsuite</td>
            <td align="center"><img height="32px" src="./docs/adopters/hootsuite.svg" alt="Hootsuite"></td>
            <td><a href="https://hootsuite.com/">hootsuite.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Segment</td>
            <td align="center"><img height="32px" src="./docs/adopters/segment.svg" alt="Segment"></td>
            <td><a href="https://segment.com/">segment.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Arduino</td>
            <td align="center"><img height="32px" src="./docs/adopters/arduino.svg" alt="Arduino"></td>
            <td><a href="https://www.arduino.cc/">arduino.cc</a></td>
        </tr>
    </tdbody>
</table>

We also want to thank all individual contributors

<img src="https://opencollective.com/ory/contributors.svg?width=890&button=false" /></a>

as well as all of our backers

<a href="https://opencollective.com/ory#backers" target="_blank"><img src="https://opencollective.com/ory/backers.svg?width=890"></a>

and past & current supporters (in alphabetical order) on
[Patreon](https://www.patreon.com/_ory): Alexander Alimovs, Billy, Chancy
Kennedy, Drozzy, Edwin Trejos, Howard Edidin, Ken Adler Oz Haven, Stefan Hans,
TheCrealm.

<em>\* Uses one of ORY's major projects in production.</em>

<!--END ADOPTERS-->

## Ecosystem

<a href="https://console.ory.sh/">
    <img align="right" width="30%" src="docs/images/sec-console.png" alt="ORY Security Console">
</a>

### ORY Security Console: Administrative User Interface

The [ORY Security Console](https://console.ory.sh/) is a visual admin interface
for managing ORY Hydra, ORY Oathkeeper, and ORY Keto.

### ORY Hydra: OAuth2 & OpenID Connect Server

[ORY Hydra](https://github.com/ory/hydra) ORY Hydra is a hardened OAuth2 and
OpenID Connect server optimized for low-latency, high throughput, and low
resource consumption. ORY Hydra is not an identity provider (user sign up, user
log in, password reset flow), but connects to your existing identity provider
through a consent app.

### ORY Keto: Access Control Policies as a Server

[ORY Keto](https://github.com/ory/keto) is a policy decision point. It uses a
set of access control policies, similar to AWS IAM Policies, in order to
determine whether a subject (user, application, service, car, ...) is authorized
to perform a certain action on a resource.

### Examples

The [ory/examples](https://github.com/ory/examples) repository contains numerous
examples of setting up this project individually and together with other
services from the ORY Ecosystem.

## Security

### Disclosing vulnerabilities

If you think you found a security vulnerability, please refrain from posting it
publicly on the forums, the chat, or GitHub and send us an email to
[hi@ory.am](mailto:hi@ory.am) instead.

## Telemetry

Our services collect summarized, anonymized data which can optionally be turned
off. Click [here](https://www.ory.sh/docs/guides/master/telemetry/) to learn
more.

## Documentation

### Guide

The Guide is available
[here](https://www.ory.sh/docs/guides/master/oathkeeper/).

### HTTP API documentation

The HTTP API is documented
[here](https://www.ory.sh/docs/api/oathkeeper?version=master).

### Upgrading and Changelog

New releases might introduce breaking changes. To help you identify and
incorporate those changes, we document these changes in
[UPGRADE.md](./UPGRADE.md) and [CHANGELOG.md](./CHANGELOG.md).

### Command line documentation

Run `oathkeeper -h` or `oathkeeper help`.

### Develop

Developing with ORY Oathkeeper is as easy as:

```shell
$ cd ~
$ go get -d -u github.com/ory/oathkeeper
$ cd $GOPATH/src/github.com/ory/oathkeeper
$ export GO111MODULE=on
$ go test ./...
```
