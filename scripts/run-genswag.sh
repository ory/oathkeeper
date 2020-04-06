#!/bin/bash

set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

swagger generate spec -m -o ./.schema/api.swagger.json
