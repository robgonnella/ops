package test_util

import (
	"github.com/robgonnella/go-lanscan/pkg/network"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func GetDBConnection(dbFile string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})

	if err != nil {
		return nil, err
	}

	return db, err
}

func Migrate(db *gorm.DB, model interface{}) error {
	return db.AutoMigrate(&model)
}

func GetTestInterfaceName() (string, error) {
	netInfo, err := network.NewDefaultNetwork()

	if err != nil {
		return "", err
	}

	return netInfo.Interface().Name, nil
}
