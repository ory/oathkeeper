SHELL=/bin/bash -o pipefail

# Formats the code
.PHONY: format
format:
		goreturns -w -local github.com/ory $$(listx .)
		npm run format:docs

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
		packr2
		GO111MODULE=on go install \
				-ldflags "-X github.com/ory/oathkeeper/x.Version=$$OATHKEEPER_LATEST -X github.com/ory/oathkeeper/x.Date=`TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ'` -X github.com/ory/oathkeeper/x.Commit=`git rev-parse HEAD`" \
				.
		packr2 clean
		git checkout master

.PHONY: install
install:
		packr2 || (GO111MODULE=on go install github.com/gobuffalo/packr/v2/packr2 && packr2)
		GO111MODULE=on go install .
		packr2 clean

.PHONY: docker
docker:
		packr2 || (GO111MODULE=on go install github.com/gobuffalo/packr/v2/packr2 && packr2)
		CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build
		packr2 clean
		docker build -t oryd/oathkeeper:unstable.$$(date +'%s') .
		rm oathkeeper
