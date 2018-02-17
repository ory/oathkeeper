# ORY Oathkeeper

Welcome to the ORY Oathkeeper documentation!

ORY Oathkeeper is a reverse proxy which evaluates incoming HTTP requests based on a set of rules. This capability
is referred to as an Identity and Access Proxy (IAP) in the context of [BeyondCorp and ZeroTrust](https://www.beyondcorp.com).

In principal, ORY Oathkeeper inspects the `Authorization` header and the full request url (e.g. `https://mydomain.com/api/foo`)
of incoming HTTP requests, applies a given rule, and either grants access to the requested url or denies access. The
decision of whether to allow or deny the access request is made using ORY Hydra. Please keep in mind that ORY Oathkeeper
is a supplement to ORY Hydra. It is thus imperative to be familiar with the core concepts of ORY Hydra.
