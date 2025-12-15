# Development

This document explains how to develop Ory Oathkeeper, run tests, and work with
the tooling around it.

## Upgrading and changelog

New releases might introduce breaking changes. To help you identify and
incorporate those changes, we document these changes in
[UPGRADE.md](./UPGRADE.md) and [CHANGELOG.md](./CHANGELOG.md).

## Command line documentation

To see available commands and flags, run:

```shell
oathkeeper -h
# or
oathkeeper help
```

## Contribution guidelines

We encourage all contributions. Before opening a pull request, read the
[contribution guidelines](./CONTRIBUTING.md).

## Prerequisites

You need Go 1.16+ and, for the test suites:

- Docker and Docker Compose
- `make`
- Node.js and npm

You can develop Ory Oathkeeper on Windows, but most guides assume a Unix shell
such as `bash` or `zsh`.

## Install from source

To install Oathkeeper from source:

```shell
cd ~
go get -d -u github.com/ory/oathkeeper
cd $GOPATH/src/github.com/ory/oathkeeper
export GO111MODULE=on
make install
```

## Formatting code

Format all code using:

```shell
make format
```

The continuous integration pipeline checks code formatting.

## Running tests

There are three types of tests:

- Short tests that do not require a SQL database
- Regular tests that require PostgreSQL, MySQL, and CockroachDB
- End to end tests that use real databases and a test browser

### Short tests

Short tests run quickly and use SQLite.

Run all short tests:

```shell
go test -short -tags sqlite ./...
```

Run short tests in a specific module:

```shell
cd internal/check
go test -short -tags sqlite .
```

## Build Docker image

To build a development Docker image:

```shell
make docker
```
