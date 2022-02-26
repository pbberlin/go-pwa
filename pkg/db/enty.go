package db

import (
	"encoding/json"
	"log"
	"strings"

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
}

type EntryShortJSON struct {
	ID      uint
	Content string

	CategoryID   uint
	CategoryName string
}

func (e Entry) MarshalJSON() ([]byte, error) {

	et := EntryShortJSON{}
	et.ID = e.ID
	et.Content = e.Content
	et.CategoryID = e.Category.ID
	et.CategoryName = e.Category.Name

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
		log.Print(row)
		js2 = append(js2, row)
	}

	ret := []byte(strings.Join(js2, "\n"))
	return ret, nil
}

// func (e *Entry) UnmarshalJSON(d []byte) error {
// 	t, err := time.Parse(customTimeFormat, string(d))
// 	if err != nil {
// 		return err
// 	}
// 	*e = customTime(t)
// 	return nil
// }
