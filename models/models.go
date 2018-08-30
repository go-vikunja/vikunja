package models

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql" // Because.
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3" // Because.
)

var (
	x *xorm.Engine

	tables []interface{}
)

func getEngine() (*xorm.Engine, error) {
	// Use Mysql if set
	if Config.Database.Type == "mysql" {
		connStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true",
			Config.Database.User, Config.Database.Password, Config.Database.Host, Config.Database.Database)
		return xorm.NewEngine("mysql", connStr)
	}

	// Otherwise use sqlite
	path := Config.Database.Path
	if path == "" {
		path = "./db.db"
	}
	return xorm.NewEngine("sqlite3", path)
}

func init() {
	tables = append(tables,
		new(User),
		new(List),
		new(ListTask),
		new(Team),
		new(TeamMember),
		new(TeamList),
		new(TeamNamespace),
		new(Namespace),
		new(ListUser),
	)
}

// SetEngine sets the xorm.Engine
func SetEngine() (err error) {
	x, err = getEngine()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %v", err)
	}

	// Cache
	//cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
	//x.SetDefaultCacher(cacher)

	x.SetMapper(core.GonicMapper{})

	// Sync dat shit
	if err = x.StoreEngine("InnoDB").Sync2(tables...); err != nil {
		return fmt.Errorf("sync database struct error: %v", err)
	}

	x.ShowSQL(Config.Database.ShowQueries)

	return nil
}
