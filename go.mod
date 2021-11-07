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
	code.vikunja.io/web v0.0.0-20210706160506-d85def955bd3
	gitea.com/xorm/xorm-redis-cache v0.2.0
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/adlio/trello v1.9.0
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/beevik/etree v1.1.0 // indirect
	github.com/c2h5oh/datasize v0.0.0-20200825124411-48ed595a09d2
	github.com/coreos/go-oidc/v3 v3.1.0
	github.com/cweill/gotests v1.6.0
	github.com/d4l3k/messagediff v1.2.1
	github.com/disintegration/imaging v1.6.2
	github.com/dustinkirkland/golang-petname v0.0.0-20191129215211-8e5a1ed0cff0
	github.com/gabriel-vasile/mimetype v1.4.0
	github.com/getsentry/sentry-go v0.11.0
	github.com/go-errors/errors v1.1.1 // indirect
	github.com/go-redis/redis/v8 v8.11.4
	github.com/go-sql-driver/mysql v1.6.0
	github.com/go-testfixtures/testfixtures/v3 v3.6.1
	github.com/golang-jwt/jwt/v4 v4.1.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/golang/snappy v0.0.4 // indirect
	github.com/iancoleman/strcase v0.2.0
	github.com/imdario/mergo v0.3.12
	github.com/labstack/echo/v4 v4.6.1
	github.com/labstack/gommon v0.3.1
	github.com/laurent22/ical-go v0.1.1-0.20181107184520-7e5d6ade8eef
	github.com/lib/pq v1.10.3
	github.com/magefile/mage v1.11.0
	github.com/mattn/go-sqlite3 v1.14.9
	github.com/olekukonko/tablewriter v0.0.5
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pquerna/otp v1.3.0
	github.com/prometheus/client_golang v1.11.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/samedi/caldav-go v3.0.0+incompatible
	github.com/spf13/afero v1.6.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/swag v1.7.4
	github.com/ulule/limiter/v3 v3.8.0
	github.com/yuin/goldmark v1.4.2
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/image v0.0.0-20211028202545-6944b10bf410
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20211107104306-e0b2ad06fe42
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/d4l3k/messagediff.v1 v1.2.1
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	src.techknowlogick.com/xgo v1.4.1-0.20210311222705-d25c33fcd864
	src.techknowlogick.com/xormigrate v1.4.0
	xorm.io/builder v0.3.9
	xorm.io/core v0.7.3
	xorm.io/xorm v1.1.2
)

replace (
	github.com/adlio/trello => github.com/kolaente/trello v1.7.1-0.20201216234312-5c4ef79b531e
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	github.com/hpcloud/tail => github.com/jeffbean/tail v1.0.1 // See https://github.com/hpcloud/tail/pull/159
	github.com/samedi/caldav-go => github.com/kolaente/caldav-go v3.0.1-0.20190524174923-9e5cd1688227+incompatible // Branch: feature/dynamic-supported-components, PR: https://github.com/samedi/caldav-go/pull/6 and https://github.com/samedi/caldav-go/pull/7
	gopkg.in/fsnotify.v1 => github.com/kolaente/fsnotify v1.4.10-0.20200411160148-1bc3c8ff4048 // See https://github.com/fsnotify/fsnotify/issues/328 and https://github.com/golang/go/issues/26904
)

go 1.15
