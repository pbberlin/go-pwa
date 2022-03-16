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

func (e Tag) MarshalJSON() ([]byte, error) {

	et := struct{ Name string }{}
	et.Name = e.Name

	j, err := json.Marshal(et)
	if err != nil {
		return nil, err
	}

	return j, nil
}
