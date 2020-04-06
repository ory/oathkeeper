---
id: install
title: Installation
---

Installing ORY Oathkeeper on any system is straight forward. We provide
pre-built binaries, Docker Images and support various package managers.

## Docker

We recommend using Docker to run ORY Oathkeeper:

```shell
$ docker pull oryd/oathkeeper
$ docker run --rm -it oryd/oathkeeper help
```

## macOS

You can install ORY Oathkeeper using [homebrew](https://brew.sh/) on macOS:

```shell
$ brew tap ory/oathkeeper
$ brew install ory/oathkeeper/oathkeeper
$ oathkeeper help
```

## Linux

On linux, you can use `curl | bash` to fetch the latest stable binary using:

```shell
$ curl https://raw.githubusercontent.com/ory/oathkeeper/master/install.sh | bash -s -- -b .
$ ./oathkeeper help
```

You may want to move ORY Oathkeeper to your `$PATH`:

```shell
$ sudo mv ./oathkeeper /usr/local/bin/
$ oathkeeper help
```

## Windows

You can install ORY Oathkeeper using [scoop](https://scoop.sh) on Windows:

```shell
> scoop bucket add ory-oathkeeper https://github.com/ory/scoop-oathkeeper.git
> scoop install oathkeeper
> oathkeeper help
```

## Download Binaries

The client and server **binaries are downloadable at the
[releases tab](https://github.com/ory/oathkeeper/releases)**. There is currently
no installer available. You have to add the Oathkeeper binary to the PATH
environment variable yourself or put the binary in a location that is already in
your `$PATH` (e.g. `/usr/local/bin`, ...).

Once installed, you should be able to run:

```shell
$ oathkeeper help
```

## Building from Source

If you wish to compile ORY Oathkeeper yourself, you need to install and set up
[Go 1.12+](https://golang.org/) and add `$GOPATH/bin` to your `$PATH`.

The following commands will check out the latest release tag of ORY Oathkeeper
and compile it and set up flags so that `oathkeeper version` works as expected.
Please note that this will only work with a linux shell like bash or sh.

```shell
$ go get -d -u github.com/ory/oathkeeper
$ cd $(go env GOPATH)/src/github.com/ory/oathkeeper
$ GO111MODULE=on make install-stable
$ $(go env GOPATH)/bin/oathkeeper help
```
