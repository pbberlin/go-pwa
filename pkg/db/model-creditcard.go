package db

import (
	"encoding/json"
)

type CreditCard struct { // tag unique - belongs to one entry - no multiple uses
	// gorm.Model
	ID uint `gorm:"primarykey"`

	// Issuer  string `gorm:"index:idx_credit_card1,unique;index:idx_credit_card2,unique"` // unique composite indexes over issuer, number - with and without entry ID
	// Number  uint   `gorm:"index:idx_credit_card1,unique;index:idx_credit_card2,unique"` //
	// EntryID uint   `gorm:"index:idx_credit_card1,unique"`                               //
	Issuer  string `gorm:"index:idx_credit_card1,unique"` // unique composite indexes over issuer, number - with and without entry ID
	Number  uint   `gorm:"index:idx_credit_card1,unique"` //
	EntryID uint   `gorm:"index:idx_credit_card1,unique"` //
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
