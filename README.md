<h1 align="center"><img src="https://raw.githubusercontent.com/ory/meta/master/static/banners/oathkeeper.svg" alt="ORY Oathkeeper - Cloud Native Identity & Access Proxy"></h1>

<h4 align="center">
    <a href="https://www.ory.sh/chat">Chat</a> |
    <a href="https://github.com/ory/oathkeeper/discussions">Discussions</a> |
    <a href="http://eepurl.com/di390P">Newsletter</a><br/><br/>
    <a href="https://www.ory.sh/oathkeeper/docs/reference/api">API Docs</a> |
    <a href="https://www.ory.sh/oathkeeper/docs/">Guide</a> |
    <a href="https://godoc.org/github.com/ory/oathkeeper">Code Docs</a><br/><br/>
    <a href="https://opencollective.com/ory">Support this project!</a>
</h4>

---

<p align="left">
    <a href="https://circleci.com/gh/ory/oathkeeper/tree/master"><img src="https://circleci.com/gh/ory/oathkeeper.svg?style=shield" alt="Build Status"></a>
    <a href="https://coveralls.io/github/ory/oathkeeper?branch=master"> <img src="https://coveralls.io/repos/ory/oathkeeper/badge.svg?branch=master&service=github" alt="Coverage Status"></a>
    <a href="https://goreportcard.com/report/github.com/ory/oathkeeper"><img src="https://goreportcard.com/badge/github.com/ory/oathkeeper" alt="Go Report Card"></a>
    <a href="https://pkg.go.dev/github.com/ory/oathkeeper"><img src="https://pkg.go.dev/badge/www.github.com/ory/oathkeeper" alt="PkgGoDev"></a>
    <a href="#backers" alt="sponsors on Open Collective"><img src="https://opencollective.com/ory/backers/badge.svg" /></a> <a href="#sponsors" alt="Sponsors on Open Collective"><img src="https://opencollective.com/ory/sponsors/badge.svg" /></a>
    <a href="https://github.com/ory/oathkeeper/blob/master/CODE_OF_CONDUCT.md" alt="Ory Code of Conduct"><img src="https://img.shields.io/badge/ory-code%20of%20conduct-green" /></a>
</p>

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
  [External Authorization HTTP Filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter)
