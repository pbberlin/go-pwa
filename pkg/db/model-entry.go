package db

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Entry struct {
	gorm.Model
	Content       string
	UpsertCounter int

	// Has one => primary key and object
	// gorm.io/docs/has_one.html
	CategoryID int
	Category   Category

	TagsU []TagU // multiple tags - unique to this entry

	Tags []Tag `gorm:"many2many:entry_tag;"` // multiple tags - reusable to other entries
}

// cutomized M to N table
type EntryTag struct {
	EntryID   int `gorm:"primaryKey"`
	TagID     int `gorm:"primaryKey"`
	Type      string
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func init() {
	// activate EntryTag as custom join table
	err := db.SetupJoinTable(&Entry{}, "Tags", &EntryTag{})
	LogErr(err)
}

type EntryShortJSON struct {
	Content   string
	Category  string
	TagUNames string
	TagNames  string
}

func (e Entry) MarshalJSON() ([]byte, error) {

	et := EntryShortJSON{}
	et.Content = fmt.Sprint(e.ID, ":  ", e.Content)
	et.Category = fmt.Sprint(e.Category.ID, ":  ", e.Category.Name)

	nms := ""
	for _, tg := range e.TagsU {
		nms += tg.Name + ", "
	}
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
