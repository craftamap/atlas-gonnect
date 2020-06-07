package gonnect

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"text/template"
)

type Addon struct {
	Config          *EnvironmentConfiguration
	Store           *Store
	AddonDescriptor string
	rootFileSystem	*http.FileSystem
}

func (addon *Addon) readAddonDescriptor() (err error) {
	vals := map[string]string{
		"BaseUrl": addon.Config.BaseUrl,
	}

	content, err := (*addon.rootFileSystem).Open("atlassian-connect.json")
	if err != nil {
		return err
	}

	temp, err := ioutil.ReadAll(content)
	if err != nil {
		return err
	}

	tmpl, err := template.New("descriptor").Parse(string(temp))
	if err != nil {
		return err
	}

	var buffer bytes.Buffer

	err = tmpl.ExecuteTemplate(&buffer, "descriptor", vals)
	if err != nil {
		return err
	}

	addon.AddonDescriptor = buffer.String()

	return nil
}

func NewAddon(root *http.FileSystem) (*Addon, error) {
	configContent, err := (*root).Open("config.json")
	if err != nil {
		return nil, err
	}

	config, err := NewConfig(configContent)
	if err != nil {
		return nil, err
	}

	store, err := NewStore(config.Store.Type, config.Store.DatabaseUrl)
	if err != nil {
		return nil, err
	}

	addon := &Addon{
		Config: config,
		Store:  store,
		rootFileSystem: root,
	}

	err = addon.readAddonDescriptor()
	if err != nil {
		return addon, err
	}

	return addon, nil
}
