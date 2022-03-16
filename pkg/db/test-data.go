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

	//
	//
	cats := []Category{
		{Name: "Groceries"},
		{Name: "Food"},
		{Name: "Clothing"},
		{Name: "Snacking"},
	}
	for _, cat := range cats {
		res := onConflictUpdateAll.Create(&cat)
		LogRes(res)
	}

	//
	//
	//
	//
	entries := []Entry{

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
			CategoryID: CategoryIDByName("Groceries"),
		},
		{
			Content:    "WC Cleanser",
			CategoryID: CategoryIDByName("Groceries"),
		},
		{
			Content:    "Coffee",
			CategoryID: CategoryIDByName("Snacking"),
		},
		{
			Content:    "Cookie",
			CategoryID: CategoryIDByName("Snacking"),
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
			CategoryID: CategoryIDByName("Snacking"),
			Model:      gorm.Model{ID: uint(13)},
		},
		{
			Content:    "Apple Pie 14",
			CategoryID: CategoryIDByName("Snacking"),
			Model:      gorm.Model{ID: uint(14)},
		},
	}

	for idx, entry := range entries {
		if entry.Model.ID < 1 {
			entry.Model = gorm.Model{ID: uint(idx + 1)}
		}
		// res := db.Create(&entry)
		// res := onDuplicateID.Create(&entry)
		res := onConflictUpdateAll.Create(&entry)
		LogRes(res)
		log.Printf("finished entry %v of %v - %v\n", idx+1, len(entries), entry.Content)
	}

	//
	if false {
		// retrieves no categories
		entries := []Entry{}
		res := db.Find(&entries)
		LogRes(res)
		dbg.Dump(entries[:5])
	}
	if false {
		// works - "Category" is the struct field name
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

		for idx, entry := range entries {
			tags := []Tag{
				{Name: "Tag1", CategoryID: idx},
				{Name: "Tag2"},
			}
			err := db.Model(&entry).Association("Tags").Append(tags)
			// no error on composite index uniqueness failure
			LogErr(err)
			log.Printf("entry %2v: tags added to %v \n", idx+1, entry.Content)
		}

		for idx, entry := range entries {
			cnt := db.Model(&entry).Association("Tags").Count()
			log.Printf("entry %2v: %v has  %v tags\n", idx+1, entry.Content, cnt)
		}
	}

}
