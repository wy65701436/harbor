module github.com/goharbor/harbor/src

go 1.12

replace github.com/goharbor/harbor => ../

require (
	cloud.google.com/go v0.38.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Masterminds/semver v1.4.2
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5 // indirect
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d // indirect
	github.com/Unknwon/goconfig v0.0.0-20160216183935-5f601ca6ef4d // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/aliyun/alibaba-cloud-sdk-go v0.0.0-20190726115642-cd293c93fd97
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496 // indirect
	github.com/astaxie/beego v1.12.1
	github.com/aws/aws-sdk-go v1.19.47
	github.com/beego/i18n v0.0.0-20140604031826-e87155e8f0c0
	github.com/bitly/go-hostpool v0.0.0-20171023180738-a3a6125de932 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/bmatcuk/doublestar v1.1.1
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/bugsnag/bugsnag-go v1.5.2 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/casbin/casbin v1.7.0
	github.com/cenkalti/backoff v2.1.1+incompatible // indirect
	github.com/cloudflare/cfssl v0.0.0-20190510060611-9c027c93ba9e // indirect
	github.com/containerd/containerd v1.3.0 // indirect
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/dghubble/sling v1.1.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/go v0.0.0-20160303222718-d30aec9fd63c // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7
	github.com/garyburd/redigo v1.6.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/errors v0.19.2
	github.com/go-openapi/runtime v0.19.5
	github.com/go-openapi/spec v0.19.4 // indirect
	github.com/go-openapi/strfmt v0.19.3
	github.com/go-openapi/validate v0.19.5 // indirect
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gocraft/work v0.5.1
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang-migrate/migrate v3.3.0+incompatible
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/certificate-transparency-go v1.0.21 // indirect
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/google/go-querystring v0.0.0-20170111101155-53e6ce116135 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.7.2
	github.com/graph-gophers/dataloader v5.0.0+incompatible
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/jinzhu/gorm v1.9.8 // indirect
	github.com/justinas/alice v0.0.0-20171023064455-03f45bd4b7da
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lib/pq v1.1.0
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/miekg/pkcs11 v0.0.0-20170220202408-7283ca79f35e // indirect
	github.com/morikuni/aec v0.0.0-20170113033406-39771216ff4c // indirect
	github.com/olekukonko/tablewriter v0.0.1
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/client_golang v1.2.1 // indirect
	github.com/robfig/cron v1.0.0
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/viper v1.4.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/theupdateframework/notary v0.6.1
	golang.org/x/crypto v0.0.0-20200128174031-69ecbb4d6d5d
	golang.org/x/net v0.0.0-20191028085509-fe3aa8a45271 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20191028164358-195ce5e7f934 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20191028173616-919d9bdd9fe6 // indirect
	google.golang.org/grpc v1.24.0 // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20150924051756-4e86f4367175 // indirect
	gopkg.in/dancannon/gorethink.v3 v3.0.5 // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5 // indirect
	gopkg.in/ldap.v2 v2.5.0
	gopkg.in/square/go-jose.v2 v2.3.0 // indirect
	gopkg.in/yaml.v2 v2.2.4
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v8.0.0+incompatible
	k8s.io/helm v2.16.1+incompatible
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
