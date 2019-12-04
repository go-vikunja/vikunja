// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2019 Vikunja and contributors. All rights reserved.
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

// +build tools

package tools

// This file is needed for go mod to recognize the tools we use.

import (
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/cweill/gotests/..."
	_ "github.com/fzipp/gocyclo"
	_ "github.com/gordonklaus/ineffassign"
	_ "github.com/swaggo/swag/cmd/swag"
	_ "golang.org/x/lint/golint"
	_ "src.techknowlogick.com/xgo"

	_ "github.com/jgautheron/goconst/cmd/goconst"
	_ "honnef.co/go/tools/cmd/staticcheck"

	_ "github.com/shurcooL/vfsgen"
)
