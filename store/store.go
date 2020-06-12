package store

import (
	"fmt"
	"path"
	"runtime"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

//import _ "github.com/jinzhu/gorm/dialects/mysql"
//import _ "github.com/jinzhu/gorm/dialects/postgres"
//import _ "github.com/jinzhu/gorm/dialects/mssql"

var LOG = logrus.New()

func init() {
	// TODO: We should propably give the programmers more control about the logging
	// How?

	LOG.SetReportCaller(true)
	LOG.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})
	// LOG.SetLevel(logrus.DebugLevel)
}

type Store struct {
	Database *gorm.DB
}

func New(dbType string, databaseUrl string) (*Store, error) {
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

func (s *Store) Get(clientKey string) (*Tenant, error) {
	tenant := Tenant{}
	LOG.Debugf("Tenant with clientKey %s requested from database", clientKey)
	if result := s.Database.Where(&Tenant{ClientKey: clientKey}).First(&tenant); result.Error != nil {
		return nil, result.Error
	}
	LOG.Debugf("Got Tenant from Database: %+v", tenant)
	return &tenant, nil
}

func (s *Store) Set(tenant *Tenant) (*Tenant, error) {
	LOG.Debugf("Tenant %+v will be inserted or updated in database", tenant)

	optionalExistingRecord := Tenant{}
	if result := s.Database.Where(&Tenant{ClientKey: tenant.ClientKey}).First(&optionalExistingRecord); result.Error != nil {
		// If no entry matching the clientKey exists, insert the tenant,
		// otherwise update the tenant
		LOG.Debugf("Tenant %+v will be inserted in database", tenant)
		if result := s.Database.Create(tenant); result.Error != nil {
			return nil, result.Error
		}
	} else {
		LOG.Debugf("Tenant %+v will be updated in database", tenant)
		if result := s.Database.Model(tenant).Where(&Tenant{ClientKey: tenant.ClientKey}).Updates(tenant).Update("AddonInstalled", tenant.AddonInstalled); result.Error != nil {
			return nil, result.Error
		}
	}

	LOG.Debugf("Tenant %+v successfully inserted or updated", tenant)
	return tenant, nil
}
