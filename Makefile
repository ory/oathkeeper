SHELL=/bin/bash -o pipefail

export GO111MODULE := on
export PATH := .bin:${PATH}
export PWD := $(shell pwd)

GO_DEPENDENCIES = github.com/ory/go-acc \
				  github.com/go-swagger/go-swagger/cmd/swagger \
				  github.com/go-bindata/go-bindata/go-bindata

define make-go-dependency
  # go install is responsible for not re-building when the code hasn't changed
  .bin/$(notdir $1): go.sum go.mod
		GOBIN=$(PWD)/.bin/ go install $1
endef
$(foreach dep, $(GO_DEPENDENCIES), $(eval $(call make-go-dependency, $(dep))))

node_modules: package-lock.json
	npm ci
	touch node_modules

.bin/clidoc: go.mod
	go build -o .bin/clidoc ./cmd/clidoc/.

.bin/goimports: Makefile
	GOBIN=$(shell pwd)/.bin go install golang.org/x/tools/cmd/goimports@latest

.bin/licenses: Makefile
	curl https://raw.githubusercontent.com/ory/ci/master/licenses/install | sh

.bin/ory: Makefile
	curl https://raw.githubusercontent.com/ory/meta/master/install.sh | bash -s -- -b .bin ory v0.1.48
	touch .bin/ory

format: .bin/goimports .bin/ory node_modules
	.bin/ory dev headers copyright --type=open-source --exclude=internal/httpclient
	goimports -w --local github.com/ory .
	gofmt -l -s -w .
	curl https://raw.githubusercontent.com/ory/ci/master/authors/authors.sh | env PRODUCT="Ory Oathkeeper" bash
	npm exec -- prettier --write .

licenses: .bin/licenses node_modules  # checks open-source licenses
	.bin/licenses

# Generates the SDK
.PHONY: sdk
sdk: .bin/swagger .bin/ory node_modules
	rm -rf internal/httpclient
	mkdir -p internal/httpclient

	swagger generate spec -m -o spec/swagger.json \
		-c github.com/ory/oathkeeper \
		-c github.com/ory/x/healthx
	ory dev swagger sanitize ./spec/swagger.json
	swagger validate ./spec/swagger.json
	CIRCLE_PROJECT_USERNAME=ory CIRCLE_PROJECT_REPONAME=oathkeeper \
		ory dev openapi migrate \
			--health-path-tags metadata \
			-p https://raw.githubusercontent.com/ory/x/master/healthx/openapi/patch.yaml \
			-p file://.schema/openapi/patches/meta.yaml \
			spec/swagger.json spec/api.json

	swagger generate client -f ./spec/swagger.json -t internal/httpclient -A Ory_Oathkeeper

	make --no-print-dir format

.PHONY: install-stable
install-stable:
	OATHKEEPER_LATEST=$$(git describe --abbrev=0 --tags)
	git checkout $$OATHKEEPER_LATEST
	GO111MODULE=on go install \
		-ldflags "-X github.com/ory/oathkeeper/x.Version=$$OATHKEEPER_LATEST -X github.com/ory/oathkeeper/x.Date=`TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ'` -X github.com/ory/oathkeeper/x.Commit=`git rev-parse HEAD`" \
		.
	git checkout master

.PHONY: install
install:
	GO111MODULE=on go install .

.PHONY: docker
docker:
	CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build
	docker build -t oryd/oathkeeper:dev .
	docker build -t oryd/oathkeeper:dev-alpine -f Dockerfile-alpine .
	rm oathkeeper

docs/cli: .bin/clidoc
	clidoc .

.PHONY: post-release
post-release:
	echo "nothing to do"
