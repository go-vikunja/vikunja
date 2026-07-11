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

package migration

import (
	"fmt"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"xorm.io/xorm"
)

// checkPostgresSchemaMismatch guards against installs whose data lives in a different
// schema than the one Vikunja operates on (e.g. a role-named schema picked up via the
// "$user" search_path default, or a pgloader import). Running migrations in that state
// would create a second, empty set of tables and report success (#3118).
func checkPostgresSchemaMismatch(x *xorm.Engine) error {
	if config.DatabaseType.GetString() != "postgres" {
		return nil
	}

	results, err := x.Query("SELECT COALESCE(current_schema(), '') AS current_schema")
	if err != nil {
		return err
	}
	currentSchema := ""
	if len(results) > 0 {
		currentSchema = string(results[0]["current_schema"])
	}

	// users + migration together identify a schema holding an existing Vikunja install.
	results, err = x.Query(`SELECT u.schemaname
		FROM pg_tables u
		JOIN pg_tables m ON m.schemaname = u.schemaname AND m.tablename = 'migration'
		WHERE u.tablename = 'users'`)
	if err != nil {
		return err
	}
	dataSchemas := make([]string, 0, len(results))
	for _, row := range results {
		dataSchemas = append(dataSchemas, string(row["schemaname"]))
	}

	return validateSchemaPlacement(currentSchema, dataSchemas)
}

func validateSchemaPlacement(currentSchema string, dataSchemas []string) error {
	others := make([]string, 0, len(dataSchemas))
	found := false
	for _, s := range dataSchemas {
		if s == currentSchema {
			found = true
			continue
		}
		others = append(others, s)
	}

	if len(dataSchemas) == 0 || found {
		if found && len(others) > 0 {
			log.Warningf("Found Vikunja tables in schema(s) %s in addition to the active schema %q. Vikunja will ignore them, but you may want to clean them up.", strings.Join(others, ", "), currentSchema)
		}
		return nil
	}

	if currentSchema == "" {
		return fmt.Errorf("the configured database schema does not exist, but existing Vikunja tables were found in schema(s) %s. Set database.schema (VIKUNJA_DATABASE_SCHEMA) to the schema containing your data", strings.Join(others, ", "))
	}

	return fmt.Errorf("existing Vikunja tables were found in schema(s) %s, but Vikunja is configured to use schema %q. Running migrations now would create a second, empty set of tables. Set database.schema (VIKUNJA_DATABASE_SCHEMA) to the schema containing your data", strings.Join(others, ", "), currentSchema)
}
