#!/bin/sh

set -euxo pipefail

ory_x_version="$(go list -f '{{.Version}}' -m github.com/ory/x)"

sed "s!ory://tracing-config!https://raw.githubusercontent.com/ory/x/$ory_x_version/otelx/config.schema.json!g;
s!ory://logging-config!https://raw.githubusercontent.com/ory/x/$ory_x_version/logrusx/config.schema.json!g;
s!/.schema/config.schema.json!https://github.com/ory/oathkeeper/schema/config.schema.json!g" spec/config.schema.json > .schema/config.schema.json

git commit --author="ory-bot <60093411+ory-bot@users.noreply.github.com>" -m "autogen: render config schema" .schema/config.schema.json || true
