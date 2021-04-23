#!/usr/bin/env bash

set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

mockgen -package proxy -destination proxy/keto_sdk_mock.go -source ./vendor/github.com/ory/keto/sdk/go/keto/sdk_warden.go WardenSDK
mockgen -package proxy -destination proxy/authenticator_oauth2_introspection_mock.go -source ./proxy/authenticator_oauth2_introspection.go authenticatorOAuth2IntrospectionHelper
