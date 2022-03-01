package db

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Tag struct { // multiple uses
	gorm.Model
	Name       string `gorm:"index:idx_namecat,unique"` // unique composite index
	CategoryID int    `gorm:"index:idx_namecat,unique"` // unique composite index
}

type TagShortJSON struct {
	Name string
}

func (e Tag) MarshalJSON() ([]byte, error) {

	et := TagUShortJSON{}
	et.Name = e.Name

	j, err := json.Marshal(et)
	if err != nil {
		return nil, err
	}

	return j, nil
}
