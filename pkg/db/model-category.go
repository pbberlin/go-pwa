package db

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

type Category struct {
	ID            uint   `gorm:"primarykey"`
	Name          string `gorm:"uniqueIndex"`
	UpsertCounter int

	// gorm.io/docs/belongs_to.html - would conflict with having the category on the Entry
	// Entry Entry `gorm:"foreignKey:CategoryID"`
}

type Categories []Category

var categories = []Category{} // if len(categories) > 20, switch to map

// ByName returns by name
func (c Category) IDByName(s string) int {

	// SELECT * FROM categories;
	if len(categories) < 1 {
		res := db.Find(&categories)
		if res.Error != nil {
			errStr := fmt.Sprintf("  %v", res.Error)
			log.Print(colorRed, errStr, res.Error, colorReset)
		} else {
			log.Printf("%2v categories cached", res.RowsAffected)
		}
		// dbg.Dump(categories[:2])
		// dbg.Dump(categories[len(categories)-2:])
	}

	for _, cat := range categories {
		if s == cat.Name {
			return int(cat.ID)
		}
	}
	return 0
}

func (c *Category) BeforeCreate(tx *gorm.DB) (err error) {
	c.UpsertCounter = 10
	return nil
}
func (c *Category) BeforeUpdate(tx *gorm.DB) (err error) {
	c.UpsertCounter++
	return nil
}
