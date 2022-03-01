package db

import (
	"fmt"
	"log"
	"time"

	"github.com/pbberlin/dbg"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestData() {

	db := Get()

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Panic 1 caught: %v", err)
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

		// {
		// 	Content: "Toothpaste without Cat",
		// },
		{
			Content:  "Toothpaste with value of cat - existing",
			Category: Category{Name: "Groceries"},
		},
		{
			Content:  "Toothpaste with value of cat - new",
			Category: Category{Name: fmt.Sprintf("Groceries-%v", time.Now().Unix())},
		},

		//
		{
			Content:    "Toothpaste",
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
			Tags:       []Tag{{Name: "Indulgence"}, {Name: "Curiosity"}, {Name: "Reward"}, {Name: "Craving"}},
		},

		// fails
		/* 		{
		   			Content:    "Apple Pie",
		   			CategoryID: CategoriesByName("Snacking"),
		   			Tags:       []Tag{{Name: "Indulgence"}, {Name: "Reward"}, {Name: "Craving"}},
		   		},
		*/
		// xxx
		{
			Content:    "Apple Pie 13",
			CategoryID: CategoriesByName("Snacking"),
			Model:      gorm.Model{ID: uint(13)},
			Tags: []Tag{
				{Model: gorm.Model{ID: uint(131)}, Name: "131"},
				{Model: gorm.Model{ID: uint(132)}, Name: "132"},
			},
		},
		{
			Content:    "Apple Pie 14",
			CategoryID: CategoriesByName("Snacking"),
			Model:      gorm.Model{ID: uint(14)},
			Tags: []Tag{
				{Model: gorm.Model{ID: uint(141)}, Name: "141"},
				{Model: gorm.Model{ID: uint(142)}, Name: "142"},
			},
		},
	}

	for idx, entry := range entries {
		if entry.Model.ID < 1 {
			entry.Model = gorm.Model{ID: uint(idx + 1)}
		}
		// res := db.Create(&entry)
		res := onDuplicateID.Create(&entry)
		LogRes(res)
		log.Printf("finished entry %v of %v - %v\n", idx+1, len(entries), entry.Content)
	}

	//
	if false {
		// retrives no categories
		entries := []Entry{}
		res := db.Find(&entries)
		LogRes(res)
		dbg.Dump(entries[:5])
	}
	if false {
		// works
		entries := []Entry{}
		res := db.Preload("Category").Find(&entries)
		LogRes(res)
		dbg.Dump(entries[:4])
	}

	{
		entries := []Entry{}
		res := db.Preload(clause.Associations).Find(&entries)
		LogRes(res)
		// dbg.Dump(entries[:4])
		// dbg.Dump(entries[5:])
		dbg.Dump(entries)
	}

}
