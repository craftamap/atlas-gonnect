package gonnect

import (
	"github.com/jinzhu/gorm"
	//import _ "github.com/jinzhu/gorm/dialects/mysql"
	//import _ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

//import _ "github.com/jinzhu/gorm/dialects/mssql"

type Store struct {
	Database *gorm.DB
}

func NewStore(dbType string, databaseUrl string) (*Store, error) {
	LOG.Info("Initializing Database Connection")
	db, err := gorm.Open(dbType, databaseUrl)
	if err != nil {
		return nil, err
	}

	LOG.Debug("Migrating Database Schemas")
	db.AutoMigrate(&Tenant{})

	LOG.Info("Database Connection initialized")
	return &Store{db}, nil
}

func (s *Store) get(clientKey string) (*Tenant, error) {
	tenant := Tenant{}
	LOG.Debugf("Tenant with clientKey %s requested from database", clientKey)
	if result := s.Database.Where(&Tenant{ClientKey: clientKey}).First(&tenant); result != nil {
		return nil, result.Error
	}
	LOG.Debugf("Got Tenant from Database: %+v", tenant)
	return &tenant, nil
}

func (s *Store) set(tenant *Tenant) (*Tenant, error) {
	LOG.Debugf("Tenant %+v will be inserted or updated in database", tenant)
	if s.Database.NewRecord(tenant) {
		LOG.Debugf("Tenant %+v will be inserted in database", tenant)
		if result := s.Database.Create(tenant); result.Error != nil {
			return nil, result.Error
		}
	} else {
		LOG.Debugf("Tenant %+v will be updated in database", tenant)
		if result := s.Database.Save(tenant); result.Error != nil {
			return nil, result.Error
		}
	}

	LOG.Debugf("Tenant %+v successfully inserted or updated", tenant)
	return tenant, nil
}
