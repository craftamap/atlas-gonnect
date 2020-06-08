package gonnect

import "time"

type Tenant struct {
	ClientKey      string `json:"clientKey" gorm:"type:varchar(255);primary_key"`
	PublicKey      string `json:"publicKey" gorm:"type:varchar(512)"`
	SharedSecret   string `json:"sharedSecret" gorm:"type:varchar(255);NOT NULL"`
	BaseURL        string `json:"baseUrl" gorm:"type:varchar(255);NOT NULL"`
	ProductType    string `json:"productType" gorm:"type:varchar(255)"`
	Description    string `json:"description" gorm:"type:varchar(255)"`
	AddonInstalled bool   `json:"-" gorm:"type:bit;NOT NULL"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewTenantFromMap(body map[string]interface{}) *Tenant {
	tenant := &Tenant{}
	//TODO: How to do this nice?
	clientKey, ok := body["clientKey"].(string)
	if ok {
		tenant.ClientKey = clientKey
	}
	publicKey, ok := body["publicKey"].(string)
	if ok {
		tenant.PublicKey = publicKey
	}
	sharedSecret, ok := body["sharedSecret"].(string)
	if ok {
		tenant.SharedSecret = sharedSecret
	}
	baseURL, ok := body["baseUrl"].(string)
	if ok {
		tenant.BaseURL = baseURL
	}
	productType, ok := body["productType"].(string)
	if ok {
		tenant.ProductType = productType
	}
	description, ok := body["description"].(string)
	if ok {
		tenant.Description = description
	}
	if body["eventType"].(string) == "installed" {
		tenant.AddonInstalled = true
	} else if body["eventType"].(string) == "installed" {
		tenant.AddonInstalled = false
	}

	return tenant
}
