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

	//Migrate the schema(s)
	db.AutoMigrate(&Tenant{})

	return &Store{db}, nil
}


func (s *Store) get(clientKey string) (*Tenant, error) {
	tenant := Tenant{}
	if result := s.Database.Where(&Tenant{ClientKey: clientKey}).First(&tenant); result != nil {
		return nil, result.Error
	}
	return &tenant, nil
}

func (s *Store) set(tenant *Tenant) (*Tenant, error) {
	if s.Database.NewRecord(tenant)	{
		if result := s.Database.Create(tenant); result != nil {
			return nil, result.Error
		}
	}  else {
		if result :=  s.Database.Save(tenant); result != nil {
			return nil, result.Error
		}
	}

	return tenant, nil
}
