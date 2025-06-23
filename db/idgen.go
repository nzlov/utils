package db

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IDGen struct {
	Model
	Name string `gorm:"uniqueIndex"` // Unique name for ID generator
	Num  int64  // Current sequence number
}

// Gen generates a new ID for the given name
// Uses Upsert to atomically increment the sequence number
func GenID(ctx context.Context, name string) (int64, error) {
	db := For(ctx)
	idGen := IDGen{
		Name: name,
		Num:  1,
	}

	n := int64(0)
	return n, Tx(ctx, func(context.Context) error {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.Assignments(map[string]any{"num": gorm.Expr("id_gens.num + 1")}),
		},
			clause.Returning{
				Columns: []clause.Column{{Name: "num"}},
			},
		).Create(&idGen).Error; err != nil {
			return err
		}
		n = idGen.Num
		return nil
	})
}
