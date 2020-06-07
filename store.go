package gonnect

import "github.com/jinzhu/gorm"
//import _ "github.com/jinzhu/gorm/dialects/mysql"
//import _ "github.com/jinzhu/gorm/dialects/postgres"
import _ "github.com/jinzhu/gorm/dialects/sqlite"
//import _ "github.com/jinzhu/gorm/dialects/mssql"

type Store struct {
	Database *gorm.DB
}

func NewStore(dbType string, databaseUrl string) (*Store, error) {
	db, err := gorm.Open(dbType, databaseUrl)
	if err != nil {
		return nil, err
	}

	return &Store{db}, nil
}
