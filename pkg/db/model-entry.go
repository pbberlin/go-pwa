package db

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Entry into the app
type Entry struct {
	gorm.Model
	Content       string
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
//   just by name
type EntryTag struct {
	EntryID   int `gorm:"primaryKey"`
	TagID     int `gorm:"primaryKey"`
	Type      string
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}

// MarshalJSON only essential data
func (e Entry) MarshalJSON() ([]byte, error) {

	et := struct {
		Content   string
		Category  string
		TagUNames string
		TagNames  string
	}{}
	et.Content = fmt.Sprint(e.ID, ":  ", e.Content)
	et.Category = fmt.Sprint(e.Category.ID, ":  ", e.Category.Name)

	nms := ""
	// for _, tg := range e.TagsU {
	// 	nms += tg.Name + ", "
	// }
	et.TagUNames = nms

	j, err := json.Marshal(et)
	if err != nil {
		return nil, err
	}

	js := strings.Split(string(j), "\n")
	js2 := []string{}
	for _, row := range js {
		if strings.Contains(row, `"CategoryName": ""`) {
			continue
		}
		// log.Print(row)
		js2 = append(js2, row)
	}

	ret := []byte(strings.Join(js2, "\n"))
	return ret, nil
}

func (e *Entry) BeforeCreate(tx *gorm.DB) (err error) {
	e.UpsertCounter = 10
	return nil
}
func (e *Entry) BeforeUpdate(tx *gorm.DB) (err error) {
	e.UpsertCounter++
	return nil
}
