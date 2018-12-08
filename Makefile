SHELL=/bin/bash -o pipefail

.PHONY: format
format:
		goreturns -w -local github.com/ory $$(listx .)

.PHONY: gen-mocks
gen-mocks:
		mockgen -package proxy -destination proxy/keto_warden_sdk_mock.go -source ./proxy/authorizer_keto_warden.go KetoWardenSDK
		mockgen -package proxy -destination proxy/authenticator_oauth2_introspection_mock.go -source ./proxy/authenticator_oauth2_introspection.go authenticatorOAuth2IntrospectionHelper

.PHONY: gen
		gen: gen-mocks gen-sdk

.PHONY: gen-sdk
gen-sdk:
		swagger generate spec -m -o ./docs/api.swagger.json
		swagger validate ./docs/api.swagger.json

		rm -rf ./sdk/go/oathkeeper/swagger
		rm -rf ./sdk/js/swagger

		java -jar scripts/swagger-codegen-cli-2.2.3.jar generate -i ./docs/api.swagger.json -l go -o ./sdk/go/oathkeeper/swagger
		java -jar scripts/swagger-codegen-cli-2.2.3.jar generate -i ./docs/api.swagger.json -l javascript -o ./sdk/js/swagger

		cd sdk/go; goreturns -w -i -local github.com/ory $$(listx .)

		git checkout HEAD -- sdk/go/oathkeeper/swagger/rule_handler.go

		rm -f ./sdk/js/swagger/package.json
		rm -rf ./sdk/js/swagger/test
