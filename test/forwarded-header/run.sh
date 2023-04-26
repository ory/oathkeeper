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

make_request() {
  url=$1
  expected_status_code=$2
  shift 2

  [[ $(curl --silent --output /dev/null -f ${url} -w '%{http_code}' "$@") -eq $expected_status_code ]]
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

run_test "Executing request against a HTTP rule -> 200" \
  make_request "http://127.0.0.1:6060/http" 200

run_test "Executing request against a HTTP rule with forwrded proto HTTP -> 200" \
  make_request "http://127.0.0.1:6060/http" 200 -H "X-Forwarded-Proto: http"

run_test "Executing request against a HTTP rule with forwrded proto HTTPS -> 404" \
  make_request "http://127.0.0.1:6060/http" 404 -H "X-Forwarded-Proto: https"

run_test "Executing request against a HTTPS rule -> 404" \
  make_request "http://127.0.0.1:6060/https" 404

run_test "Executing request against a HTTPS rule with forwarded proto HTTP -> 404" \
  make_request "http://127.0.0.1:6060/https" 404 -H "X-Forwarded-Proto: http"

run_test "Executing request against a HTTPS rule with forwarded proto HTTPS -> 200" \
  make_request "http://127.0.0.1:6060/https" 200 -H "X-Forwarded-Proto: https"


echo "PASS: ${#SUCCESS_TEST[@]}"
for value in "${SUCCESS_TEST[@]}"
do
  echo "- $value"
done

if [[ "${#FAILED_TEST[@]}" -gt 0 ]]; then
  echo "FAILED: ${#FAILED_TEST[@]}"
  for value in "${FAILED_TEST[@]}"
  do
    echo "- $value"
  done

  exit 1
fi

kill %1 || true

trap - EXIT
