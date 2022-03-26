package db

import (
	"log"

	"github.com/pbberlin/dbg"

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

	setCompoundLiterals()

	log.Printf("------create-------")

	for idx, entry := range entriesLit {
		if entry.ID < 1 {
			entry.ID = uint(idx + 1)
		} else {
			log.Printf("id %v for %v", entry.ID, entry.Name)
		}
		res := onDuplicateID.Create(&entry)
		LogRes(res)
		// entry now contains IDs of associations
		// log.Printf("upserted entry %v of %v - %v\n", idx+1, len(entries), entry.Content)
	}

	//
	log.Printf("------adding tags----------")
	{
		res := db.Preload(clause.Associations).Find(&entriesLit)
		LogRes(res)

		for _, id := range []uint{13, 14} {

			e, err := Entry{}.ByID(id)
			if err != nil {
				log.Print(err)
				continue
			}

			tags := []Tag{
				{Name: "Tag1"},                 // only one is created because uniqueness
				{Name: "Tag2", CategoryID: id}, // each is created
			}

			err = db.Model(&e).Association("Tags").Append(tags)
			// no error on composite index uniqueness failure
			LogErr(err)
			log.Printf("entry %2v: tags added to %v \n", id, e.Name)

			e.UpsertCounter++
			// res := onDuplicateID.Create(&e)
			// res := db.UpdateColumn("UpsertCounter", e.UpsertCounter)
			res := db.Save(&e)
			LogRes(res)

		}

		// for idx, entry := range entries {
		// 	cnt := db.Model(&entry).Association("Tags").Count()
		// 	log.Printf("entry %2v: %v has  %v tags\n", idx+1, entry.Content, cnt)
		// }

	}

	//
	log.Printf("------save----------")
	{
		ToInfo()
		e, err := Entry{}.ByName("Apple Pie")
		if err != nil {
			log.Print(err)
		} else {
			e.Comment += " saved"
			res := db.Save(&e)
			LogRes(res)
		}
		// ToWarn()
	}
	{
		e := Entry{
			ID:         uint(16),
			Name:       "By Save 1",
			Comment:    "id 16, cat by ID",
			CategoryID: Category{}.IDByName("Food"),
		}
		res := db.Save(&e)
		LogRes(res)
	}
	{
		// ToInfo()
		e := Entry{
			Name:       "By Save 2",
			Comment:    "cat, ccs, tags by Val",
			CategoryID: Category{}.IDByName("Food"),
		}
		e.CreditCards = []CreditCard{
			{Issuer: "VISA", Number: 232233339090},
			{Issuer: "AMEX", Number: 909090909090},
		}
		e.Tags = []Tag{
			{Name: "Indulgence"},
			{Name: "Reward"},
			{Name: "Craving"},
			{Name: "Topor"},
		}

		res := db.Save(&e)
		LogRes(res)
		ToWarn()
	}

	//
	if false {
		// no categories, need db.Preload
		entries := []Entry{}
		res := db.Find(&entries)
		LogRes(res)
		dbg.Dump(entries[:5])
	}

	res := db.Preload(clause.Associations).Find(&entriesLit)
	LogRes(res)
	// dbg.Dump(entries[:4])
	// dbg.Dump(entries[5:])
	dbg.Dump(entriesLit)

	Close()
}
