---
id: traefik-proxy-integration
title: Traefik Proxy Integration
---

[Traefik Proxy](https://doc.traefik.io/traefik/) is modern HTTP proxy and load balancer for microservices, oathkeeper can be integrated with via the [ForwardAuth Middleware](https://doc.traefik.io/traefik/middlewares/http/forwardauth/) by making use of the available [Access Control Decision API](index.md#access-control-decision-api).

To achieve this,
* configure traefik
  * to make use of the aforesaid ForwardAuth middleware by setting the `address` property to the decision URL endpoint and
  * by including the required header name(s), the oathkeeper sets in the HTTP responses into the `authResponseHeaders` property.
* configure the route of your service to make use of this middleware

Example (using Docker labels):

```.yaml
edge-router:
  image: traefik
  # further configuration
  labels:
    - traefik.http.middlewares.oathkeeper.forwardauth.address=http://oathkeeper:4456/decisions
    - traefik.http.middlewares.oathkeeper.forwardauth.authResponseHeaders=X-Id-Token,Authorization
    # further labels

service:
  image: my-service
  # further configuration
  labels:
    - traefik.http.routers.service.middlewares=oathkeeper
    # further labels
```



