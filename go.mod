// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

module code.vikunja.io/api

require (
	4d63.com/tz v1.2.0
	code.vikunja.io/web v0.0.0-20200809154828-8767618f181f
	gitea.com/xorm/xorm-redis-cache v0.2.0
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/beevik/etree v1.1.0 // indirect
	github.com/c2h5oh/datasize v0.0.0-20200825124411-48ed595a09d2
	github.com/client9/misspell v0.3.4
	github.com/cweill/gotests v1.5.3
	github.com/d4l3k/messagediff v1.2.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/disintegration/imaging v1.6.2
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/fzipp/gocyclo v0.0.0-20150627053110-6acd4345c835
	github.com/gabriel-vasile/mimetype v1.1.1
	github.com/getsentry/sentry-go v0.7.0
	github.com/go-redis/redis/v7 v7.4.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-testfixtures/testfixtures/v3 v3.4.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/gordonklaus/ineffassign v0.0.0-20200809085317-e36bfde3bb78
	github.com/iancoleman/strcase v0.1.2
	github.com/imdario/mergo v0.3.11
	github.com/jgautheron/goconst v0.0.0-20200920201509-8f5268ce89d5
	github.com/kr/text v0.2.0 // indirect
	github.com/labstack/echo/v4 v4.1.17
	github.com/labstack/gommon v0.3.0
	github.com/laurent22/ical-go v0.1.1-0.20181107184520-7e5d6ade8eef
	github.com/lib/pq v1.8.0
	github.com/magefile/mage v1.10.0
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.4
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/onsi/gomega v1.9.0 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pquerna/otp v1.2.0
	github.com/prometheus/client_golang v1.7.1
	github.com/samedi/caldav-go v3.0.0+incompatible
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/vfsgen v0.0.0-20200824052919-0d455de96546
	github.com/spf13/afero v1.4.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/swaggo/swag v1.6.7
	github.com/ulule/limiter/v3 v3.5.0
	golang.org/x/crypto v0.0.0-20201002094018-c90954cbb977
	golang.org/x/image v0.0.0-20200927104501-e162460cd6b5
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/sync v0.0.0-20200930132711-30421366ff76
	golang.org/x/tools v0.0.0-20200410194907-79a7a3126eef // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/d4l3k/messagediff.v1 v1.2.1
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	honnef.co/go/tools v0.0.1-2020.1.5
	src.techknowlogick.com/xgo v1.1.1-0.20200811225412-bff6512e7c9c
	src.techknowlogick.com/xormigrate v1.3.0
	xorm.io/builder v0.3.7
	xorm.io/core v0.7.3
	xorm.io/xorm v1.0.2
)

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	github.com/hpcloud/tail => github.com/jeffbean/tail v1.0.1 // See https://github.com/hpcloud/tail/pull/159
	github.com/samedi/caldav-go => github.com/kolaente/caldav-go v3.0.1-0.20190524174923-9e5cd1688227+incompatible // Branch: feature/dynamic-supported-components, PR: https://github.com/samedi/caldav-go/pull/6 and https://github.com/samedi/caldav-go/pull/7
	gopkg.in/fsnotify.v1 => github.com/kolaente/fsnotify v1.4.10-0.20200411160148-1bc3c8ff4048 // See https://github.com/fsnotify/fsnotify/issues/328 and https://github.com/golang/go/issues/26904
)

go 1.13
