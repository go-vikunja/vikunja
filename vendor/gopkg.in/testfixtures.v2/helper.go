package testfixtures

import (
	"database/sql"
	"fmt"
)

const (
	paramTypeDollar = iota + 1
	paramTypeQuestion
	paramTypeColon
)

type loadFunction func(tx *sql.Tx) error

// Helper is the generic interface for the database helper
type Helper interface {
	init(*sql.DB) error
	disableReferentialIntegrity(*sql.DB, loadFunction) error
	paramType() int
	databaseName(queryable) (string, error)
	tableNames(queryable) ([]string, error)
	isTableModified(queryable, string) (bool, error)
	afterLoad(queryable) error
	quoteKeyword(string) string
	whileInsertOnTable(*sql.Tx, string, func() error) error
}

type queryable interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
}

var (
	_ Helper = &MySQL{}
	_ Helper = &PostgreSQL{}
	_ Helper = &SQLite{}
	_ Helper = &Oracle{}
	_ Helper = &SQLServer{}
)

type baseHelper struct{}

func (baseHelper) init(_ *sql.DB) error {
	return nil
}

func (baseHelper) quoteKeyword(str string) string {
	return fmt.Sprintf(`"%s"`, str)
}

func (baseHelper) whileInsertOnTable(_ *sql.Tx, _ string, fn func() error) error {
	return fn()
}

func (baseHelper) isTableModified(_ queryable, _ string) (bool, error) {
	return true, nil
}

func (baseHelper) afterLoad(_ queryable) error {
	return nil
}
