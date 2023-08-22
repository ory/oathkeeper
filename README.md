<h1 align="center"><img src="https://raw.githubusercontent.com/ory/meta/master/static/banners/oathkeeper.svg" alt="ORY Oathkeeper - Cloud Native Identity & Access Proxy"></h1>

<h4 align="center">
    <a href="https://www.ory.sh/chat">Chat</a> |
    <a href="https://github.com/ory/oathkeeper/discussions">Discussions</a> |
    <a href="http://eepurl.com/di390P">Newsletter</a><br/><br/>
    <a href="https://www.ory.sh/oathkeeper/docs/reference/api">API Docs</a> |
    <a href="https://www.ory.sh/oathkeeper/docs/">Guide</a> |
    <a href="https://godoc.org/github.com/ory/oathkeeper">Code Docs</a><br/><br/>
    <a href="https://opencollective.com/ory">Support this project!</a><br/><br/>
    <a href="https://www.ory.sh/jobs/">Work in Open Source, Ory is hiring!</a>
</h4>

---

<p align="left">
    <a href="https://github.com/ory/oathkeeper/actions/workflows/ci.yml"><img src="https://github.com/ory/oathkeeper/actions/workflows/ci.yml/badge.svg" alt="Build Status"></a>
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

## Ory Network Hybrid Support Plan

