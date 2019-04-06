module github.com/ory/oathkeeper

require (
	cloud.google.com/go v0.37.2 // indirect
	github.com/Microsoft/go-winio v0.4.12 // indirect
	github.com/Songmu/retry v0.1.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20180720115003-f9ffefc3facf
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-errors/errors v1.0.1
	github.com/go-openapi/analysis v0.19.0 // indirect
	github.com/go-openapi/errors v0.19.0
	github.com/go-openapi/inflect v0.19.0 // indirect
	github.com/go-openapi/loads v0.19.0 // indirect
	github.com/go-openapi/runtime v0.19.0
	github.com/go-openapi/spec v0.19.0 // indirect
	github.com/go-openapi/strfmt v0.19.0
	github.com/go-openapi/swag v0.19.0
	github.com/go-openapi/validate v0.19.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/go-swagger/go-swagger v0.19.0
	github.com/go-swagger/scan-repo-boundary v0.0.0-20180623220736-973b3573c013 // indirect
	github.com/gobuffalo/packr v1.24.1 // indirect
	github.com/golang/gddo v0.0.0-20190312205958-5a2505f3dbf0 // indirect
	github.com/golang/mock v1.2.0
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/gorilla/handlers v1.4.0 // indirect
	github.com/gorilla/mux v1.7.1 // indirect
	github.com/gorilla/sessions v1.1.3 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/jmoiron/sqlx v1.2.0
	github.com/julienschmidt/httprouter v1.2.0
	github.com/lib/pq v1.0.0
	github.com/luna-duclos/instrumentedsql v0.0.0-20190316074304-ecad98b20aec // indirect
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/mattn/goveralls v0.0.2
	github.com/meatballhat/negroni-logrus v0.0.0-20170801195057-31067281800f
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/gox v1.0.0
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/ory/dockertest v3.3.4+incompatible
	github.com/ory/fosite v0.29.1
	github.com/ory/go-acc v0.0.0-20181118080137-ddc355013f90
	github.com/ory/go-convenience v0.1.0
	github.com/ory/graceful v0.1.1
	github.com/ory/herodot v0.6.0
	github.com/ory/hydra v0.0.0-20181208123928-e4bc6c269c6f
	github.com/ory/keto v0.0.0-20181213093025-a8d7f9f546ae
	github.com/ory/ladon v1.0.1
	github.com/ory/x v0.0.40
	github.com/pborman/uuid v1.2.0
	github.com/pelletier/go-toml v1.3.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/rubenv/sql-migrate v0.0.0-20190327083759-54bad0a9b051
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.3.0
	github.com/tcnksm/ghr v0.12.0
	github.com/tcnksm/go-gitconfig v0.1.2 // indirect
	github.com/tcnksm/go-latest v0.0.0-20170313132115-e3007ae9052e // indirect
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce
	github.com/urfave/negroni v1.0.0
	go.opencensus.io v0.20.0 // indirect
	golang.org/x/crypto v0.0.0-20190404164418-38d8ce5564a5 // indirect
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3 // indirect
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	golang.org/x/tools v0.0.0-20190404132500-923d25813098
	google.golang.org/appengine v1.5.0 // indirect
	google.golang.org/genproto v0.0.0-20190404172233-64821d5d2107 // indirect
	google.golang.org/grpc v1.19.1 // indirect
	gopkg.in/resty.v1 v1.10.3 // indirect
	gopkg.in/square/go-jose.v2 v2.3.0
)

// Fix for https://github.com/golang/lint/issues/436
replace github.com/golang/lint => github.com/golang/lint v0.0.0-20190227174305-8f45f776aaf1
