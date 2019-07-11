// Vikunja is a todo-list application to facilitate your life.
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
	cloud.google.com/go v0.34.0 // indirect
	code.vikunja.io/web v0.0.0-20190628075253-b457b5a1a332
	github.com/alecthomas/template v0.0.0-20160405071501-a0175ee3bccc
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/beevik/etree v1.1.0 // indirect
	github.com/client9/misspell v0.3.4
	github.com/d4l3k/messagediff v1.2.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fzipp/gocyclo v0.0.0-20150627053110-6acd4345c835
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.0 // indirect
	github.com/go-openapi/jsonreference v0.19.0 // indirect
	github.com/go-openapi/spec v0.19.0 // indirect
	github.com/go-openapi/swag v0.19.0 // indirect
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/go-sql-driver/mysql v1.4.1
	github.com/go-xorm/builder v0.3.4
	github.com/go-xorm/core v0.6.2
	github.com/go-xorm/tests v0.5.6 // indirect
	github.com/go-xorm/xorm v0.7.1
	github.com/go-xorm/xorm-redis-cache v0.0.0-20180727005610-859b313566b2
	github.com/gordonklaus/ineffassign v0.0.0-20180909121442-1003c8bd00dc
	github.com/imdario/mergo v0.3.7
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jgautheron/goconst v0.0.0-20170703170152-9740945f5dcb
	github.com/labstack/echo/v4 v4.1.7-0.20190627175217-8fb7b5be270f
	github.com/labstack/gommon v0.2.9
	github.com/laurent22/ical-go v0.1.1-0.20181107184520-7e5d6ade8eef
	github.com/mailru/easyjson v0.0.0-20190403194419-1ea4449da983 // indirect
	github.com/mattn/go-oci8 v0.0.0-20181130072307-052f5d97b9b6 // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mattn/go-sqlite3 v1.10.0
	github.com/olekukonko/tablewriter v0.0.1
	github.com/onsi/ginkgo v1.7.0 // indirect
	github.com/onsi/gomega v1.4.3 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/prometheus/client_golang v0.9.2
	github.com/samedi/caldav-go v3.0.0+incompatible
	github.com/shurcooL/httpfs v0.0.0-20190527155220-6a4d4a70508b
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.3.0
	github.com/swaggo/swag v1.5.0
	golang.org/x/crypto v0.0.0-20190621222207-cc06ce4a13d4
	golang.org/x/lint v0.0.0-20190409202823-959b441ac422
	google.golang.org/appengine v1.5.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/d4l3k/messagediff.v1 v1.2.1
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/testfixtures.v2 v2.5.3
	honnef.co/go/tools v0.0.0-20190418001031-e561f6794a2a
	src.techknowlogick.com/xgo v0.0.0-20190507142556-a5b29ecb0ff4
	src.techknowlogick.com/xormigrate v0.0.0-20190321151057-24497c23c09c
)

replace github.com/samedi/caldav-go => github.com/kolaente/caldav-go v3.0.1-0.20190524174923-9e5cd1688227+incompatible // Branch: feature/dynamic-supported-components, PR: https://github.com/samedi/caldav-go/pull/6 and https://github.com/samedi/caldav-go/pull/7
