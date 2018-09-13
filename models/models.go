package models

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql" // Because.
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3" // Because.
	xrc "github.com/go-xorm/xorm-redis-cache"

	"github.com/spf13/viper"
	"encoding/gob"
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
	if viper.GetBool("cache.enabled") {
		switch viper.GetString("cache.type") {
		case "memory":
			cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), viper.GetInt("cache.maxelementsize"))
			x.SetDefaultCacher(cacher)
			break
		case "redis":
			cacher := xrc.NewRedisCacher(viper.GetString("cache.redishost"), viper.GetString("cache.redispassword"), xrc.DEFAULT_EXPIRATION, x.Logger())
			x.SetDefaultCacher(cacher)
			gob.Register(tables)
			break
		default:
			fmt.Println("Did not find a valid cache type. Caching disabled. Please refer to the docs for poosible cache types.")
		}
	}

	x.SetMapper(core.GonicMapper{})

	// Sync dat shit
	if err = x.StoreEngine("InnoDB").Sync2(tables...); err != nil {
		return fmt.Errorf("sync database struct error: %v", err)
	}

	x.ShowSQL(viper.GetBool("database.showqueries"))

	return nil
}
