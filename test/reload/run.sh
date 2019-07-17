#!/bin/bash

set -euxo pipefail

waitport() {
    while ! nc -z localhost $1 ; do sleep 1 ; done
}

cd "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

killall oathkeeper || true

export OATHKEEPER_PROXY=http://127.0.0.1:6060
export OATHKEEPER_API=http://127.0.0.1:6061
export GO111MODULE=on

(cd ../../; make install)

cp config.1.yaml config.yaml

LOG_LEVEL=debug oathkeeper --config ./config.yaml serve >> ./oathkeeper.log 2>&1 &

waitport 6060
waitport 6061

sleep 2

echo "Executing request against no configured rules -> 404"
[[ $(curl --silent --output /dev/null -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 404 ]] && exit 1

echo "Executing request against a now configured rule -> 200"
cp config.2.yaml config.yaml; sleep 1; [[ $(curl --silent --output /dev/null -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 200 ]] && exit 1

echo "Executing request against updated rule with deny in it -> 403"
cp config.3.yaml config.yaml; sleep 1; [[ $(curl --silent --output /dev/null -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 403 ]] && exit 1

echo "Executing request against updated rule with deny deny disabled -> 500"
cp config.4.yaml config.yaml; sleep 1; [[ $(curl --silent --output /dev/null -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 500 ]] && exit 1

echo "Making sure route to be registered is not available yet -> 404"
[[ $(curl --silent --output /dev/null -f ${OATHKEEPER_PROXY}/other-rules -w '%{http_code}') -ne 404 ]] && exit 1

echo "Load another rule set -> all 200"
cp rules.3.1.json rules.3.json
cp config.5.yaml config.yaml; sleep 1
[[ $(curl --silent --output /dev/null -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 200 ]] && exit 1
[[ $(curl --silent --output /dev/null -f ${OATHKEEPER_PROXY}/other-rules -w '%{http_code}') -ne 200 ]] && exit 1

echo "Make changes to other rule -> should 403"
cp rules.3.2.json rules.3.json; sleep 1
[[ $(curl --silent --output /dev/null -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 200 ]] && exit 1
[[ $(curl --silent --output /dev/null -f ${OATHKEEPER_PROXY}/other-rules -w '%{http_code}') -ne 403 ]] && exit 1

echo "Remove all configs -> all 404"
cp config.6.yaml config.yaml; sleep 1
[[ $(curl --silent --output /dev/null -f ${OATHKEEPER_PROXY}/rules -w '%{http_code}') -ne 404 ]] && exit 1
[[ $(curl --silent --output /dev/null -f ${OATHKEEPER_PROXY}/other-rules -w '%{http_code}') -ne 404 ]] && exit 1

kill %1 || true
exit 0
