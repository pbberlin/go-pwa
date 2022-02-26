package db

import (
	"log"

	"github.com/pbberlin/go-pwa/pkg/dbg"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func LogRes(res *gorm.DB) {
	log.Printf("rows created/updated: %v", res.RowsAffected)
	if res.Error != nil {
		log.Printf("error %v", res.Error)
	}
	// log.Printf("statement \n %v", res.Statement)
}

func TestData() {

	db := Get()

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Panic caught: %v", err)
			dbg.StackTrace()
		}
	}()

	counterSet := clause.Assignments(
		map[string]interface{}{
			"upsert_counter": 4,
		},
	)
	counterInc := clause.Assignments(
		map[string]interface{}{
			// "upsert_counter": gorm.Expr("GREATEST(upsert_counter, VALUES(upsert_counter))"),
			"upsert_counter": gorm.Expr("upsert_counter+4"),
		},
	)
	_, _ = counterSet, counterInc

	onDuplicateName := db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: counterInc,
			// UpdateAll: true, // prevents DoUpdates from execution
		},
	)
	onDuplicateID := db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: counterInc,
			// UpdateAll: true, // prevents DoUpdates from execution
		},
	)

	// DoUpdates is not executed - despite conflict
	onConflictUpdateAll := db.Clauses(
		clause.OnConflict{
			UpdateAll: true,
			DoUpdates: counterInc,
		},
	)
	_ = onConflictUpdateAll

	//
	//
	cats := []Category{
		{Name: "Groceries"},
		{Name: "Food"},
		{Name: "Clothing"},
		{Name: "Snacking"},
	}
	for _, cat := range cats {

		res := onDuplicateName.Create(&cat)
		LogRes(res)
	}

	//
	//
	//
	//
	entries := []Entry{
		{
			Content: "Tootpaste without Cat",
		},
		{
			Content: "Tootpaste",
			// Category: Category{Name: "Groceries"},
			CategoryID: CategoriesByName("Groceries"),
		},
		{

			Content:    "WC Cleanser",
			CategoryID: CategoriesByName("Groceries"),
		},
		{

			Content:    "Coffee",
			CategoryID: CategoriesByName("Snacking"),
		},
		{

			Content:    "Cookie",
			CategoryID: CategoriesByName("Snacking"),
		},
		{

			Content:    "Apple Pie",
			CategoryID: CategoriesByName("Snacking"),
		},
	}

	for idx, entry := range entries {

		//
		entry.Model = gorm.Model{ID: uint(idx + 1)}
		// res := db.Create(&entry)
		res := onDuplicateID.Create(&entry)
		LogRes(res)

	}

}
