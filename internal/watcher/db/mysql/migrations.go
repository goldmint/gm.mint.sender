package mysql

import (
	"github.com/jinzhu/gorm"
	"github.com/void616/gm-mint-sender/internal/watcher/db/mysql/model"
	gormigrate "gopkg.in/gormigrate.v1"
)

// Migrations array
var Migrations = []*gormigrate.Migration{

	// initial
	&gormigrate.Migration{
		ID: "2019-08-28T18:09:42.107Z",
		Migrate: func(tx *gorm.DB) error {
			return tx.
				CreateTable(&model.Wallet{}).
				AddUniqueIndex("ux_watcher_wallets_publickeyservice", "public_key", "service").
				CreateTable(&model.Incoming{}).
				AddUniqueIndex("ux_watcher_incomings_servicetodigest", "service", "to", "digest").
				AddIndex("ix_watcher_incomings_notified", "notified").
				AddIndex("ix_watcher_incomings_notifyat", "notify_at").
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
