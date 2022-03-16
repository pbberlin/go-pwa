package db

import (
	"encoding/json"

	"gorm.io/gorm"
)

type CreditCard struct { // tag unique - belongs to one entry - no multiple uses
	gorm.Model
	Issuer  string `gorm:"index:idx_creditcard,unique"` // unique composite index
	Number  uint   `gorm:"index:idx_creditcard,unique"` // unique composite index
	EntryID uint   `gorm:"index:idx_creditcard,unique"` // unique composite index
}

// MarshalJSON only essential data
func (e CreditCard) MarshalJSON() ([]byte, error) {

	et := struct{ Name string }{}
	et.Name = e.Issuer

	j, err := json.Marshal(et)
	if err != nil {
		return nil, err
	}

	return j, nil
}
