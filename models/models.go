package models

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql" // Because.
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3" // Because.
	"github.com/spf13/viper"
)

var (
	x *xorm.Engine

	tables []interface{}
)

func getEngine() (*xorm.Engine, error) {
	// Use Mysql if set
	if viper.GetString("database.type") == "mysql" {
		connStr := fmt.Sprintf(
			"%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true",
			viper.GetString("database.user"),
			viper.GetString("database.password"),
			viper.GetString("database.host"),
			viper.GetString("database.database"))
		return xorm.NewEngine("mysql", connStr)
	}

	// Otherwise use sqlite
	path := viper.GetString("database.path")
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
		new(NamespaceUser),
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

	x.ShowSQL(viper.GetBool("database.showqueries"))

	return nil
}
