package mysql

import (
	"github.com/jinzhu/gorm"
	"github.com/void616/gm-mint-sender/internal/watcher/db/mysql/model"
	gormigrate "gopkg.in/gormigrate.v1"
)

// AutoMigration (dev only)
var AutoMigration = []interface{}{
	&model.Wallet{},
	&model.Incoming{},
}

// Migrations array
var Migrations = []*gormigrate.Migration{

	// initial
	&gormigrate.Migration{
		ID: "2019-04-05T15:02:12.934Z",
		Migrate: func(tx *gorm.DB) error {
			return tx.
				CreateTable(&model.Wallet{}).
				CreateTable(&model.Incoming{}).AddIndex("ix_rfl_incomings_sent", "sent").
				Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.
				DropTable(&model.Wallet{}).
				DropTable(&model.Incoming{}).
				Error
		},
	},
}
