package testfixtures

import (
	"database/sql"
	"fmt"
)

// SQLServer is the helper for SQL Server for this package.
// SQL Server >= 2008 is required.
type SQLServer struct {
	baseHelper

	tables []string
}

func (h *SQLServer) init(db *sql.DB) error {
	var err error

	h.tables, err = h.tableNames(db)
	if err != nil {
		return err
	}

	return nil
}

func (*SQLServer) paramType() int {
	return paramTypeQuestion
}

func (*SQLServer) quoteKeyword(str string) string {
	return fmt.Sprintf("[%s]", str)
}

func (*SQLServer) databaseName(q queryable) (string, error) {
	var dbName string
	err := q.QueryRow("SELECT DB_NAME()").Scan(&dbName)
	return dbName, err
}

func (*SQLServer) tableNames(q queryable) ([]string, error) {
	rows, err := q.Query("SELECT table_name FROM information_schema.tables")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err = rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

func (*SQLServer) tableHasIdentityColumn(q queryable, tableName string) bool {
	sql := `
		SELECT COUNT(*)
		FROM SYS.IDENTITY_COLUMNS
		WHERE OBJECT_NAME(OBJECT_ID) = ?
	`
	var count int
	q.QueryRow(sql, tableName).Scan(&count)
	return count > 0

}

func (h *SQLServer) whileInsertOnTable(tx *sql.Tx, tableName string, fn func() error) (err error) {
	if h.tableHasIdentityColumn(tx, tableName) {
		defer func() {
			_, err2 := tx.Exec(fmt.Sprintf("SET IDENTITY_INSERT %s OFF", h.quoteKeyword(tableName)))
			if err2 != nil && err == nil {
				err = err2
			}
		}()

		_, err := tx.Exec(fmt.Sprintf("SET IDENTITY_INSERT %s ON", h.quoteKeyword(tableName)))
		if err != nil {
			return err
		}
	}
	return fn()
}

func (h *SQLServer) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) (err error) {
	// ensure the triggers are re-enable after all
	defer func() {
		var sql string
		for _, table := range h.tables {
			sql += fmt.Sprintf("ALTER TABLE %s WITH CHECK CHECK CONSTRAINT ALL;", h.quoteKeyword(table))
		}
		if _, err2 := db.Exec(sql); err2 != nil && err == nil {
			err = err2
		}
	}()

	var sql string
	for _, table := range h.tables {
		sql += fmt.Sprintf("ALTER TABLE %s NOCHECK CONSTRAINT ALL;", h.quoteKeyword(table))
	}
	if _, err := db.Exec(sql); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err = loadFn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
