#!/usr/bin/env bash

set -euxo pipefail

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

killall oathkeeper || true

export OATHKEEPER_PROXY=http://127.0.0.1:6060
export OATHKEEPER_API=http://127.0.0.1:6061
export GO111MODULE=on

cp config.1.yaml config.yaml

go build -o . ../..
LOG_LEVEL=debug ./oathkeeper --config ./config.yaml serve > ./oathkeeper.log 2>&1 &

waitport 6060
waitport 6061

function finish {
  killall oathkeeper || true
  echo ::group::Config
  cat ./config.yaml
  cat ./rules.3.json || true
  echo ::endgroup::
  echo ::group::Log
  cat ./oathkeeper.log
  echo ::endgroup::
}
trap finish EXIT

sleep 5

echo "Executing request against no configured rules -> 404"
[[ $(curl --retry 7 --retry-connrefused -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 404 ]] && exit 1

echo "Executing request against a now configured rule -> 200"
cp config.2.yaml config.yaml; sleep 3; [[ $(curl --retry 7 --retry-connrefused -s -o /dev/null -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 200 ]] && exit 1

echo "Executing request against updated rule with deny in it -> 403"
cp config.3.yaml config.yaml; sleep 3; [[ $(curl --retry 7 --retry-connrefused -s -o /dev/null -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 403 ]] && exit 1

echo "Executing request against updated rule with deny deny disabled -> 500"
cp config.4.yaml config.yaml; sleep 3; [[ $(curl --retry 7 --retry-connrefused -s -o /dev/null -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 500 ]] && exit 1

echo "Making sure route to be registered is not available yet -> 404"
[[ $(curl --retry 7 --retry-connrefused -s -o /dev/null -f ${OATHKEEPER_PROXY}/other-rules -w '%{http_code}') -ne 404 ]] && exit 1

echo "Load another rule set -> all 200"
cp rules.3.1.json rules.3.json
cp config.5.yaml config.yaml; sleep 3
[[ $(curl --retry 7 --retry-connrefused -s -o /dev/null -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 200 ]] && exit 1
[[ $(curl --retry 7 --retry-connrefused -s -o /dev/null -f ${OATHKEEPER_PROXY}/other-rules -w '%{http_code}') -ne 200 ]] && exit 1

echo "Make changes to other rule -> should 403"
cp rules.3.2.json rules.3.json; sleep 3
[[ $(curl --retry 7 --retry-connrefused -s -o /dev/null -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 200 ]] && exit 1
[[ $(curl --retry 7 --retry-connrefused -s -o /dev/null -f ${OATHKEEPER_PROXY}/other-rules -w '%{http_code}') -ne 403 ]] && exit 1

echo "Remove all configs -> all 404"
cp config.6.yaml config.yaml; sleep 3
[[ $(curl --retry 7 --retry-connrefused -s -o /dev/null -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 404 ]] && exit 1
[[ $(curl --retry 7 --retry-connrefused -s -o /dev/null -f ${OATHKEEPER_PROXY}/other-rules -w '%{http_code}') -ne 404 ]] && exit 1

kill %1 || true

trap - EXIT
