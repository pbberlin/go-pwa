package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func Get() *gorm.DB {
	return db
}

func Initialize() {

	dbCfg := &gorm.Config{
		CreateBatchSize: 10,

		// gorm.io/docs/associations.html#Association-Mode
		// FullSaveAssociations: true,
	}

	var err error
	db, err = gorm.Open(sqlite.Open("./app-bucket/server-config/main.sqlite"), dbCfg)
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Category{})
	db.AutoMigrate(&Entry{})

}
