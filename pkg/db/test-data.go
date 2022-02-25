package db

import (
	"log"

	"github.com/pbberlin/go-pwa/pkg/stacktrace"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestData() {

	db := Get()

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Panic caught: %v", err)
			stacktrace.Log()
		}
	}()

	setUpdateCounter := clause.Assignments(
		map[string]interface{}{
			"upsert_counter": 4,
		},
	)
	incrementUpdateCounter := clause.Assignments(
		map[string]interface{}{
			// "upsert_counter": gorm.Expr("GREATEST(upsert_counter, VALUES(upsert_counter))"),
			"upsert_counter": gorm.Expr("upsert_counter+4"),
		},
	)
	_, _ = setUpdateCounter, incrementUpdateCounter

	onDuplicateName := db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: incrementUpdateCounter,
			// UpdateAll: true, // prevents DoUpdates from execution
		},
	)
	onDuplicateID := db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: incrementUpdateCounter,
			// UpdateAll: true, // prevents DoUpdates from execution
		},
	)

	// DoUpdates is not executed - despite conflict
	onConflictUpdateAll := db.Clauses(
		clause.OnConflict{
			UpdateAll: true,
			DoUpdates: incrementUpdateCounter,
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

		log.Printf("rows created/updated: %v", res.RowsAffected)
		if res.Error != nil {
			log.Printf("error %v", res.Error)
		}
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
			CategoryID: 1,
		},
		{

			Content: "WC Cleanser",
			// Category: Category{Name: "Groceries"},
			CategoryID: 1,
		},
		{

			Content: "Coffee",
			// Category: Category{Name: "Snacking"},
			CategoryID: 2,
		}, {

			Content: "Cookie",
			// Category: Category{Name: "Snacking"},
			CategoryID: 2,
		},
	}

	for idx, entry := range entries {

		//
		entry.Model = gorm.Model{ID: uint(idx + 1)}
		// res := db.Create(&entry)
		// res := onConflictUpdateAll.Create(&entry)
		res := onDuplicateID.Create(&entry)

		log.Printf("rows created/updated: %v", res.RowsAffected)
		if res.Error != nil {
			log.Printf("error %v", res.Error)
		}
		// log.Printf("statement \n %v", res.Statement)
	}

	a := 2 - 2
	x := 1 / a
	_ = x

}
