package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pbberlin/dbg"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gorm.io/gorm/clause"
)

//
// go test  -run ^TestBulk$ github.com/pbberlin/go-pwa/pkg/db -v
func TestBulk(t *testing.T) {

	defer func() {
		if err := recover(); err != nil {
			t.Logf("Panic 1 caught: %v", err)
			dbg.StackTrace()
		}
	}()

	// working dir for test is the package
	//   => switch to app dir
	wd, _ := os.Getwd()
	suffix := filepath.Join("pkg", "db")
	if strings.HasSuffix(wd, suffix) {
		os.Chdir(filepath.Join("..", "..")) // stepping up to app dir
		wd, _ = os.Getwd()
		t.Logf("new working dir is %v", wd)
	}

	testDB := "main_test"
	pth := fmt.Sprintf("./app-bucket/server-config/%v.sqlite", testDB)
	err := os.Remove(pth)
	if err != nil {
		if !os.IsNotExist(err) {
			t.Fatalf("Cannot remove previous test DB; still opened by SQLiteViewer? \n%v", err)
		}
	}

	Init(testDB)

	var pc *Category // for calling methods on categories
	var pe *Entry    // for calling methods on entries

	db := Get()

	//
	//
	for _, cat := range categoriesTestSeed {
		res := onDuplicateNameUpdate.Create(&cat)
		LogRes(res)
	}

	dynInitEntriesTestSeed()

	t.Logf("------create-------")

	for idx, e := range entriesTestSeed {
		if e.ID < 1 {
			e.ID = uint(idx + 1)
		} else {
			t.Logf("id %v for %v", e.ID, e.Name)
		}
		res := onDuplicateIDUpdate.Omit("Tags").Create(&e)
		LogRes(res)
	}

	//
	t.Logf("------adding tags using Association(...).Append(...)----------")
	entries = nil // forcing reload
	{

		for _, id := range []uint{13, 14} {

			e, err := pe.ByID(id)
			if err != nil {
				t.Log(err)
				continue
			}

			tags := []Tag{
				{Name: "Tag1"},                 // only one is created because uniqueness
				{Name: "Tag2", CategoryID: id}, // each is created
			}

			err = db.Model(&e).Association("Tags").Append(tags)
			// no error on composite index uniqueness failure
			LogErr(err)
			t.Logf("entry %2v: tags added to %v \n", id, e.Name)

			e.UpsertCounter++
			// res := onDuplicateID.Create(&e)
			// res := db.UpdateColumn("UpsertCounter", e.UpsertCounter)
			res := db.Save(&e)
			LogRes(res)

		}

		// for idx, entry := range entries {
		// 	cnt := db.Model(&entry).Association("Tags").Count()
		// 	t.Logf("entry %2v: %v has  %v tags\n", idx+1, entry.Content, cnt)
		// }

	}

	//
	t.Logf("------save----------")
	entries = nil // forcing reload

	{
		ToInfo()
		e, err := pe.ByName("Apple Pie")
		if err != nil {
			t.Log(err)
		} else {
			e.Comment += " saved"
			res := db.Save(&e)
			LogRes(res)
		}
		ToWarn()
	}
	{
		e := Entry{
			ID:         uint(16),
			Name:       "By Save 1",
			Comment:    "id 16, cat by ID",
			CategoryID: pc.ByName("Food"),
		}
		res := db.Save(&e)
		LogRes(res)
	}
	{
		e := Entry{
			Name:       "By Save 2",
			Comment:    "cat, ccs, tags by Val",
			CategoryID: pc.ByName("Food"),
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

		// save()
		{
			// id 17
			res := db.Save(&e)
			LogRes(res)
		}
		{
			e.Tags = []Tag{{Name: "Tag-Omitted-1"}}
			res := db.Omit("Tags").Save(&e)
			LogRes(res)
		}
		{
			e.Tags = []Tag{{Name: "Tag-Not-Omitted-1"}}
			res := db.Omit("TagsXX").Save(&e)
			LogRes(res)
		}
		{
			e.ID = 20 // this causes the credit cards with ID>0 being transferred to the new entry

			// new credit card associations
			e.CreditCards = []CreditCard{
				{Issuer: "VISA", Number: 232233339090},
				{Issuer: "AMEX", Number: 909090909090},
			}

			e.Comment = "save"
			{
				e.Tags = []Tag{{Name: "Tag-Omitted-2"}}
				res := db.Omit("Tags").Save(&e)
				LogRes(res)

			}
			{
				e.Tags = []Tag{{Name: "Tag-Not-Omitted-2"}}
				res := db.Omit("TagsXX").Save(&e)
				LogRes(res)
			}
		}

		t.Logf("------create and save again----------")
		{
			e.ID = 21 // this causes the credit cards with ID>0 being transferred to the new entry

			// new credit card associations
			e.CreditCards = []CreditCard{
				{Issuer: "VISA", Number: 232233339090},
				{Issuer: "AMEX", Number: 909090909090},
			}

			e.Comment = "create"
			{
				e.Tags = []Tag{{Name: "Tag-Omitted-3"}}
				res := onDuplicateIDUpdate.Omit("Tags").Create(&e)
				LogRes(res)
			}
			{
				e.Tags = []Tag{{Name: "Tag-Not-Omitted-3"}}
				res := onDuplicateIDUpdate.Omit("TagsXX").Create(&e)
				LogRes(res)
			}
		}
	}

	//
	if false {
		// no categories, need db.Preload
		entries := []Entry{}
		res := db.Find(&entries)
		LogRes(res)
		dbg.Dump(entries[:5])
	}

	// forcing reload
	res := db.Preload(clause.Associations).Find(&entries)
	LogRes(res)
	// dbg.Dump(entries[:4])
	// dbg.Dump(entries[5:])
	// dbg.Dump(&entries)

	err = os.MkdirAll("./app-bucket/tests", 0777)
	if err != nil {
		t.Fatalf("Cannot create ./app-bucket/tests \n%v", err)
	}

	got := dbg.Dump2String(&entries)
	err = os.WriteFile("./app-bucket/tests/tmp-gormtest_got.json", []byte(got), 0777)
	if err != nil {
		t.Fatalf("Cannot write ./app-bucket/tests/tmp-gormtest_got.json \n%v", err)
	}

	wntBts, err := os.ReadFile("./app-bucket/tests/gormtest_wnt.json")
	if err != nil {
		t.Fatalf("Cannot read ./app-bucket/tests/gormtest_wnt.json \n%v", err)
	}
	wnt := string(wntBts)

	//
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(got, wnt, false)
	t.Log(dmp.DiffPrettyText(diffs))

	Close()
}
