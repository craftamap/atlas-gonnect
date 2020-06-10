package store

import (
	"encoding/json"
	"io"
	"time"
)

type Tenant struct {
	ClientKey      string `json:"clientKey" gorm:"type:varchar(255);primary_key"`
	PublicKey      string `json:"publicKey" gorm:"type:varchar(512)"`
	SharedSecret   string `json:"sharedSecret" gorm:"type:varchar(255);NOT NULL"`
	OauthClientId  string `json:"oauthClientId" gorm:"type:varchar(255)"`
	BaseURL        string `json:"baseUrl" gorm:"type:varchar(255);NOT NULL"`
	ProductType    string `json:"productType" gorm:"type:varchar(255)"`
	Description    string `json:"description" gorm:"type:varchar(255)"`
	AddonInstalled bool   `json:"-" gorm:"type:bit;NOT NULL"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	EventType      string `json:"eventType" gorm:"-"`
}

func NewTenantFromReader(r io.Reader) (*Tenant, error) {
	tenant := &Tenant{}
	err := json.NewDecoder(r).Decode(tenant)
	if err != nil {
		return nil, err
	}
	if tenant.EventType == "installed" {
		tenant.AddonInstalled = true
	} else if tenant.EventType == "installed" {
		tenant.AddonInstalled = false
	}
	LOG.Debugf("Created new Tenant instance from reader; tenant: %+v\n", *tenant)
	return tenant, nil
}
