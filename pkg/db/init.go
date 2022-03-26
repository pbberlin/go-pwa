package db

import (
	"fmt"
	"log"

	"github.com/pbberlin/dbg"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// https://golangbyexample.com/print-output-text-color-console/
const (
	colorCyan  = "\033[36m"
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

func LogRes(res *gorm.DB) {

	// log.Printf("%2v stmt", res.Statement.SQL.String())
	if res.Error != nil {
		errStr := fmt.Sprintf("  %v", res.Error)
		log.Print(colorRed, errStr, res.Error, colorReset)
	}

	if res.Error != nil || res.RowsAffected != 1 {
		log.Print(colorCyan, dbg.CallingLine(0), colorReset)
		log.Printf("%2v affected rows", res.RowsAffected)
	}

	// log.Printf("statement \n %v", res.Statement)
	// res.Error = nil
}

func LogErr(err error) {
	if err != nil {
		errStr := fmt.Sprintf("  %v", err)
		log.Print(colorCyan, dbg.CallingLine(0), colorReset)
		log.Print(colorRed, errStr, err, colorReset)
	}
}

var db *gorm.DB

func Get() *gorm.DB {
	return db
}

func ToInfo() {
	db.Config.Logger = logger.Default.LogMode(logger.Info)
}
func ToWarn() {
	db.Config.Logger = logger.Default.LogMode(logger.Warn)
}

// Initialize should be called on application start after config load
func Initialize() {

	if db != nil {
		// making sure, gorm.Open is called only once;
		// to close an existing db, use Close()
		return
	}

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

	// activate EntryTag as custom join table
	err = db.SetupJoinTable(&Entry{}, "Tags", &EntryTag{})
	LogErr(err)

	db.AutoMigrate(&Category{})
	db.AutoMigrate(&CreditCard{})
	db.AutoMigrate(&Tag{})
	db.AutoMigrate(&Entry{})

	initClauses()

}

// Close releases the database; as long as gorm.Open() was called only once on the db
func Close() {
	if db != nil {
		db.Commit()
		sqlDB, err := db.DB() // underlying golang sql.DB
		if err != nil {
			log.Printf("failed to get sql.DB from gorm.DB: %v", err)
			return
		}
		err = sqlDB.Close()
		if err != nil {
			log.Printf("cannot close sql.DB %v", err)
		} else {
			log.Printf("sql.DB closing...")
			db = nil
		}
	}
}
