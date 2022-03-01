package db

import (
	"encoding/json"

	"gorm.io/gorm"
)

type TagU struct { // tag unique - belongs to one entry - no multiple uses
	gorm.Model
	Name    string `gorm:"uniqueIndex"`
	EntryID uint
}

type TagUShortJSON struct {
	Name string
}

func (e TagU) MarshalJSON() ([]byte, error) {

	et := TagUShortJSON{}
	et.Name = e.Name

	j, err := json.Marshal(et)
	if err != nil {
		return nil, err
	}

	return j, nil
}
