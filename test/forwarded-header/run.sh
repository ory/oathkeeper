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


export GO111MODULE=on

(cd ../../; make install)

run_oathkekeper() {
  killall oathkeeper || true

  export OATHKEEPER_PROXY=http://127.0.0.1:6060
  export OATHKEEPER_API=http://127.0.0.1:6061

  LOG_LEVEL=debug oathkeeper --config ./config.yaml serve >> ./oathkeeper.log 2>&1 &

  waitport 6060
  waitport 6061
}

function make_request {
  local url=$1
  local expected_status_code=$2
  shift 2

  if [[ $(curl --silent --output /dev/null -f ${url} -w '%{http_code}' "$@") -ne $expected_status_code ]]
  then
    exit 1
  fi
}

function finish {
  cat ./config.yaml
  cat ./rules.1.json
  cat ./oathkeeper.log
}
trap finish EXIT

run_oathkekeper

echo "Executing request against a HTTP rule -> 200"
make_request "http://127.0.0.1:6060/http" 200

echo "Executing request against a HTTP rule with forwrded proto HTTP -> 200"
make_request "http://127.0.0.1:6060/http" 200 -H "X-Forwarded-Proto: http"

echo "Executing request against a HTTP rule with forwrded proto HTTPS -> 404"
make_request "http://127.0.0.1:6060/http" 404 -H "X-Forwarded-Proto: https"

echo "Executing request against a HTTPS rule -> 404"
make_request "http://127.0.0.1:6060/https" 404

echo "Executing request against a HTTPS rule with forwarded proto HTTP -> 404"
make_request "http://127.0.0.1:6060/https" 404 -H "X-Forwarded-Proto: http"

echo "Executing request against a HTTPS rule with forwarded proto HTTPS -> 200"
make_request "http://127.0.0.1:6060/https" 200 -H "X-Forwarded-Proto: https"

kill %1 || true

trap - EXIT
