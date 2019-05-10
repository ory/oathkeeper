SHELL=/bin/bash -o pipefail

# Formats the code
.PHONY: format
format:
		goreturns -w -local github.com/ory $$(listx .)

.PHONY: mocks
mocks:

.PHONY: gen
		gen: mocks sdk

.PHONY: sdk
sdk:
		GO111MODULE=on go mod tidy
		GO111MODULE=on go mod vendor
		GO111MODULE=off swagger generate spec -m -o ./docs/api.swagger.json
		GO111MODULE=off swagger validate ./docs/api.swagger.json

		rm -rf ./sdk/go/oathkeeper/*
		rm -rf ./sdk/js/swagger

		GO111MODULE=off swagger generate client -f ./docs/api.swagger.json -t sdk/go/oathkeeper -A Ory_Oathkeeper

		java -jar scripts/swagger-codegen-cli-2.2.3.jar generate -i ./docs/api.swagger.json -l javascript -o ./sdk/js/swagger

		cd sdk/go; goreturns -w -i -local github.com/ory $$(listx .)

		rm -f ./sdk/js/swagger/package.json
		rm -rf ./sdk/js/swagger/test
		rm -rf ./vendor

.PHONY: install-stable
install-stable:
		OATHKEEPER_LATEST=$$(git describe --abbrev=0 --tags)
		git checkout $$OATHKEEPER_LATEST
		GO111MODULE=on go install \
				-ldflags "-X github.com/ory/oathkeeper/cmd.Version=$$OATHKEEPER_LATEST -X github.com/ory/oathkeeper/cmd.Date=`TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ'` -X github.com/ory/oathkeeper/cmd.Commit=`git rev-parse HEAD`" \
				.
		git checkout master

.PHONY: install
install:
		GO111MODULE=on go install .
