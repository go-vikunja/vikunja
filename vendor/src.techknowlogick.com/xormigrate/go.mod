module src.techknowlogick.com/xormigrate

go 1.12

require (
	github.com/denisenkom/go-mssqldb v0.0.0-20200206145737-bbfc9a55622e
	github.com/go-sql-driver/mysql v1.5.0
	github.com/joho/godotenv v1.3.0
	github.com/lib/pq v1.3.0
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/stretchr/testify v1.4.0
	google.golang.org/appengine v1.6.1 // indirect
	xorm.io/core v0.7.3 // indirect
	xorm.io/xorm v0.8.1
)

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20190330180304-d0100b6
