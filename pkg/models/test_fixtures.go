package models

import (
	"gopkg.in/testfixtures.v2"
)

var fixtures *testfixtures.Context

// InitFixtures initialize test fixtures for a test database
func InitFixtures(helper testfixtures.Helper, dir string) (err error) {
	testfixtures.SkipDatabaseNameCheck(true)
	fixtures, err = testfixtures.NewFolder(x.DB().DB, helper, dir)
	return err
}

// LoadFixtures load fixtures for a test database
func LoadFixtures() error {
	return fixtures.Load()
}
