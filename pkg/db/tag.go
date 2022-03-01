package db

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Tag struct {
	gorm.Model
	Name    string `gorm:"uniqueIndex"`
	EntryID uint
}

type TagShortJSON struct {
	Name string
}

func (e Tag) MarshalJSON() ([]byte, error) {

	et := TagShortJSON{}
	et.Name = e.Name

	j, err := json.Marshal(et)
	if err != nil {
		return nil, err
	}

	return j, nil
}
