#!/bin/bash

set -euo pipefail

waitport() {
  i=0
  while ! nc -z localhost "$1" ; do
    sleep 1
    if [ $i -gt 10 ]; then
      cat ./config.yaml
      cat ./oathkeeper.log
      exit 1
    fi
    i=$((i+1))
  done
}

cd "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"


run_oathkekeper() {
  killall oathkeeper || true

  export OATHKEEPER_PROXY=http://127.0.0.1:6060
  export OATHKEEPER_API=http://127.0.0.1:6061

  go build -o . ../..
  LOG_LEVEL=debug ./oathkeeper --config ./config.yaml serve > ./oathkeeper.log 2>&1 &

  waitport 6060
  waitport 6061
}

run_api(){
  killall okapi || true

  PORT=6662 go run ./okapi > ./api.log 2>&1 &

  waitport 6662
}

SUCCESS_TEST=()
FAILED_TEST=()

run_test() {
  label=$1
  shift 1

  result="0"
  "$@" || result="1"

  if [[ "$result" -eq "0" ]]; then
    SUCCESS_TEST+=("$label")
  else
    FAILED_TEST+=("$label")
  fi
}

function finish {
  echo ::group::Config
  cat ./config.yaml
  cat ./rules.1.json
  echo ::endgroup::
  echo ::group::Log
  cat ./oathkeeper.log
  echo ::endgroup::
}
trap finish EXIT

run_oathkekeper
run_api

curl -X POST -f http://127.0.0.1:6060/test?token=token -F fk=fv -H "Content-Type: application/x-www-form-urlencoded" -i

kill %1 || true

trap - EXIT
