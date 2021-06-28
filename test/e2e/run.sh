#!/usr/bin/env bash
# Copyright 2021 Ory GmbH
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


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

(cd ../../; make install)
go install github.com/ory/oathkeeper/test/e2e/okapi
go install github.com/ory/oathkeeper/test/e2e/okclient

oathkeeper --config ./config.yml serve >> ./oathkeeper.e2e.log 2>&1 &
PORT=6662 okapi >> ./api.e2e.log 2>&1 &

waitport 6660
waitport 6661
waitport 6662

function finish {
  cat ./oathkeeper.e2e.log
  echo "-----"
  cat ./api.e2e.log
}
trap finish EXIT

okclient

kill %1 || true
kill %2 || true

trap - EXIT
