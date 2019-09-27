package mysql

import (
	"strings"

	mysqld "github.com/go-sql-driver/mysql"
	gormigrate "gopkg.in/gormigrate.v1"
)

// Available impl.
func (d *Database) Available() bool {
	return d.DB.DB().Ping() == nil
}

// DuplicateError impl.
func (d *Database) DuplicateError(err error) bool {
	if err != nil {
		if merr, yes := err.(*mysqld.MySQLError); yes {
			return merr.Number == 1062
		}
	}
	return false
}

// MaxPacketError impl.
func (d *Database) MaxPacketError(err error) bool {
	if err != nil {
		return strings.Contains(strings.ToLower(err.Error()), "max_allowed_packet")
	}
	return false
}

// Migrate implementation
func (d *Database) Migrate() error {
	opts := gormigrate.DefaultOptions
	opts.TableName = d.tablePrefix + "dbmigrations"
	mig := gormigrate.New(d.DB, opts, migrations)
	return mig.Migrate()
}
