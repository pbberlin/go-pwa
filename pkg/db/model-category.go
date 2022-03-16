package db

import (
	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	Name          string `gorm:"uniqueIndex"`
	UpsertCounter int

	// gorm.io/docs/belongs_to.html - would conflict with having the category on the Entry
	// Entry Entry `gorm:"foreignKey:CategoryID"`
}

type Categories []Category

var categories = []Category{} // if len(categories) > 20, switch to map

func CategoryIDByName(s string) int {

	// retrieving all objects
	// SELECT * FROM users;
	if len(categories) < 1 {
		res := db.Find(&categories)
		LogRes(res)
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
