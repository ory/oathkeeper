#!/bin/bash

set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

mockgen -package evaluator -destination evaluator/hydra_sdk_mock.go -source ./vendor/github.com/ory/hydra/sdk/go/hydra/sdk_api.go SDK