Ory offers a support plan for Ory Network Hybrid, including Ory on private cloud
deployments. If you have a self-hosted solution and would like help, consider a
support plan! The team at Ory has years of experience in cloud computing. Ory's
offering is the only official program for qualified support from the
maintainers. For more information see the
**[website](https://www.ory.sh/support/)** or
**[book a meeting](https://www.ory.sh/contact/)**!

## Project Renaming

The Ory Oathkeeper project was started in 2017 in Germany and owes its name to
the Sword [Oathkeeper](https://gameofthrones.fandom.com/wiki/Oathkeeper) from
Game of Thrones. We also understand that the name is politically charged in the
US as it is shared with a far-right militia organization in the US called "Oath
Keepers".

To take a stand against extremism and avoid any confusion to the name's origin,
we will be renaming the project in the near future. Please be patient with us as
we work on this complicated change of various CIs, tools, scripts, and
automations.

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
maintainers. The Ory team thanks everyone involved - from submitting bug reports
and feature requests, to contributing patches and documentation. The Ory
community counts more than 33.000 members and is growing rapidly. The Ory stack
protects 60.000.000.000+ API requests every month with over 400.000+ active
service nodes. None of this would have been possible without each and everyone
of you!

The following list represents companies that have accompanied us along the way
and that have made outstanding contributions to our ecosystem. _If you think
that your company deserves a spot here, reach out to
<a href="mailto:office@ory.sh">office@ory.sh</a> now_!

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
            <td>Adopter *</td>
            <td>Raspberry PI Foundation</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/raspi.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/raspi.svg" alt="Raspberry PI Foundation">
                </picture>
            </td>
            <td><a href="https://www.raspberrypi.org/">raspberrypi.org</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Kyma Project</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/kyma.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/kyma.svg" alt="Kyma Project">
                </picture>
            </td>
            <td><a href="https://kyma-project.io">kyma-project.io</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Tulip</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/tulip.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/tulip.svg" alt="Tulip Retail">
                </picture>
            </td>
            <td><a href="https://tulip.com/">tulip.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Cashdeck / All My Funds</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/allmyfunds.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/allmyfunds.svg" alt="All My Funds">
                </picture>
            </td>
            <td><a href="https://cashdeck.com.au/">cashdeck.com.au</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Hootsuite</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/hootsuite.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/hootsuite.svg" alt="Hootsuite">
                </picture>
            </td>
            <td><a href="https://hootsuite.com/">hootsuite.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Segment</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/segment.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/segment.svg" alt="Segment">
                </picture>
            </td>
            <td><a href="https://segment.com/">segment.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Arduino</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/arduino.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/arduino.svg" alt="Arduino">
                </picture>
            </td>
            <td><a href="https://www.arduino.cc/">arduino.cc</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>DataDetect</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/datadetect.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/datadetect.svg" alt="Datadetect">
                </picture>
            </td>
            <td><a href="https://unifiedglobalarchiving.com/data-detect/">unifiedglobalarchiving.com/data-detect/</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Sainsbury's</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/sainsburys.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/sainsburys.svg" alt="Sainsbury's">
                </picture>
            </td>
            <td><a href="https://www.sainsburys.co.uk/">sainsburys.co.uk</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Contraste</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/contraste.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/contraste.svg" alt="Contraste">
                </picture>
            </td>
            <td><a href="https://www.contraste.com/en">contraste.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Reyah</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/reyah.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/reyah.svg" alt="Reyah">
                </picture>
            </td>
            <td><a href="https://reyah.eu/">reyah.eu</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Zero</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/commitzero.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/commitzero.svg" alt="Project Zero by Commit">
                </picture>
            </td>
            <td><a href="https://getzero.dev/">getzero.dev</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Padis</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/padis.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/padis.svg" alt="Padis">
                </picture>
            </td>
            <td><a href="https://padis.io/">padis.io</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Cloudbear</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/cloudbear.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/cloudbear.svg" alt="Cloudbear">
                </picture>
            </td>
            <td><a href="https://cloudbear.eu/">cloudbear.eu</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Security Onion Solutions</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/securityonion.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/securityonion.svg" alt="Security Onion Solutions">
                </picture>
            </td>
            <td><a href="https://securityonionsolutions.com/">securityonionsolutions.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Factly</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/factly.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/factly.svg" alt="Factly">
                </picture>
            </td>
            <td><a href="https://factlylabs.com/">factlylabs.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Nortal</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/nortal.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/nortal.svg" alt="Nortal">
                </picture>
            </td>
            <td><a href="https://nortal.com/">nortal.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>OrderMyGear</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/ordermygear.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/ordermygear.svg" alt="OrderMyGear">
                </picture>
            </td>
            <td><a href="https://www.ordermygear.com/">ordermygear.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Spiri.bo</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/spiribo.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/spiribo.svg" alt="Spiri.bo">
                </picture>
            </td>
            <td><a href="https://spiri.bo/">spiri.bo</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Strivacity</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/strivacity.svg" />
                    <img height="16px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/strivacity.svg" alt="Spiri.bo">
                </picture>
            </td>
            <td><a href="https://strivacity.com/">strivacity.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Hanko</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/hanko.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/hanko.svg" alt="Hanko">
                </picture>
            </td>
            <td><a href="https://hanko.io/">hanko.io</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Rabbit</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/rabbit.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/rabbit.svg" alt="Rabbit">
                </picture>
            </td>
            <td><a href="https://rabbit.co.th/">rabbit.co.th</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>inMusic</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/inmusic.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/inmusic.svg" alt="InMusic">
                </picture>
            </td>
            <td><a href="https://inmusicbrands.com/">inmusicbrands.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Buhta</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/buhta.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/buhta.svg" alt="Buhta">
                </picture>
            </td>
            <td><a href="https://buhta.com/">buhta.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Connctd</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/connctd.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/connctd.svg" alt="Connctd">
                </picture>
            </td>
            <td><a href="https://connctd.com/">connctd.com</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>Paralus</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/paralus.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/paralus.svg" alt="Paralus">
                </picture>
            </td>
            <td><a href="https://www.paralus.io/">paralus.io</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>TIER IV</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/tieriv.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/tieriv.svg" alt="TIER IV">
                </picture>
            </td>
            <td><a href="https://tier4.jp/en/">tier4.jp</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>R2Devops</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/r2devops.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/r2devops.svg" alt="R2Devops">
                </picture>
            </td>
            <td><a href="https://r2devops.io/">r2devops.io</a></td>
        </tr>
        <tr>
            <td>Adopter *</td>
            <td>LunaSec</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/lunasec.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/lunasec.svg" alt="LunaSec">
                </picture>
            </td>
            <td><a href="https://www.lunasec.io/">lunasec.io</a></td>
        </tr>
            <tr>
            <td>Adopter *</td>
            <td>Serlo</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/serlo.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/serlo.svg" alt="Serlo">
                </picture>
            </td>
            <td><a href="https://serlo.org/">serlo.org</a></td>
        </tr>
        </tr>
            <tr>
            <td>Adopter *</td>
            <td>dyrector.io</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/dyrector_io.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/dyrector_io.svg" alt="dyrector.io">
                </picture>
            </td>
            <td><a href="https://dyrector.io/">dyrector.io</a></td>
        </tr>
        </tr>
            <tr>
            <td>Adopter *</td>
            <td>Stackspin</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/stackspin.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/stackspin.svg" alt="stackspin.net">
                </picture>
            </td>
            <td><a href="https://www.stackspin.net/">stackspin.net</a></td>
        </tr>
        </tr>
            <tr>
            <td>Adopter *</td>
            <td>Amplitude</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/amplitude.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/amplitude.svg" alt="amplitude.com">
                </picture>
            </td>
            <td><a href="https://amplitude.com/">amplitude.com</a></td>
        </tr>
         <tr>
            <td>Adopter *</td>
            <td>Pinniped</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/pinniped.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/pinniped.svg" alt="pinniped.dev">
                </picture>
            </td>
            <td><a href="https://pinniped.dev/">pinniped.dev</a></td>
        </tr>         
        <tr>
            <td>Adopter *</td>
            <td>Pvotal</td>
            <td align="center">
                <picture>
                    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/ory/meta/master/static/adopters/light/pvotal.svg" />
                    <img height="32px" src="https://raw.githubusercontent.com/ory/meta/master/static/adopters/dark/pvotal.svg" alt="pvotal.tech">
                </picture>
            </td>
            <td><a href="https://pvotal.tech/">pvotal.tech</a></td>
        </tr>
    </tbody>
</table>

Many thanks to all individual contributors

<a href="https://opencollective.com/ory" target="_blank"><img src="https://opencollective.com/ory/contributors.svg?width=890&limit=714&button=false" /></a>

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

Ory's architecture is designed to run best on a Container Orchestration system
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
(MFA/2FA), Account Recovery and Verification, Profile, and Account Management.

### Ory Hydra: OAuth2 & OpenID Connect Server

[Ory Hydra](https://github.com/ory/hydra) is an OpenID Certified™ OAuth2 and
OpenID Connect Provider which easily connects to any existing identity system by
writing a tiny "bridge" application. It gives absolute control over the user
interface and user experience flows.

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
publicly on the forums, the chat, or GitHub. You can find all info for
responsible disclosure in our
[security.txt](https://www.ory.sh/.well-known/security.txt).

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
