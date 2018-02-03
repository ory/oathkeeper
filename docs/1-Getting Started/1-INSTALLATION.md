# Installing ORY Oathkeeper

<!-- toc -->

The only way to install ORY Oathkeeper currently is through [Docker Hub]. ORY Oathkeeper is a private docker repository
and to gain access to it you have create an ORY Gatekeeper instance at the [ORY Admin Console](https://admin.ory.am).

Once your account is authorized to access the repository, log in to docker using `docker login`. Then you should
be able to fetch the ory/oathkeeper image and also be able to see it on [Docker Hub](https://hub.docker.com/r/oryd/oathkeeper/).

To run any oathkeeper command, do:

```
$ docker run oryd/oathkeeper:v<version> <command>
```

for example (this guide is written for ORY Oathkeeper 0.0.20:

```
$ docker run oryd/oathkeeper:v0.0.20 help
```
