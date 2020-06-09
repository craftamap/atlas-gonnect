package gonnect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"runtime"
	"text/template"

	"github.com/sirupsen/logrus"
)

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

type Addon struct {
	Config          *EnvironmentConfiguration
	Store           *Store
	AddonDescriptor map[string]interface{}
	rootFileSystem  *http.FileSystem
	Key             *string
	Name            *string
	Logger          *logrus.Logger
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

	return json.Unmarshal(buffer.Bytes(), &addon.AddonDescriptor)
}


func NewAddon(root *http.FileSystem) (*Addon, error) {
	LOG.Info("Initializing new Addon")

	LOG.Debug("Reading config.json")
	configContent, err := (*root).Open("config.json")
	if err != nil {
		LOG.Errorf("Could not read config: %s\n", err)
		return nil, err
	}

	LOG.Debug("Create new config object")
	config, err := NewConfig(configContent)
	if err != nil {
		LOG.Errorf("Could not create new config object: %s\n", err)
		return nil, err
	}

	LOG.Debug("Creating new store")
	store, err := NewStore(config.Store.Type, config.Store.DatabaseUrl)
	if err != nil {
		LOG.Errorf("Could not create new store: %s\n", err)
		return nil, err
	}

	addon := &Addon{
		Config:         config,
		Store:          store,
		Logger:         LOG,
		rootFileSystem: root,
	}

	LOG.Debug("Reading AddonDescriptor")
	err = addon.readAddonDescriptor()
	if err != nil {
		LOG.Errorf("Could not read AddonDescriptor: %s\n", err)
		return addon, err
	}

	name, ok := addon.AddonDescriptor["name"].(string)
	if !ok {
		return nil, errors.New("name could not be read from AddonDescriptor")
	}
	addon.Name = &name

	key, ok := addon.AddonDescriptor["key"].(string)
	if !ok {
		return nil, errors.New("name could not be read from AddonDescriptor")
	}
	addon.Key = &key

	LOG.Info("Addon successfully initialized!")
	return addon, nil
}
