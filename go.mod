module github.com/ory/oathkeeper

go 1.16

replace gopkg.in/DataDog/dd-trace-go.v1 => gopkg.in/DataDog/dd-trace-go.v1 v1.27.1

replace github.com/dgrijalva/jwt-go => github.com/form3tech-oss/jwt-go v1.0.3-0.20210625141045-a211650c6ae1

replace github.com/oleiade/reflections => github.com/oleiade/reflections v1.0.1

require (
	github.com/Azure/azure-pipeline-go v0.2.2
	github.com/Azure/azure-storage-blob-go v0.9.0
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/auth0/go-jwt-middleware v1.0.1
	github.com/aws/aws-sdk-go v1.34.28
	github.com/blang/semver v3.5.1+incompatible
	github.com/bxcodec/faker v2.0.1+incompatible
	github.com/dgraph-io/ristretto v0.1.0
	github.com/dlclark/regexp2 v1.2.0
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/fsnotify/fsnotify v1.5.1
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/errors v0.20.1
	github.com/go-openapi/jsonreference v0.19.4 // indirect
	github.com/go-openapi/runtime v0.20.0
	github.com/go-openapi/strfmt v0.20.3
	github.com/go-openapi/swag v0.19.15
	github.com/go-openapi/validate v0.19.10
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gobuffalo/httptest v1.0.2
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/gobwas/glob v0.2.3
	github.com/golang-jwt/jwt/v4 v4.0.0
	github.com/golang/gddo v0.0.0-20190904175337-72a348e765d2
	github.com/golang/mock v1.6.0
	github.com/google/go-replayers/httpreplay v0.1.0
	github.com/google/uuid v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/imdario/mergo v0.3.11
	github.com/julienschmidt/httprouter v1.2.0
	github.com/lib/pq v1.3.0
	github.com/mattn/goveralls v0.0.6
	github.com/mitchellh/copystructure v1.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/ory/analytics-go/v4 v4.0.2
	github.com/ory/fosite v0.36.1
	github.com/ory/go-convenience v0.1.0
	github.com/ory/gojsonschema v1.2.0
	github.com/ory/graceful v0.1.1
	github.com/ory/herodot v0.9.12
	github.com/ory/jsonschema/v3 v3.0.5
	github.com/ory/ladon v1.1.0
	github.com/ory/viper v1.7.5
	github.com/ory/x v0.0.344
	github.com/pborman/uuid v1.2.1
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/cors v1.8.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/square/go-jose v2.3.1+incompatible
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.9.4
	github.com/tidwall/sjson v1.2.2
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce
	github.com/urfave/negroni v1.0.0
	gocloud.dev v0.20.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
	google.golang.org/api v0.44.0
	gopkg.in/square/go-jose.v2 v2.6.0
)
