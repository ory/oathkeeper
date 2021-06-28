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

mockgen -package proxy -destination proxy/keto_sdk_mock.go -source ./vendor/github.com/ory/keto/sdk/go/keto/sdk_warden.go WardenSDK
mockgen -package proxy -destination proxy/authenticator_oauth2_introspection_mock.go -source ./proxy/authenticator_oauth2_introspection.go authenticatorOAuth2IntrospectionHelper
