package models

import (
	"database/sql"
	"fmt"
	"github.com/asiainfoLDP/datafoundry_recharge/log"
	"os"
	"sync"
	"time"
)

var (
	logger = log.GetLogger()
)

func InitDB() {

	for i := 0; i < 3; i++ {
		connectDB()

		if DB() == nil {
			select {
			case <-time.After(time.Second * 10):
				continue
			}
		} else {
			break
		}
	}

	if DB() == nil {
		logger.Error("dbInstance is nil.")
		return
	}

	upgradeDB()

	go updateDB()

	logger.Info("Init db succeed.")
	return
}

func updateDB() {
	var err error
	ticker := time.Tick(5 * time.Second)
	for range ticker {
		db := GetDB()
		if db == nil {
			connectDB()
		} else if err = db.Ping(); err != nil {
			db.Close()
			// setDB(nil) // draw snake feet
			connectDB()
		}
	}
}

func GetDB() *sql.DB {
	if IsServing() {
		dbMutex.Lock()
		defer dbMutex.Unlock()
		return dbInstance
	} else {
		return nil
	}
}

func setDB(db *sql.DB) {
	dbMutex.Lock()
	dbInstance = db
	dbMutex.Unlock()
}

var (
	dbInstance *sql.DB
	dbMutex    sync.Mutex
)

func DB() *sql.DB {
	return dbInstance
}

func connectDB() {
	DB_ADDR, DB_PORT := MysqlAddrPort()
	DB_DATABASE, DB_USER, DB_PASSWORD := MysqlDatabaseUsernamePassword()
	logger.Info("Mysql_addr: %s\n"+
		"Mysql_port: %s\n"+
		"Myql_database: %s\n"+
		"Mysql_user: %s\n"+
		"Mysql_password: %s", DB_ADDR, DB_PORT, DB_DATABASE, DB_USER, DB_PASSWORD)

	DB_URL := fmt.Sprintf(`%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true`, DB_USER, DB_PASSWORD, DB_ADDR, DB_PORT, DB_DATABASE)

	logger.Info("connect to %s.", DB_URL)
	db, err := sql.Open("mysql", DB_URL) // ! here, err is always nil, db is never nil.
	if err == nil {
		err = db.Ping()
	}

	if err != nil {
		logger.Error("connect db error: %s.", err)
		//logger.Alert("connect db error: %s.", err)
	} else {
		setDB(db)
	}
}

func upgradeDB() {
	err := TryToUpgradeDatabase(DB(), "datafoundry:recharge", os.Getenv("MYSQL_CONFIG_DONT_UPGRADE_TABLES") != "yes") // don't change the name
	if err != nil {
		logger.Error("TryToUpgradeDatabase error: %v.", err)
	}
}

func MysqlAddrPort() (string, string) {
	return os.Getenv(os.Getenv("ENV_NAME_MYSQL_ADDR")),
		os.Getenv(os.Getenv("ENV_NAME_MYSQL_PORT"))
}

func MysqlDatabaseUsernamePassword() (string, string, string) {

	return os.Getenv(os.Getenv("ENV_NAME_MYSQL_DATABASE")),
		os.Getenv(os.Getenv("ENV_NAME_MYSQL_USER")),
		os.Getenv(os.Getenv("ENV_NAME_MYSQL_PASSWORD"))

}
