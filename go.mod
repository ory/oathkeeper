module github.com/ory/oathkeeper

replace (
	github.com/dgrijalva/jwt-go => github.com/form3tech-oss/jwt-go v1.0.3-0.20210625141045-a211650c6ae1
	github.com/mattn/go-sqlite3 => github.com/mattn/go-sqlite3 v1.14.10
	github.com/oleiade/reflections => github.com/oleiade/reflections v1.0.1
	github.com/ory/oldx => github.com/ory/x v0.0.164
	gopkg.in/DataDog/dd-trace-go.v1 => gopkg.in/DataDog/dd-trace-go.v1 v1.27.1
)

require (
	github.com/Azure/azure-pipeline-go v0.2.2
	github.com/Azure/azure-storage-blob-go v0.9.0
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/auth0/go-jwt-middleware v1.0.1
	github.com/aws/aws-sdk-go v1.34.28
	github.com/blang/semver v3.5.1+incompatible
	github.com/bxcodec/faker v2.0.1+incompatible
	github.com/dgraph-io/ristretto v0.1.0
	github.com/dlclark/regexp2 v1.2.0
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/fsnotify/fsnotify v1.5.4
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/errors v0.20.1
	github.com/go-openapi/runtime v0.20.0
	github.com/go-openapi/strfmt v0.19.5
	github.com/go-openapi/swag v0.19.9
	github.com/go-openapi/validate v0.19.10
	github.com/go-sql-driver/mysql v1.6.0
	github.com/go-swagger/go-swagger v0.25.0
	github.com/gobuffalo/httptest v1.0.2
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/gobwas/glob v0.2.3
	github.com/golang-jwt/jwt/v4 v4.0.0
	github.com/golang/gddo v0.0.0-20190904175337-72a348e765d2
	github.com/golang/mock v1.6.0
	github.com/google/go-replayers/httpreplay v1.1.1
	github.com/google/martian/v3 v3.3.3-0.20220816151257-0f7e6797a04d // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2
	github.com/julienschmidt/httprouter v1.3.0
	github.com/knadh/koanf v1.4.0
	github.com/lib/pq v1.10.6
	github.com/mattn/goveralls v0.0.6
	github.com/mitchellh/copystructure v1.2.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/ory/analytics-go/v4 v4.0.3
	github.com/ory/cli v0.0.10
	github.com/ory/fosite v0.36.1
	github.com/ory/go-acc v0.2.6
	github.com/ory/go-convenience v0.1.0
	github.com/ory/gojsonschema v1.2.0
	github.com/ory/graceful v0.1.1
	github.com/ory/herodot v0.9.13
	github.com/ory/jsonschema/v3 v3.0.7
	github.com/ory/ladon v1.1.0
	github.com/ory/oldx v0.0.0-00010101000000-000000000000
	github.com/ory/viper v1.7.5
	github.com/ory/x v0.0.458
	github.com/pborman/uuid v1.2.1
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/cors v1.8.2
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/cobra v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/sqs/goreturns v0.0.0-20181028201513-538ac6014518
	github.com/square/go-jose v2.3.1+incompatible
	github.com/stretchr/testify v1.8.0
	github.com/tidwall/gjson v1.14.3
	github.com/tidwall/sjson v1.2.4
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce
	github.com/urfave/negroni v1.0.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	gocloud.dev v0.20.0
	golang.org/x/crypto v0.0.0-20220517005047-85d78b3ac167
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	golang.org/x/tools v0.1.11
	google.golang.org/api v0.47.0
	gopkg.in/square/go-jose.v2 v2.6.0
)

go 1.16
