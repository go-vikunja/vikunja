// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

module code.vikunja.io/api

require (
	dario.cat/mergo v1.0.1
	github.com/ThreeDotsLabs/watermill v1.4.6
	github.com/adlio/trello v1.12.0
	github.com/arran4/golang-ical v0.3.2
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/bbrks/go-blurhash v1.1.1
	github.com/c2h5oh/datasize v0.0.0-20231215233829-aa82cc1e6500
	github.com/coreos/go-oidc/v3 v3.14.1
	github.com/cweill/gotests v1.6.0
	github.com/d4l3k/messagediff v1.2.1
	github.com/disintegration/imaging v1.6.2
	github.com/dustinkirkland/golang-petname v0.0.0-20240422154211-76c06c4bde6b
	github.com/gabriel-vasile/mimetype v1.4.8
	github.com/ganigeorgiev/fexpr v0.5.0
	github.com/getsentry/sentry-go v0.31.1
	github.com/getsentry/sentry-go/echo v0.31.1
	github.com/go-ldap/ldap/v3 v3.4.10
	github.com/go-sql-driver/mysql v1.9.1
	github.com/go-testfixtures/testfixtures/v3 v3.14.0
	github.com/gocarina/gocsv v0.0.0-20231116093920-b87c2d0e983a
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-version v1.7.0
	github.com/hhsnopek/etag v0.0.0-20171206181245-aea95f647346
	github.com/huandu/go-clone/generic v1.7.2
	github.com/iancoleman/strcase v0.3.0
	github.com/jinzhu/copier v0.4.0
	github.com/jszwedko/go-datemath v0.1.1-0.20230526204004-640a500621d6
	github.com/labstack/echo-jwt/v4 v4.3.1
	github.com/labstack/echo/v4 v4.13.3
	github.com/labstack/gommon v0.4.2
	github.com/lib/pq v1.10.9
	github.com/magefile/mage v1.15.0
	github.com/mattn/go-sqlite3 v1.14.27
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/olekukonko/tablewriter v0.0.5
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pquerna/otp v1.4.0
	github.com/prometheus/client_golang v1.21.1
	github.com/redis/go-redis/v9 v9.7.3
	github.com/robfig/cron/v3 v3.0.1
	github.com/samedi/caldav-go v3.0.0+incompatible
	github.com/spf13/afero v1.14.0
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
	github.com/swaggo/swag v1.16.4
	github.com/tkuchiki/go-timezone v0.2.3
	github.com/typesense/typesense-go/v2 v2.0.0
	github.com/typesense/typesense-go/v3 v3.2.0
	github.com/ulule/limiter/v3 v3.11.2
	github.com/wneessen/go-mail v0.6.2
	github.com/yuin/goldmark v1.7.8
	golang.org/x/crypto v0.36.0
	golang.org/x/image v0.25.0
	golang.org/x/oauth2 v0.28.0
	golang.org/x/sync v0.13.0
	golang.org/x/sys v0.32.0
	golang.org/x/term v0.30.0
	golang.org/x/text v0.23.0
	gopkg.in/d4l3k/messagediff.v1 v1.2.1
	mvdan.cc/xurls/v2 v2.6.0
	src.techknowlogick.com/xgo v1.8.1-0.20241105013731-313dedef864f
	src.techknowlogick.com/xormigrate v1.7.1
	xorm.io/builder v0.3.13
	xorm.io/xorm v1.3.9
)

require (
	cel.dev/expr v0.16.1 // indirect
	cloud.google.com/go v0.116.0 // indirect
	cloud.google.com/go/auth v0.13.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.6 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	cloud.google.com/go/iam v1.2.2 // indirect
	cloud.google.com/go/longrunning v0.6.2 // indirect
	cloud.google.com/go/monitoring v1.21.2 // indirect
	cloud.google.com/go/spanner v1.73.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/GoogleCloudPlatform/grpc-gcp-go/grpcgcp v1.5.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.25.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/beevik/etree v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20240905190251-b4127c9b8d78 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/envoyproxy/go-control-plane v0.13.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.7 // indirect
	github.com/go-chi/chi/v5 v5.1.0 // indirect
	github.com/go-jose/go-jose/v4 v4.0.5 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.3 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/googleapis/go-sql-spanner v1.7.4 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/go-clone v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/laurent22/ical-go v0.1.1-0.20181107184520-7e5d6ade8eef // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oapi-codegen/runtime v1.1.1 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/onsi/gomega v1.16.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sony/gobreaker v1.0.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/tj/assert v0.0.3 // indirect
	github.com/urfave/cli/v2 v2.3.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/yosssi/gohtml v0.0.0-20201013000340-ee4748c638f4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.29.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.54.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.54.0 // indirect
	go.opentelemetry.io/otel v1.29.0 // indirect
	go.opentelemetry.io/otel/metric v1.29.0 // indirect
	go.opentelemetry.io/otel/sdk v1.29.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.29.0 // indirect
	go.opentelemetry.io/otel/trace v1.29.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	golang.org/x/tools v0.28.0 // indirect
	google.golang.org/api v0.215.0 // indirect
	google.golang.org/genproto v0.0.0-20241118233622-e639e219e697 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241209162323-e6fa225c2576 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241223144023-3abc09e42ca8 // indirect
	google.golang.org/grpc v1.67.3 // indirect
	google.golang.org/protobuf v1.36.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace github.com/samedi/caldav-go => github.com/kolaente/caldav-go v3.0.1-0.20190610114120-2a4eb8b5dcc9+incompatible // Branch: feature/dynamic-supported-components, PR: https://github.com/samedi/caldav-go/pull/6 and https://github.com/samedi/caldav-go/pull/7

replace github.com/labstack/echo/v4 => github.com/kolaente/echo/v4 v4.0.0-20250124112709-682dfde74c31 // https://github.com/labstack/echo/pull/2738

go 1.23.0

toolchain go1.24.2
