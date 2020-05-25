package db

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/riceChuang/tahmkench/config"
	"strconv"
)

var mysqlClient *gorm.DB

func SetUpMysqlClient(config *config.MysqlSetting, isUpdate bool) (client *gorm.DB, err error) {

	if len(config.Address) <= 0 {
		return
	}

	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true",
		config.Username, config.Password, config.Address, config.DBName)

	client, err = gorm.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}

	debugOpt, err := strconv.ParseBool(config.Debug)
	if err != nil {
		debugOpt = false
	}

	client.LogMode(debugOpt)
	client.SingularTable(true)

	if isUpdate {
		mysqlClient = client
	}
	return
}

func GetMysqlClient() *gorm.DB {
	return mysqlClient
}
