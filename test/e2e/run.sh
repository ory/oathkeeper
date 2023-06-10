#!/usr/bin/env bash

set -euxo pipefail

waitport() {
  i=0
  while ! nc -z localhost "$1" ; do
    sleep 1
    if [ $i -gt 10 ]; then
      cat ./oathkeeper.e2e.log
      echo "-----"
      cat ./api.e2e.log
      exit 1
    fi
    i=$((i+1))
  done
}

cd "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

killall oathkeeper || true
killall okapi || true
killall okclient || true

export OATHKEEPER_PROXY=http://127.0.0.1:6660
export OATHKEEPER_API=http://127.0.0.1:6661
export GO111MODULE=on

go build -o . ../..
./oathkeeper --config ./config.yml serve > ./oathkeeper.e2e.log 2>&1 &
PORT=6662 go run ./okapi > ./api.e2e.log 2>&1 &

waitport 6660
waitport 6661
waitport 6662

function finish {
  echo ::group::Oathkeeper Log
  cat ./oathkeeper.e2e.log
  echo ::endgroup::
  echo ::group::OK API Log
  cat ./api.e2e.log
  echo ::endgroup::
}
trap finish EXIT

go run ./okclient

kill %1 || true
kill %2 || true

trap - EXIT
