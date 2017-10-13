package mysql

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // run init method
)

type Driver struct {
	db *gorm.DB
}

func connectionString(host, port, username, password, database string) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		username, password, host, port, database,
	)
}

func NewDriver(host, port, username, password, database string) (*Driver, error) {
	db, err := gorm.Open("mysql", connectionString(host, port, username, password, database))
	if err != nil {
		return nil, err
	}
	db.DB().SetConnMaxLifetime(1 * time.Minute)

	driver := &Driver{
		db: db,
	}
	return driver, nil
}
