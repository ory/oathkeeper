#!/usr/bin/env bash

set -euxo pipefail

schema_version="${1:-$(git rev-parse --short HEAD)}"

sed "s!ory://tracing-config!https://raw.githubusercontent.com/ory/oathkeeper/$schema_version/oryx/otelx/config.schema.json!g;
s!ory://logging-config!https://raw.githubusercontent.com/ory/oathkeeper/$schema_version/oryx/logrusx/config.schema.json!g;
s!/.schema/config.schema.json!https://github.com/ory/oathkeeper/schema/config.schema.json!g" spec/config.schema.json > .schema/config.schema.json

git commit --author="ory-bot <60093411+ory-bot@users.noreply.github.com>" -m "autogen: render config schema" .schema/config.schema.json || true
