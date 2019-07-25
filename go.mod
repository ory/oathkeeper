module github.com/ory/oathkeeper

replace github.com/ory/hive => ../hive

require (
	github.com/Microsoft/go-winio v0.4.12 // indirect
	github.com/asaskevich/govalidator v0.0.0-20180720115003-f9ffefc3facf
	github.com/auth0/go-jwt-middleware v0.0.0-20170425171159-5493cabe49f7
	github.com/bxcodec/faker v2.0.1+incompatible
	github.com/codegangsta/negroni v1.0.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fsnotify/fsnotify v1.4.7
	github.com/ghodss/yaml v1.0.0
	github.com/go-errors/errors v1.0.1
	github.com/go-openapi/analysis v0.19.0 // indirect
	github.com/go-openapi/errors v0.19.0
	github.com/go-openapi/loads v0.19.0 // indirect
	github.com/go-openapi/runtime v0.19.0
	github.com/go-openapi/spec v0.19.0 // indirect
	github.com/go-openapi/strfmt v0.19.0
	github.com/go-openapi/swag v0.19.0
	github.com/go-openapi/validate v0.19.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/go-swagger/go-swagger v0.19.0
	github.com/gobuffalo/packr/v2 v2.0.0-rc.15
	github.com/golang/gddo v0.0.0-20190312205958-5a2505f3dbf0 // indirect
	github.com/golang/mock v1.3.1
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.1 // indirect
	github.com/hashicorp/golang-lru v0.5.1
	github.com/julienschmidt/httprouter v1.2.0
	github.com/lib/pq v1.0.0
	github.com/luna-duclos/instrumentedsql v0.0.0-20190316074304-ecad98b20aec // indirect
	github.com/mattn/goveralls v0.0.2
	github.com/meatballhat/negroni-logrus v0.0.0-20170801195057-31067281800f
	github.com/opencontainers/runc v1.0.0-rc5 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/ory/fosite v0.29.2
	github.com/ory/go-acc v0.0.0-20181118080137-ddc355013f90
	github.com/ory/go-convenience v0.1.0
	github.com/ory/gojsonschema v0.0.0-20190720140244-a64d4f892691
	github.com/ory/graceful v0.1.1
	github.com/ory/herodot v0.6.2
	github.com/ory/hive v0.0.0-00010101000000-000000000000
	github.com/ory/ladon v1.0.1
	github.com/ory/viper v1.5.6
	github.com/ory/x v0.0.66
	github.com/pborman/uuid v1.2.0
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/errors v0.8.1
	github.com/rs/cors v1.6.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/sqs/goreturns v0.0.0-20181028201513-538ac6014518
	github.com/square/go-jose v2.3.1+incompatible
	github.com/stretchr/testify v1.3.0
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce
	github.com/urfave/negroni v1.0.0
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/tools v0.0.0-20190711191110-9a621aea19f8
	gopkg.in/square/go-jose.v2 v2.3.0
)

// Fix for https://github.com/golang/lint/issues/436
replace github.com/golang/lint => github.com/golang/lint v0.0.0-20190227174305-8f45f776aaf1
