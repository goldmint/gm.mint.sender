package mysql

import (
	"time"

	mysqld "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	// mysql driver init
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Database data
type Database struct {
	*gorm.DB
	tablePrefix string
}

// New instance
func New(connection, tablePrefix string, multiStatements bool, maxPacket uint32) (*Database, error) {
	conf, err := mysqld.ParseDSN(connection)
	if err != nil {
		return nil, err
	}

	conf.MaxAllowedPacket = int(maxPacket)
	conf.Loc = time.UTC
	conf.ParseTime = true
	conf.MultiStatements = multiStatements

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return tablePrefix + defaultTableName
	}

	db, err := gorm.Open("mysql", conf.FormatDSN())
	if err != nil {
		return nil, err
	}

	return &Database{
		DB:          db,
		tablePrefix: tablePrefix,
	}, nil
}
