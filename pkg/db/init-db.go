package db

import (
	"fmt"
	"log"

	"github.com/pbberlin/dbg"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func LogRes(res *gorm.DB) {

	// https://golangbyexample.com/print-output-text-color-console/
	colorCyan := "\033[36m"
	colorRed := "\033[31m"
	colorReset := "\033[0m"

	// log.Printf("%2v stmt", res.Statement.SQL.String())
	if res.Error != nil {
		errStr := fmt.Sprintf("  %v", res.Error)
		log.Print(string(colorRed), errStr, res.Error, string(colorReset))
	}
	log.Print(string(colorCyan), dbg.CallingLine(0), string(colorReset))
	log.Printf("%2v affected rows", res.RowsAffected)
	// log.Printf("statement \n %v", res.Statement)
}

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

	if false {
		dbCfg.Logger = logger.Default.LogMode(logger.Info)
	}

	var err error
	db, err = gorm.Open(sqlite.Open("./app-bucket/server-config/main.sqlite"), dbCfg)
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Category{})
	db.AutoMigrate(&Tag{})
	db.AutoMigrate(&Entry{})

}
