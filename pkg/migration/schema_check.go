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

// checkPostgresSchemaMismatch refuses to run migrations when an existing install's tables
// live only in a schema other than the active one, instead of creating a second empty set (#3118).
func checkPostgresSchemaMismatch(x *xorm.Engine) error {
	if config.DatabaseType.GetString() != "postgres" {
		return nil
	}

	results, err := x.Query("SELECT COALESCE(current_schema(), '') AS current_schema")
	if err != nil {
		return fmt.Errorf("could not determine the current schema: %w", err)
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
		return fmt.Errorf("could not check for existing Vikunja tables: %w", err)
	}
	dataSchemas := make([]string, 0, len(results))
	for _, row := range results {
		dataSchemas = append(dataSchemas, string(row["schemaname"]))
	}

	return validateSchemaPlacement(config.DatabaseSchema.GetString(), currentSchema, dataSchemas)
}

func validateSchemaPlacement(configuredSchema, currentSchema string, dataSchemas []string) error {
	// current_schema() falls back to the next valid search_path entry (e.g. public) when the
	// configured schema does not exist, so compare against the configured value explicitly.
	if configuredSchema != "" && currentSchema != configuredSchema {
		return fmt.Errorf("the configured schema %q does not exist or is not accessible to the database user (active schema: %q). Create it or set database.schema (VIKUNJA_DATABASE_SCHEMA) to an existing schema", configuredSchema, currentSchema)
	}

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
