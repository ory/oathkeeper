SHELL=/bin/bash -o pipefail

export GO111MODULE	:= on
export PATH					:= .bin:${PATH}
export PWD					:= $(shell pwd)
export IMAGE_TAG		:= $(if $(IMAGE_TAG),$(IMAGE_TAG),dev)

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
	curl --retry 7 --retry-connrefused https://raw.githubusercontent.com/ory/ci/master/licenses/install | sh

.bin/ory: Makefile
	curl --retry 7 --retry-connrefused https://raw.githubusercontent.com/ory/meta/master/install.sh | bash -s -- -b .bin ory v0.2.2
	touch .bin/ory

.bin/golangci-lint: Makefile
	curl --retry 7 --retry-connrefused -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -d -b .bin v2.4.0

authors:  # updates the AUTHORS file
	curl --retry 7 --retry-connrefused https://raw.githubusercontent.com/ory/ci/master/authors/authors.sh | env PRODUCT="Ory Oathkeeper" bash

.PHONY: format
format: .bin/goimports .bin/ory node_modules
	.bin/ory dev headers copyright --type=open-source --exclude=internal/httpclient --exclude=oryx
	goimports -w --local github.com/ory .
	gofmt -l -s -w .
	npm exec -- prettier --write .

.PHONY: lint
lint: .bin/golangci-lint
	.bin/golangci-lint run --timeout 10m0s

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
	DOCKER_BUILDKIT=1 DOCKER_CONTENT_TRUST=1 docker build -t oryd/oathkeeper:${IMAGE_TAG} --progress=plain -f .docker/Dockerfile-build . 

.PHONY: docker-k3d
docker-k3d:
	CGO_ENABLED=0 GOOS=linux go build
	DOCKER_BUILDKIT=1 DOCKER_CONTENT_TRUST=1 docker build -t k3d-localhost:5111/oryd/oathkeeper:dev --push -f .docker/Dockerfile-distroless-static . 
	rm oathkeeper

docs/cli: .bin/clidoc
	clidoc .

.PHONY: pre-release
pre-release:
	echo "nothing to do"

.PHONY: post-release
post-release:
	echo "nothing to do"
