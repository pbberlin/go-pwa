package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var counterSet clause.Set
var counterInc clause.Set
var onDuplicateName *gorm.DB
var onDuplicateID *gorm.DB
var onConflictUpdateAll *gorm.DB

func initClauses() {

	//
	// assignments
	//   subset of clauses
	counterSet = clause.Assignments(
		map[string]interface{}{
			"upsert_counter": 4,
		},
	)
	counterInc = clause.Assignments(
		map[string]interface{}{
			// "upsert_counter": gorm.Expr("GREATEST(upsert_counter, VALUES(upsert_counter))"),
			"upsert_counter": gorm.Expr("upsert_counter+4"),
		},
	)

	//
	// clauses conflict
	onDuplicateName = db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			UpdateAll: true, // prevents DoUpdates from execution
			// DoUpdates: counterInc,
		},
	)
	onDuplicateID = db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true, // prevents DoUpdates from execution
			// DoUpdates: counterInc,
		},
	)

	onConflictUpdateAll = db.Clauses(
		clause.OnConflict{
			// REQUIRED Columns
			UpdateAll: true,
			// DoUpdates: counterInc,
		},
	)

}
