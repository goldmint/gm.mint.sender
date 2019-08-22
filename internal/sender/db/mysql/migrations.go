package mysql

import (
	"github.com/jinzhu/gorm"
	"github.com/void616/gm-mint-sender/internal/sender/db/mysql/model"
	gormigrate "gopkg.in/gormigrate.v1"
)

// AutoMigration (dev only)
var AutoMigration = []interface{}{
	&model.Wallet{},
	&model.Sending{},
}

// Migrations array
var Migrations = []*gormigrate.Migration{

	// initial
	&gormigrate.Migration{
		ID: "2019-04-08T10:56:35.838Z",
		Migrate: func(tx *gorm.DB) error {
			return tx.
				CreateTable(&model.Wallet{}).
				CreateTable(&model.Sending{}).
				AddUniqueIndex("ux_snd_sendings_requestid", "request_id").
				AddIndex("ix_snd_sendings_status", "status").
				AddIndex("ix_snd_sendings_sentatblock", "sent_at_block").
				Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.
				DropTable(&model.Sending{}).
				DropTable(&model.Wallet{}).
				Error
		},
	},

	// notification flag as a separated from status field
	&gormigrate.Migration{
		ID: "2019-08-21T18:08:12.929Z",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&model.Sending{}).Error; err != nil {
				return err
			}
			return tx.
				Model(&model.Sending{}).
				Where("`status`=3"). // `notified`
				Updates(map[string]interface{}{
					"status":   2, // confirmed
					"notified": 1,
				}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
}
