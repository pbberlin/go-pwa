package db

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Entry into the app
type Entry struct {
	gorm.Model
	Name          string
	Desc          string
	Comment       string
	UpsertCounter int

	// has one => primary key and object
	// gorm.io/docs/has_one.html
	// only a single category; category is not unique
	CategoryID int
	Category   Category

	// multiple tags - unique to this entry - for instance also addresses
	CreditCards []CreditCard

	Tags []Tag `gorm:"many2many:entry_tags;"` // multiple tags - reusable to other entries
}

// EntryTag for cutomized M to N table
//   established via SetupJoinTable()
type EntryTag struct {
	EntryID   int `gorm:"primaryKey"`
	TagID     int `gorm:"primaryKey"`
	Type      string
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}

// ByName returns by name
func (e Entry) ByName(s string) (Entry, error) {
	for i := 0; i < len(entries); i++ {
		if entries[i].Name == s {
			return entries[i], nil
		}
	}
	return Entry{}, fmt.Errorf("Entry %q not found", s)
}

// MarshalJSON only essential data
func (e Entry) MarshalJSON() ([]byte, error) {

	et := struct {
		Cnt  string
		Cat  string
		CCs  string
		Tags string
	}{}
	et.Cnt = fmt.Sprintf("ID%v: %v - %v", e.ID, e.Name, e.Comment)
	if e.Category.ID > 0 {
		et.Cat = fmt.Sprintf("ID%v: %v", e.Category.ID, e.Category.Name)
	}

	ccs := ""
	for _, cc := range e.CreditCards {
		ccs = fmt.Sprintf("%v;   ID%v-%v-%v", ccs, cc.ID, cc.Issuer, cc.Number)
	}
	et.CCs = ccs

	tgs := ""
	for _, tg := range e.Tags {
		tgs = fmt.Sprintf("%v;   ID%v-%v-%v", tgs, tg.ID, tg.Name, tg.CategoryID)
	}
	et.Tags = tgs

	j, err := json.Marshal(et)
	if err != nil {
		return nil, err
	}

	js := strings.Split(string(j), "\n")
	js2 := []string{}
	for _, row := range js {
		if strings.Contains(row, `"Cat":""`) {
			row = strings.ReplaceAll(row, `"Cat":""`, "")
		}
		if strings.Contains(row, `"CCs":""`) {
			row = strings.ReplaceAll(row, `"CCs":""`, "")
			// continue
		}
		if strings.Contains(row, `"Tags":""`) {
			row = strings.ReplaceAll(row, `"Tags":""`, "")
			// continue
		}
		row = strings.ReplaceAll(row, ",,,", ",")
		row = strings.ReplaceAll(row, ",,", ",")
		row = strings.ReplaceAll(row, ",}", "}")
		if false {
			log.Printf("\trow\t%v", row)
		}
		js2 = append(js2, row)
	}

	ret := []byte(strings.Join(js2, "\n"))
	return ret, nil
}

func (e *Entry) BeforeCreate(tx *gorm.DB) (err error) {
	e.UpsertCounter += 10
	return nil
}
func (e *Entry) BeforeUpdate(tx *gorm.DB) (err error) {
	e.UpsertCounter++
	return nil
}
