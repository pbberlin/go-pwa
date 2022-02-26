package db

import (
	"errors"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

// https://gorm.io/docs/create.html#Create-With-Associations
type CreditCard struct {
	gorm.Model
	UserID     uint // foreign key to table user
	Number     string
	ValidUntil string
	Issuer     string `gorm:"default:VISA"`
}

type Pet struct {
	gorm.Model
	UserID uint   // foreign key to table user
	Name   string `gorm:"default:Castor"`
}

type User struct {
	gorm.Model
	FirstName  string
	Name       string
	Age        *int `gorm:"default:18"` // without pointer, 0 would not be stored
	CreditCard CreditCard
	Pets       []Pet
	// FullName   string `gorm:"->;type:GENERATED ALWAYS AS (concat(firstname,' ',name));default:(-);"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Name == "test" {
		return errors.New("invalid name")
	}
	return
}

func Initialize2(db *gorm.DB) {

	//
	//
	// ---------------------

	db.AutoMigrate(&Product{})

	// insert
	db.Create(&Product{Code: "D42", Price: 100})

	// read
	var product Product
	db.First(&product, 1)                 // find product with integer primary key
	db.First(&product, "code = ?", "D42") // find product with code D42

	//
	db.Model(&product).Update("Price", 200)
	// update multiple fields
	db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// db.Delete(&product, 1)

	db.AutoMigrate(&Pet{})
	db.AutoMigrate(&CreditCard{})
	db.AutoMigrate(&User{})

	db.Create(
		&User{
			Name: "Alice Sixpack",
			CreditCard: CreditCard{
				Number:     "411111111111",
				ValidUntil: "2024/04",
			},
		},
	)
	db.Create(
		&User{
			Name: "Bob Sinclair",
			CreditCard: CreditCard{
				Number:     "3216832168",
				ValidUntil: "2023/07",
			},
		},
	)

	//
	db.Omit("Name", "CreatedAt").Create(&User{Name: "no name"})
	db.Omit("CreditCard").Create(&User{Name: "Mr no credit card"})
	//
	usersBatch1 := []User{{Name: "Meyer"}, {Name: "Müller"}}
	db.CreateInBatches(usersBatch1, 100)

	usersBatch2 := []User{}
	for i := 0; i < 30; i++ {
		twpPets := []Pet{{}, {Name: "Pollux"}}
		usersBatch2 = append(usersBatch2, User{Name: "Schulze", Pets: twpPets})
	}
	db.Create(usersBatch2)

	// update all columns, except primary keys,
	// to new value on conflict
	for i := 0; i < 30; i++ {
		*usersBatch2[i].Age++
	}
	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&usersBatch2)

	// without onConflict
	for i := 0; i < 30; i++ {
		*usersBatch2[i].Age = -1
	}
	tx := db.Create(&usersBatch2)
	if tx.Error != nil {
		log.Printf("Create without conflict: %v", tx.Error)
	}

	// SELECT * FROM "orders" WHERE amount > (SELECT AVG(amount) FROM "orders");
	db.Where("age > (?)", db.Table("users").Select("AVG(age)")).Find(&usersBatch2)
}
