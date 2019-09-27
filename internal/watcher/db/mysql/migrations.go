package mysql

import (
	"github.com/jinzhu/gorm"
	"github.com/void616/gm-mint-sender/internal/watcher/db/mysql/model"
	gormigrate "gopkg.in/gormigrate.v1"
)

var migrations = []*gormigrate.Migration{

	// initial
	&gormigrate.Migration{
		ID: "2019-09-27T10:08:24.153Z",
		Migrate: func(tx *gorm.DB) error {
			return tx.
				CreateTable(&model.Service{}).
				AddUniqueIndex("ux_watcher_services_name", "name").
				CreateTable(&model.Wallet{}).
				AddUniqueIndex("ux_watcher_wallets_pubkeysvcid", "public_key", "service_id").
				AddForeignKey("service_id", tx.NewScope(&model.Service{}).TableName()+"(id)", "RESTRICT", "RESTRICT").
				CreateTable(&model.Incoming{}).
				AddUniqueIndex("ux_watcher_incomings_svcidtodigest", "service_id", "to", "digest").
				AddIndex("ix_watcher_incomings_notified", "notified").
				AddIndex("ix_watcher_incomings_notifyat", "notify_at").
				AddForeignKey("service_id", tx.NewScope(&model.Service{}).TableName()+"(id)", "RESTRICT", "RESTRICT").
				CreateTable(&model.Setting{}).
				Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.
				DropTable(&model.Wallet{}).
				DropTable(&model.Incoming{}).
				DropTable(&model.Setting{}).
				Error
		},
	},
}
