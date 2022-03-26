package db

import (
	"log"

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
	for _, cat := range categoriesLit {
		res := onDuplicateName.Create(&cat)
		LogRes(res)
	}

	log.Printf("------create-------")

	for idx, entry := range entries {
		if entry.Model.ID < 1 {
			entry.Model = gorm.Model{ID: uint(idx + 1)}
		} else {
			log.Printf("id %v for %v", entry.ID, entry.Name)
		}
		res := onDuplicateID.Create(&entry)
		LogRes(res)
		// entry now contains IDs of associations
		// log.Printf("upserted entry %v of %v - %v\n", idx+1, len(entries), entry.Content)
	}

	log.Printf("------save----------")

	ToInfo()
	{
		e, err := Entry{}.ByName("Apple Pie")
		if err != nil {
			log.Print(err)
		} else {
			e.Name += " saved"
			res := db.Save(e)
			LogRes(res)
		}
	}
	{
		saved := Entry{
			Name:    "By Save 1",
			Comment: "cat by ID",
		}
		saved.CategoryID = Category{}.IDByName("Food")
		res := db.Save(&saved)
		LogRes(res)
	}
	{
		saved := Entry{
			Name:    "By Save 2",
			Comment: "cat, ccs, tags by Val",
		}
		saved.Category = Category{Name: "Food"}
		saved.CreditCards = []CreditCard{
			{Issuer: "VISA", Number: 232233339090},
			{Issuer: "AMEX", Number: 909090909090},
		}
		saved.Tags = []Tag{
			{Name: "Indulgence"},
			{Name: "Reward"},
			{Name: "Craving"},
		}

		res := db.Save(&saved)
		LogRes(res)
	}
	ToWarn()

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
				log.Printf("entry %2v: tags added to %v \n", idx+1, entry.Name)
			}

			entry.UpsertCounter++
			// res := onDuplicateID.Create(&entry)
			// res := db.UpdateColumn("UpsertCounter", entry.UpsertCounter)
			res := db.Save(&entry)
			LogRes(res)

			if idx > 2 {
				break
			}

		}

		// for idx, entry := range entries {
		// 	cnt := db.Model(&entry).Association("Tags").Count()
		// 	log.Printf("entry %2v: %v has  %v tags\n", idx+1, entry.Content, cnt)
		// }
	}

	Close()
}
