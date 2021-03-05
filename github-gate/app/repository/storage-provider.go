package repository

import (
	"fmt"
	"github-gate/pckg/runtimeinfo"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TypeStorage int

const (
	TypeStoragePostgres TypeStorage = 1
	TypeStorageMySql    TypeStorage = 2
)

type DSNTemplate string

const (
	// Username, Password, Proto, Address, Port, DBName
	DSNMySQL DSNTemplate = "%s:%s@%s(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local"
	// Username, Password, DBName, Port, SSLMode
	DSNPostgres DSNTemplate = "user=%s password=%s dbname=%s port=%s sslmode=%s"
)

type ApplicationStorageProvider struct {
	SqlDB  *gorm.DB
}

func SQLCreateConnection(typeStorage TypeStorage, template DSNTemplate, config *gorm.Config, params ...string) *ApplicationStorageProvider {
	interfaces := make([]interface{}, len(params))
	for i, v := range params {
		interfaces[i] = v
	}
	dsn := fmt.Sprintf(string(template), interfaces...)
	if config == nil {
		config = &gorm.Config{}
	}
	var openConnection *gorm.DB
	switch typeStorage {
	case TypeStoragePostgres:
		connect, err := gorm.Open(postgres.Open(dsn), config)
		if err != nil {
			runtimeinfo.LogFatal(err)
		}
		openConnection = connect
	case TypeStorageMySql:
		connect, err := gorm.Open(mysql.Open(dsn), config)
		if err != nil {
			runtimeinfo.LogFatal(err)
		}
		openConnection = connect
	default:
		runtimeinfo.LogFatal("Non valid DB type. ")
	}
	storage := new(ApplicationStorageProvider)
	storage.SqlDB = openConnection
	return storage
}

