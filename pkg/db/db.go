package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// https://gorm.io/docs/

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func Initialize() {

	db, err := gorm.Open(sqlite.Open("./app-bucket/server-config/main.sqlite"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Product{})

	// insert
	db.Create(&Product{Code: "D42", Price: 100})

	// read
	var product Product
	db.First(&product, 1)                 // find product with integer primary key
	db.First(&product, "code = ?", "D42") // find product with code D42

	//
	db.Model(&product).Update("Price", 200)
	// update multiple fields
	db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// db.Delete(&product, 1)
}
