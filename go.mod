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
	code.vikunja.io/web v0.0.0-20210131201003-26386be9a9ae
	gitea.com/xorm/xorm-redis-cache v0.2.0
	github.com/ThreeDotsLabs/watermill v1.1.1
	github.com/adlio/trello v1.8.0
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/beevik/etree v1.1.0 // indirect
	github.com/c2h5oh/datasize v0.0.0-20200825124411-48ed595a09d2
	github.com/client9/misspell v0.3.4
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/cweill/gotests v1.6.0
	github.com/d4l3k/messagediff v1.2.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/disintegration/imaging v1.6.2
	github.com/dustinkirkland/golang-petname v0.0.0-20191129215211-8e5a1ed0cff0
	github.com/fzipp/gocyclo v0.3.1
	github.com/gabriel-vasile/mimetype v1.2.0
	github.com/getsentry/sentry-go v0.10.0
	github.com/go-errors/errors v1.1.1 // indirect
	github.com/go-redis/redis/v8 v8.6.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-testfixtures/testfixtures/v3 v3.5.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/golang/snappy v0.0.2 // indirect
	github.com/gordonklaus/ineffassign v0.0.0-20210225214923-2e10b2664254
	github.com/iancoleman/strcase v0.1.3
	github.com/imdario/mergo v0.3.12
	github.com/jgautheron/goconst v1.4.0
	github.com/kr/text v0.2.0 // indirect
	github.com/labstack/echo/v4 v4.2.1
	github.com/labstack/gommon v0.3.0
	github.com/laurent22/ical-go v0.1.1-0.20181107184520-7e5d6ade8eef
	github.com/lib/pq v1.10.0
	github.com/magefile/mage v1.11.0
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-sqlite3 v1.14.6
	github.com/mitchellh/mapstructure v1.3.2 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/olekukonko/tablewriter v0.0.5
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20200921180117-858c6e7e6b7e // indirect
	github.com/pquerna/otp v1.3.0
	github.com/prometheus/client_golang v1.9.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/samedi/caldav-go v3.0.0+incompatible
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20200824052919-0d455de96546
	github.com/spf13/afero v1.5.1
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/swag v1.7.0
	github.com/ulule/limiter/v3 v3.8.0
	golang.org/x/crypto v0.0.0-20210317152858-513c2a44f670
	golang.org/x/image v0.0.0-20210220032944-ac19c3e999fb
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5
	golang.org/x/oauth2 v0.0.0-20210313182246-cd4f82c27b84
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	golang.org/x/term v0.0.0-20210317153231-de623e64d2a6
	golang.org/x/text v0.3.5 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/d4l3k/messagediff.v1 v1.2.1
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/ini.v1 v1.57.0 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200605160147-a5ece683394c
	honnef.co/go/tools v0.0.1-2020.1.5
	src.techknowlogick.com/xgo v1.4.1-0.20210311222705-d25c33fcd864
	src.techknowlogick.com/xormigrate v1.4.0
	xorm.io/builder v0.3.8
	xorm.io/core v0.7.3
	xorm.io/xorm v1.0.7
)

replace (
	github.com/adlio/trello => github.com/kolaente/trello v1.7.1-0.20201216234312-5c4ef79b531e
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	github.com/hpcloud/tail => github.com/jeffbean/tail v1.0.1 // See https://github.com/hpcloud/tail/pull/159
	github.com/samedi/caldav-go => github.com/kolaente/caldav-go v3.0.1-0.20190524174923-9e5cd1688227+incompatible // Branch: feature/dynamic-supported-components, PR: https://github.com/samedi/caldav-go/pull/6 and https://github.com/samedi/caldav-go/pull/7
	gopkg.in/fsnotify.v1 => github.com/kolaente/fsnotify v1.4.10-0.20200411160148-1bc3c8ff4048 // See https://github.com/fsnotify/fsnotify/issues/328 and https://github.com/golang/go/issues/26904
)

go 1.13
