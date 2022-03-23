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

	// normally only called once in main() after config load
	Initialize()

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
		res := onDuplicateName.Create(&cat)
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
			CreditCards: []CreditCard{
				{Issuer: "VISA", Number: 232233339090},
				{Issuer: "AMEX", Number: 909090909090},
			},
		},

		//
		{
			Content:    "Toothpaste",
			CategoryID: CategoryIDByName("Groceries"),
			CreditCards: []CreditCard{
				{Issuer: "VISA", Number: 232233339090},
			},
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

		{
			Content:    "Apple Pie 15",
			CategoryID: CategoryIDByName("Snacking"),
			Tags: []Tag{
				{Name: "Indulgence"},
				{Name: "Reward"},
				{Name: "Craving"},
			},
		},

		// xxx
	}

	for idx, entry := range entries {
		if entry.Model.ID < 1 {
			entry.Model = gorm.Model{ID: uint(idx + 1)}
		}
		// res := db.Create(&entry)

		res := onDuplicateID.Create(&entry)
		LogRes(res)

		// entry now contains IDs of associations
		log.Printf("upserted entry %v of %v - %v\n", idx+1, len(entries), entry.Content)
	}

	//
	if false {
		// retrieves no categories, we need db preload
		entries := []Entry{}
		res := db.Find(&entries)
		LogRes(res)
		dbg.Dump(entries[:5])
	}

	{
		entries := []Entry{}
		res := db.Preload(clause.Associations).Find(&entries)
		LogRes(res)
		// dbg.Dump(entries[:4])
		// dbg.Dump(entries[5:])
		dbg.Dump(entries)

		for idx, entry := range entries {

			if true {
				tags := []Tag{
					{Name: "Tag1", CategoryID: idx},
					{Name: "Tag2"},
				}

				err := db.Model(&entry).Association("Tags").Append(tags)
				// no error on composite index uniqueness failure
				LogErr(err)
				log.Printf("entry %2v: tags added to %v \n", idx+1, entry.Content)
			}

			entry.UpsertCounter++
			// res := onDuplicateID.Create(&entry)
			// res := db.UpdateColumn("UpsertCounter", entry.UpsertCounter)
			res := db.Save(&entry)
			LogRes(res)

		}

		for idx, entry := range entries {
			cnt := db.Model(&entry).Association("Tags").Count()
			log.Printf("entry %2v: %v has  %v tags\n", idx+1, entry.Content, cnt)
		}
	}

	Close()
}