- [AWS API Gateway](https://aws.amazon.com/api-gateway/) via
  [Custom Authorizers](https://aws.amazon.com/de/blogs/compute/introducing-custom-authorizers-in-amazon-api-gateway/)
- [Nginx](https://www.nginx.com) via
  [Authentication Based on Subrequest Result](https://docs.nginx.com/nginx/admin-guide/security-controls/configuring-subrequest-authentication/)

among others.

This service is stable, but under active development and may introduce breaking
changes in future releases. Any breaking change will have extensive
documentation and upgrade instructions.

---

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Installation](#installation)
- [Who's using it?](#whos-using-it)
- [Ecosystem](#ecosystem)
  - [ORY Kratos: Identity and User Infrastructure and Management](#ory-kratos-identity-and-user-infrastructure-and-management)
  - [ORY Hydra: OAuth2 & OpenID Connect Server](#ory-hydra-oauth2--openid-connect-server)
  - [ORY Oathkeeper: Identity & Access Proxy](#ory-oathkeeper-identity--access-proxy)
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

Head over to the
[ORY Developer Documentation](https://www.ory.sh/oathkeeper/docs/install) to
learn how to install ORY Oathkeeper on Linux, macOS, Windows, and Docker and how
to build ORY Oathkeeper from source.

## Who's using it?

<!--BEGIN ADOPTERS-->

The Ory community stands on the shoulders of individuals, companies, and
maintainers. We thank everyone involved - from submitting bug reports and
feature requests, to contributing patches, to sponsoring our work. Our community
is 1000+ strong and growing rapidly. The Ory stack protects 16.000.000.000+ API
requests every month with over 250.000+ active service nodes. We would have
never been able to achieve this without each and everyone of you!

The following list represents companies that have accompanied us along the way
and that have made outstanding contributions to our ecosystem. _If you think
that your company deserves a spot here, reach out to
<a href="mailto:office-muc@ory.sh">office-muc@ory.sh</a> now_!

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
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/raspi.svg" alt="Raspberry PI Foundation"></td>
            <td><a href="https://www.raspberrypi.org/">raspberrypi.org</a></td>
        </tr>
        <tr>
            <td>Contributor</td>
            <td>Kyma Project</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/kyma.svg" alt="Kyma Project"></td>
            <td><a href="https://kyma-project.io">kyma-project.io</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>ThoughtWorks</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/tw.svg" alt="ThoughtWorks"></td>
            <td><a href="https://www.thoughtworks.com/">thoughtworks.com</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Tulip</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/tulip.svg" alt="Tulip Retail"></td>
            <td><a href="https://tulip.com/">tulip.com</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Cashdeck / All My Funds</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/allmyfunds.svg" alt="All My Funds"></td>
            <td><a href="https://cashdeck.com.au/">cashdeck.com.au</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>3Rein</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/3R-horiz.svg" alt="3Rein"></td>
            <td><a href="https://3rein.com/">3rein.com</a></td>
        </tr>
        <tr>
            <td>Contributor</td>
            <td>Hootsuite</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/hootsuite.svg" alt="Hootsuite"></td>
            <td><a href="https://hootsuite.com/">hootsuite.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Segment</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/segment.svg" alt="Segment"></td>
            <td><a href="https://segment.com/">segment.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Arduino</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/arduino.svg" alt="Arduino"></td>
            <td><a href="https://www.arduino.cc/">arduino.cc</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>DataDetect</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/datadetect.svg" alt="Datadetect"></td>
            <td><a href="https://unifiedglobalarchiving.com/data-detect/">unifiedglobalarchiving.com/data-detect/</a></td>
        </tr>        
        <tr>
            <td>Adopter *</td>
            <td>Sainsbury's</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/sainsburys.svg" alt="Sainsbury's"></td>
            <td><a href="https://www.sainsburys.co.uk/">sainsburys.co.uk</a></td>
        </tr>
                <tr>
            <td>Adopter *</td>
            <td>Contraste</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/contraste.svg" alt="Contraste"></td>
            <td><a href="https://www.contraste.com/en">contraste.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Reyah</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/reyah.svg" alt="Reyah"></td>
            <td><a href="https://reyah.eu/">reyah.eu</a></td>
        </tr>        
        <tr>
            <td>Adopter *</td>
            <td>Zero</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/commitzero.svg" alt="Project Zero by Commit"></td>
            <td><a href="https://getzero.dev/">getzero.dev</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>OrderMyGear</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/ordermygear.svg" alt="OrderMyGear"></td>
            <td><a href="https://www.ordermygear.com/">ordermygear.com</a></td>
        </tr>
        <tr>
            <td>Sponsor</td>
            <td>Spiri.bo</td>
            <td align="center"><img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/spiribo.svg" alt="Spiri.bo"></td>
            <td><a href="https://spiri.bo/">spiri.bo</a></td>
        </tr>
    </tbody>
</table>

We also want to thank all individual contributors

<a href="https://opencollective.com/ory" target="_blank"><img src="https://opencollective.com/ory/contributors.svg?width=890&button=false" /></a>

as well as all of our backers

<a href="https://opencollective.com/ory#backers" target="_blank"><img src="https://opencollective.com/ory/backers.svg?width=890"></a>

and past & current supporters (in alphabetical order) on
[Patreon](https://www.patreon.com/_ory): Alexander Alimovs, Billy, Chancy
Kennedy, Drozzy, Edwin Trejos, Howard Edidin, Ken Adler Oz Haven, Stefan Hans,
TheCrealm.

<em>\* Uses one of Ory's major projects in production.</em>

<!--END ADOPTERS-->

## Ecosystem

<!--BEGIN ECOSYSTEM-->

We build Ory on several guiding principles when it comes to our architecture
design:

- Minimal dependencies
- Runs everywhere
- Scales without effort
- Minimize room for human and network errors

Ory's architecture designed to run best on a Container Orchestration Systems
such as Kubernetes, CloudFoundry, OpenShift, and similar projects. Binaries are
small (5-15MB) and available for all popular processor types (ARM, AMD64, i386)
and operating systems (FreeBSD, Linux, macOS, Windows) without system
dependencies (Java, Node, Ruby, libxml, ...).

### Ory Kratos: Identity and User Infrastructure and Management

[Ory Kratos](https://github.com/ory/kratos) is an API-first Identity and User
Management system that is built according to
[cloud architecture best practices](https://www.ory.sh/docs/next/ecosystem/software-architecture-philosophy).
It implements core use cases that almost every software application needs to
deal with: Self-service Login and Registration, Multi-Factor Authentication
(MFA/2FA), Account Recovery and Verification, Profile and Account Management.

### Ory Hydra: OAuth2 & OpenID Connect Server

[Ory Hydra](https://github.com/ory/hydra) is an OpenID Certified™ OAuth2 and
OpenID Connect Provider which easily connects to any existing identity system by
writing a tiny "bridge" application. Gives absolute control over user interface
and user experience flows.

### Ory Oathkeeper: Identity & Access Proxy

[Ory Oathkeeper](https://github.com/ory/oathkeeper) is a BeyondCorp/Zero Trust
Identity & Access Proxy (IAP) with configurable authentication, authorization,
and request mutation rules for your web services: Authenticate JWT, Access
Tokens, API Keys, mTLS; Check if the contained subject is allowed to perform the
request; Encode resulting content into custom headers (`X-User-ID`), JSON Web
Tokens and more!

### Ory Keto: Access Control Policies as a Server

[Ory Keto](https://github.com/ory/keto) is a policy decision point. It uses a
set of access control policies, similar to AWS IAM Policies, in order to
determine whether a subject (user, application, service, car, ...) is authorized
to perform a certain action on a resource.

<!--END ECOSYSTEM-->

## Security

### Disclosing vulnerabilities

If you think you found a security vulnerability, please refrain from posting it
publicly on the forums, the chat, or GitHub and send us an email to
[hi@ory.sh](mailto:hi@ory.sh) instead.

## Telemetry

Our services collect summarized, anonymized data which can optionally be turned
off. Click [here](https://www.ory.sh/docs/ecosystem/sqa) to learn more.

## Documentation

### Guide

The Guide is available [here](https://www.ory.sh/oathkeeper/docs/).

### HTTP API documentation

The HTTP API is documented
[here](https://www.ory.sh/oathkeeper/docs/reference/api).

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
