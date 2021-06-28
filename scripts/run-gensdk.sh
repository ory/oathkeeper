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


set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

scripts/run-genswag.sh

rm -rf ./sdk/go/oathkeeper/swagger
rm -rf ./sdk/js/swagger

# curl -O scripts/swagger-codegen-cli-2.2.3.jar http://central.maven.org/maven2/io/swagger/swagger-codegen-cli/2.2.3/swagger-codegen-cli-2.2.3.jar

java -jar scripts/swagger-codegen-cli-2.2.3.jar generate -i ./spec/api.json -l go -o ./sdk/go/oathkeeper/swagger
java -jar scripts/swagger-codegen-cli-2.2.3.jar generate -i ./spec/api.json -l javascript -o ./sdk/js/swagger

scripts/run-format.sh

git checkout HEAD -- sdk/go/oathkeeper/swagger/rule_handler.go

git add -A .

rm -f ./sdk/js/swagger/package.json
rm -rf ./sdk/js/swagger/test

npm run prettier
