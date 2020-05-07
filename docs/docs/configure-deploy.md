---
id: configure-deploy
title: Configure and Deploy
---

import useBaseUrl from '@docusaurus/useBaseUrl'

The ORY Oathkeeper HTTP serve process `oathkeeper serve` opens two ports
exposing the

- [reverse proxy](index.md#reverse-proxy)
- REST API which serves the
  [Access Control Decision API](index.md#access-control-decision-api) as well as
  other API endpoints such as health checks, JSON Web Key Sets, and a list of
  available rules.

For this guide we are using Docker. ORY Oathkeeper however can be
[installed in a variety of ways](install.md).

## Configure

ORY Oathkeeper can be configured via the filesystem as well as environment
variables. For more information on mapping the keys to environment variables
please head over to the [configuration chapter](configuration.md).

First, create an empty directory and `cd` into it:

```shell
$ mkdir oathkeeper-demo
$ cd oathkeeper-demo
```

Create a file called `config.yaml` with the following content:

```shell
$ cat << EOF > config.yaml
serve:
  proxy:
    port: 4455 # run the proxy at port 4455
  api:
    port: 4456 # run the api at port 4456

access_rules:
  repositories:
    - file:///rules.json

errors:
  fallback:
    - json
  handlers:
    json:
      enabled: true
      config:
        verbose: true
    redirect:
      enabled: true
      config:
        to: https://www.ory.sh/docs

mutators:
  header:
    enabled: true
    config:
      headers:
        X-User: "{{ print .Subject }}"
        # You could add some other headers, for example with data from the
        # session.
        # X-Some-Arbitrary-Data: "{{ print .Extra.some.arbitrary.data }}"
  noop:
    enabled: true
  id_token:
    enabled: true
    config:
      issuer_url: http://localhost:4455/
      jwks_url: file:///jwks.json

authorizers:
  allow:
    enabled: true
  deny:
    enabled: true

authenticators:
  anonymous:
    enabled: true
    config:
      subject: guest
EOF
```

This configuration file will run the proxy at port 4455, the api at port 4456,
and enable the anonymous authenticator, the allow and deny authorizers, and the
noop and id_token mutators.

### Access Rules

We will be using [httpbin.org](https://httpbin.org) as the upstream server. The
service echoes incoming HTTP Requests and is perfect for seeing how ORY
Oathkeeper works. Let's define three rules:

1. An access rule that allowing anonymous access to
   `https://httpbin.org/anything/cookie` and using the `cookie` mutator.
2. An access rule denying every access to `https://httpbin.org/anything/deny`.
   If the request header has `Accept: application/json`, we will receive a JSON
   response. If however the accept header has `Accept: text/*`, a HTTP Redirect
   will be sent (to `https://www.ory.sh/docs` as configured above).
3. An access rule allowing anonymous access to
   `https://httpbin.org/anything/id_token` and using the `id_token` mutator.

```shell
$ cat << EOF > rules.json
[
  {
    "id": "allow-anonymous-with-header-mutator",
    "version": "v0.36.0-beta.4",
    "upstream": {
      "url": "https://httpbin.org/anything/header"
    },
    "match": {
      "url": "http://<127.0.0.1|localhost>:4455/anything/header",
      "methods": [
        "GET"
      ]
    },
    "authenticators": [
      {
        "handler": "anonymous"
      }
    ],
    "authorizer": {
      "handler": "allow"
    },
    "mutators": [
      {
        "handler": "header",
        "config": {
          "headers": {
            "X-User": "{{ print .Subject }}"
          }
        }
      }
    ]
  },
  {
    "id": "deny-anonymous",
    "version": "v0.36.0-beta.4",
    "upstream": {
      "url": "https://httpbin.org/anything/deny"
    },
    "match": {
      "url": "http://<127.0.0.1|localhost>:4455/anything/deny",
      "methods": [
        "GET"
      ]
    },
    "authenticators": [
      {
        "handler": "anonymous"
      }
    ],
    "authorizer": {
      "handler": "deny"
    },
    "mutators": [
      {
        "handler": "noop"
      }
    ],
    "errors": [
      {
        "handler": "json",
        "config": {
          "when": [
            {
              "request": {
                "header": {
                  "accept": ["application/json"]
                }
              }
            }
          ]
        }
      },
      {
        "handler": "redirect",
        "config": {
          "when": [
            {
              "request": {
                "header": {
                  "accept": ["text/*"]
                }
              }
            }
          ]
        }
      }
    ]
  },
  {
    "id": "allow-anonymous-with-id-token-mutator",
    "version": "v0.36.0-beta.4",
    "upstream": {
      "url": "https://httpbin.org/anything/id_token"
    },
    "match": {
      "url": "http://<127.0.0.1|localhost>:4455/anything/id_token",
      "methods": [
        "GET"
      ]
    },
    "authenticators": [
      {
        "handler": "anonymous"
      }
    ],
    "authorizer": {
      "handler": "allow"
    },
    "mutators": [
      {
        "handler": "id_token"
      }
    ]
  }
]
EOF
```

### Cryptographic Keys

The `id_token` mutator creates a signed JSON Web Token. For that to work, a
public/private key is required. Luckily, ORY Oathkeeper can assist you in
creating such keys. All common JWT algorithms are supported (RS256, ES256,
HS256, ...). Let's generate a key for the RS256 algorithm that will be used by
the id_token mutator:

```sh
$ docker run oryd/oathkeeper:v0.38.0-beta.2 credentials generate --alg RS256 > jwks.json
```

### Dockerfile

Next we will be creating a custom Docker Image that adds these configuration
files to the image:

```shell
$ cat << EOF > Dockerfile
FROM oryd/oathkeeper:v0.38.0-beta.2

ADD config.yaml /config.yaml
ADD rules.json /rules.json
ADD jwks.json /jwks.json
EOF
```

We are doing this for demonstration purposes only. In a production environment
you would separate these configuration values from the build artifact itself. In
Kuberentes, it would make most sense to provide the JSON Web Keys as a
Kubernetes Secret mounted as in a directory, for example.

We encourage you to check out our [helm charts](https://k8s.ory.sh/helm/) which
apply these best practices.

## Build & Run

Before building the Docker Image, we need to make sure that the local ORY
Oathkeeper Docker Image is on the most recent version:

```sh
$ docker pull oryd/oathkeeper:v0.38.0-beta.2
```

Next we will build our custom Docker Image

```sh
$ docker build -t ory-oathkeeper-demo .
```

and run it

```
$ docker run --rm \
  --name ory-oathkeeper-demo \
  -p 4455:4455 \
  -p 4456:4456 \
  ory-oathkeeper-demo \
  --config /config.yaml \
  serve
```

Let's open a new terminal and check if it is alive:

```
$ curl http://127.0.0.1:4456/health/alive
{"status":"ok"}

$ curl http://127.0.0.1:4456/health/ready
{"status":"ok"}
```

Let's also check if the rules have been imported properly:

```
$ curl http://127.0.0.1:4456/rules
[{"id":"allow-anonymous-with-header-mutator","description":"","match":{"methods":["GET"],...
```

## Authorizing Requests

Everything is up and running and configured! Let's make some requests:

```
$ curl -X GET http://127.0.0.1:4455/anything/header
{
  "args": {},
  "data": "",
  "files": {},
  "form": {},
  "headers": {
    "Accept": "*/*",
    "Accept-Encoding": "gzip",
    "Host": "httpbin.org",
    "User-Agent": "curl/7.54.0",
    "X-User": "guest"
  },
  "json": null,
  "method": "GET",
  "origin": "172.17.0.1, 82.135.11.242, 172.17.0.1",
  "url": "https://httpbin.org/anything/header/anything/header"
}

# Make request and accept JSON (we get an error response)
$ curl -H "Accept: application/json" -X GET http://127.0.0.1:4455/anything/deny
{
  "error":{
    "code":403,
    "status":"Forbidden",
    "message":"Access credentials are not sufficient to access this resource"
  }
}

# Make request and accept text/* (we get a redirect response).
$ curl -H "Accept: text/html" -X GET http://127.0.0.1:4455/anything/deny
<a href="https://www.ory.sh/docs">Found</a>.

$ curl -X GET http://127.0.0.1:4455/anything/id_token
{
  "args": {},
  "data": "",
  "files": {},
  "form": {},
  "headers": {
    "Accept": "*/*",
    "Accept-Encoding": "gzip",
    "Authorization": "Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjU3N2E2NWE0LTUzM2YtNDFhYi1hODI2LTgxNDliMDM2NDQ0MyIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NTgwMTg1MTcsImlhdCI6MTU1ODAxODQ1NywiaXNzIjoiaHR0cDovL2xvY2FsaG9zdDo0NDU1LyIsImp0aSI6IjExNmRiNzhmLTQyMjEtNDU2ZC05OWIzLTY4NGJkMWVjYThjZSIsIm5iZiI6MTU1ODAxODQ1Nywic3ViIjoiZ3Vlc3QifQ.2VKW-oYtzkFGRPgK3sb4iRlObDSzW8PyHzgNiQubppFSlp0bzJLl4Rnt56orJndPqIa7hwsm8YIskf-Wp-FA1piv-aG_XljkUjgilKr3cncMXDP15yDRwZj8g0iVKEhnugQsw_zWf5gMU2YBev2Eyv4xciJxbhrKCat-X8xNT9SvAbwpY-VxQdu_rnpu1GKCA54DyIX6r-Qh5bQPrrT7NvIupA7jJQ23qq83m4C1cQfBgzlhm7dcCuPqKunYKRsc7NZuER3lT6TjkhsF1qhf7o7BZmCnhz6VuH8L8TwMZS8IJWKSjJd8dEKKwxwPkNXOcZO8A3hIO8SZx4Yd7jrONA",
    "Host": "httpbin.org",
    "User-Agent": "curl/7.54.0"
  },
  "json": null,
  "method": "GET",
  "origin": "172.17.0.1, 82.135.11.242, 172.17.0.1",
  "url": "https://httpbin.org/anything/id_token/anything/id_token"
}
```

That's it! You can now clean up the demo using:

```
$ docker rm -f ory-oathkeeper-demo
$ docker rmi -f ory-oathkeeper-demo
$ rm -rf oathkeeper-demo
```

## Monitoring

Oathkeeper provides an endpoint for Prometheus to scrape as a target. This
endpoint can be accessed by default at:
[http://localhost:9000/metrics](http://localhost:9000/metrics):

You can adjust the settings within Oathkeeper's config.

```shell
$ cat << EOF > config.yaml
serve:
  prometheus:
    port: 9000
    host: localhost
    metrics_path: /metrics
EOF
```

Prometheus can easily be run as a docker container. More information are
available on
[https://github.com/prometheus/prometheus](https://github.com/prometheus/prometheus).
Start with setting up a prometheus configuration:

```shell
$ cat << EOF > prometheus.yml
global:
  scrape_interval: 15s # By default, scrape targets every 15 seconds.

scrape_configs:
  - job_name: 'prometheus'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:9090']
  - job_name: 'oathkeeper'
    scrape_interval: 15s
    metrics_path: /metrics
    static_configs:
      # The target needs to match what you've configured above
      - targets: ['localhost:9000']
```

Then start the prometheus server and access it on
[http://localhost:9090](http://localhost:9090).

```shell
$ docker run \
  --config.file=/etc/prometheus/prometheus.yml \
  -v ./prometheus.yml:/etc/prometheus/prometheus.yml \
  --name prometheus \
  -d \
  --net=host
  -p 9090:9090 \
  prom/prometheus
```

Now where you have a basic monitoring setup running you can extend it by
building up nice visualizations eg. using Grafana. More information are
available on
[https://prometheus.io/docs/visualization/grafana/](https://prometheus.io/docs/visualization/grafana/).

We have a pre built Dashboard which you can use to get started quickly:
[Oathkeeper-Dashboard.json](https://github.com/ory/oathkeeper/tree/master/contrib/grafana/Oathkeeper-Dashboard.json).

<img alt="ORY Oathkeeper with Prometheus and Grafana"
src={useBaseUrl('img/docs/grafana.png')} />
