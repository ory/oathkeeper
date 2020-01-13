SHELL=/bin/bash -o pipefail

.PHONY: tools
tools:
		GO111MODULE=on go install github.com/ory/go-acc github.com/ory/x/tools/listx github.com/go-swagger/go-swagger/cmd/swagger github.com/sqs/goreturns github.com/ory/sdk/swagutil

# Formats the code
.PHONY: format
format:
		goreturns -w -local github.com/ory $$(listx .)

.PHONY: gen
		gen: mocks sdk

# Generates the SDKs
.PHONY: sdk
sdk:
		$$(go env GOPATH)/bin/swagger generate spec -m -o ./docs/api.swagger.json -x internal/httpclient
		$$(go env GOPATH)/bin/swagutil sanitize ./docs/api.swagger.json
		$$(go env GOPATH)/bin/swagger flatten --with-flatten=remove-unused -o ./docs/api.swagger.json ./docs/api.swagger.json
		$$(go env GOPATH)/bin/swagger validate ./docs/api.swagger.json
		rm -rf internal/httpclient
		mkdir -p internal/httpclient
		$$(go env GOPATH)/bin/swagger generate client -f ./docs/api.swagger.json -t internal/httpclient -A Ory_Oathkeeper
		make format


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
		docker build -t oryd/oathkeeper:dev .
		docker build -t oryd/oathkeeper:dev-alpine -f Dockerfile-alpine .
		rm oathkeeper
