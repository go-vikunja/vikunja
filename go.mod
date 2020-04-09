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
	4d63.com/embedfiles v1.0.0 // indirect
	4d63.com/tz v1.1.0
	code.vikunja.io/web v0.0.0-20200208214421-c90649369427
	gitea.com/xorm/tests v0.5.6 // indirect
	gitea.com/xorm/xorm-redis-cache v0.0.0-20191113062523-5a6a9e2ab9f2
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/beevik/etree v1.1.0 // indirect
	github.com/c2h5oh/datasize v0.0.0-20200112174442-28bbd4740fee
	github.com/client9/misspell v0.3.4
	github.com/cweill/gotests v1.5.3
	github.com/d4l3k/messagediff v1.2.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fzipp/gocyclo v0.0.0-20150627053110-6acd4345c835
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/spec v0.19.4 // indirect
	github.com/go-redis/redis v6.14.0+incompatible
	github.com/go-redis/redis/v7 v7.2.0 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-testfixtures/testfixtures/v3 v3.1.1
	github.com/go-xorm/core v0.6.2 // indirect
	github.com/go-xorm/xorm v0.7.9 // indirect
	github.com/gordonklaus/ineffassign v0.0.0-20200309095847-7953dde2c7bf
	github.com/imdario/mergo v0.3.9
	github.com/jgautheron/goconst v0.0.0-20200227150835-cda7ea3bf591
	github.com/labstack/echo/v4 v4.1.16
	github.com/labstack/gommon v0.3.0
	github.com/laurent22/ical-go v0.1.1-0.20181107184520-7e5d6ade8eef
	github.com/lib/pq v1.3.0
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/olekukonko/tablewriter v0.0.4
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/prometheus/client_golang v0.9.4
	github.com/samedi/caldav-go v3.0.0+incompatible
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.7
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.5.1
	github.com/swaggo/swag v1.6.3
	github.com/ulule/limiter/v3 v3.3.0
	github.com/urfave/cli v1.22.2 // indirect
	golang.org/x/crypto v0.0.0-20200406173513-056763e48d71
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/d4l3k/messagediff.v1 v1.2.1
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	honnef.co/go/tools v0.0.0-20190418001031-e561f6794a2a
	src.techknowlogick.com/xgo v0.0.0-20200408234745-bb0faa361273
	src.techknowlogick.com/xormigrate v1.1.0
	xorm.io/builder v0.3.6
	xorm.io/core v0.7.3
	xorm.io/xorm v0.8.1
)

replace github.com/samedi/caldav-go => github.com/kolaente/caldav-go v3.0.1-0.20190524174923-9e5cd1688227+incompatible // Branch: feature/dynamic-supported-components, PR: https://github.com/samedi/caldav-go/pull/6 and https://github.com/samedi/caldav-go/pull/7

go 1.13
