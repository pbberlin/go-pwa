package db

import (
	"encoding/json"
	"fmt"
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

	Tags []Tag
}

type EntryShortJSON struct {
	Content  string
	Category string
	TagNames string
}

func (e Entry) MarshalJSON() ([]byte, error) {

	et := EntryShortJSON{}
	et.Content = fmt.Sprint(e.ID, ":  ", e.Content)
	et.Category = fmt.Sprint(e.Category.ID, ":  ", e.Category.Name)

	nms := ""
	for _, tg := range e.Tags {
		nms += tg.Name + ", "
	}
	et.TagNames = nms

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
