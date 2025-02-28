package database

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Me-Nodeslist/database/logs"
	"github.com/mitchellh/go-homedir"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var GlobalDataBase *gorm.DB
var logger = logs.Logger("database")

var blockNumberKey = "block_number_key"

type DABlockNumber struct {
	BlockNumberKey string `gorm:"primarykey;column:key"`
	BlockNumber    int64
}

func InitDatabase(path string) error {
	dir, err := homedir.Expand(path)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0666)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
	}

	db, err := gorm.Open(sqlite.Open(filepath.Join(dir, "server.db")), &gorm.Config{})
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Second * 30)

	err = sqlDB.Ping()
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	db.AutoMigrate(&LicenseInfo{}, &DelMEMOMintInfo{}, &DelMEMOTransferInfo{}, &RedeemInfo{}, &NodeInfo{}, &NodeDailyDelegation{})
	GlobalDataBase = db
	return nil
}

func SetBlockNumber(blockNumber int64) error {
	var daBlockNumber = DABlockNumber{
		BlockNumberKey: blockNumberKey,
		BlockNumber:    blockNumber,
	}
	return GlobalDataBase.Save(&daBlockNumber).Error
}

func GetBlockNumber() (int64, error) {
	var blockNumber DABlockNumber
	err := GlobalDataBase.Model(&DABlockNumber{}).First(&blockNumber).Error

	return blockNumber.BlockNumber, err
}
