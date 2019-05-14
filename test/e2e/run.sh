#!/bin/bash

set -euxo pipefail

waitport() {
    while ! nc -z localhost $1 ; do sleep 1 ; done
}

cd "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

killall oathkeeper || true
killall okapi || true
killall okclient || true

export OATHKEEPER_PROXY=http://127.0.0.1:6660
export OATHKEEPER_API=http://127.0.0.1:6661
export GO111MODULE=on

go install github.com/ory/oathkeeper
go install github.com/ory/oathkeeper/test/e2e/okapi
go install github.com/ory/oathkeeper/test/e2e/okclient

oathkeeper --config ./config.yml serve >> ./oathkeeper.e2e.log 2>&1 &
PORT=6662 okapi >> ./api.e2e.log 2>&1 &

waitport 6660
waitport 6661
waitport 6662

okclient

kill %1 || true
kill %2 || true
