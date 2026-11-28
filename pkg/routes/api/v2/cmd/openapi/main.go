// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Command openapi prints the canonical OpenAPI document.
// It is a program of its own so the magefile does not have to import, and with
// that compile, the whole API.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	apiv2 "code.vikunja.io/api/pkg/routes/api/v2"
)

func main() {
	api, err := apiv2.NewCanonicalAPI()
	if err == nil {
		err = json.NewEncoder(os.Stdout).Encode(api.OpenAPI())
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
