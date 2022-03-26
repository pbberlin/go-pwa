package db

import (
	"encoding/json"
	"time"
)

type Tag struct { // multiple uses
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Name       string `gorm:"index:idx_tags_namecat,unique"` // unique composite index
	CategoryID uint   `gorm:"index:idx_tags_namecat,unique"` // unique composite index
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
