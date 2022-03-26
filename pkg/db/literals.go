package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

var categoriesLit = []Category{
	{Name: "Groceries"},
	{Name: "Food"},
	{Name: "Clothing"},
	{Name: "Snacking"},
}

//
//
//
//
var entries = []Entry{

	// id 1-3
	{
		Name:     "Soap",
		Comment:  "new cat - name exists",
		Category: Category{Name: "Groceries"},
	},
	{
		Name:     "Toothpaste",
		Comment:  "new cat - name not exists",
		Category: Category{Name: fmt.Sprintf("Groceries-%v", time.Now().Unix())},
	},
	{
		Name:       "WC Cleaner",
		Comment:    "cat by ID",
		CategoryID: Category{}.IDByName("Groceries"),
	},

	// id 4,5
	{
		Name:       "Coffee",
		Comment:    "two new credit cards",
		CategoryID: Category{}.IDByName("Snacking"),
		CreditCards: []CreditCard{
			{Issuer: "VISA", Number: 232233339090},
			{Issuer: "AMEX", Number: 909090909090},
		},
	},
	{
		Name:       "Cookie",
		Comment:    "same new credit card - independent",
		CategoryID: Category{}.IDByName("Snacking"),
		CreditCards: []CreditCard{
			{Issuer: "VISA", Number: 232233339090}, // gets duplicated
		},
	},

	// explicit ID
	{
		Name:       "Apple Pie",
		CategoryID: Category{}.IDByName("Snacking"),
		Model:      gorm.Model{ID: uint(13)},
	},
	{
		Name:       "Nougat",
		Comment:    "three new tags",
		CategoryID: Category{}.IDByName("Snacking"),
		Model:      gorm.Model{ID: uint(14)},
		Tags: []Tag{
			{Name: "Indulgence"},
			{Name: "Reward"},
			{Name: "Craving"},
		},
	},
	{
		Name:       "Marzipan",
		Comment:    "three new tags again",
		CategoryID: Category{}.IDByName("Snacking"),
		Model:      gorm.Model{ID: uint(15)},
		Tags: []Tag{
			{Name: "Indulgence"}, // neither inserted (unique) nor omitted; m-n table contains wrong entry with TagID 0
			{Name: "Reward"},
			{Name: "Craving"},
			{Name: "Sloth"}, // is this added, though the others fail?
		},
	},

	// xxx
}
