package gonnect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"runtime"
	"text/template"

	"github.com/craftamap/atlas-gonnect/store"
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
}

type Addon struct {
	Config          *Profile
	CurrentProfile  string
	Store           *store.Store
	AddonDescriptor map[string]interface{}
	Key             *string
	Name            *string
	Logger          *logrus.Logger
	Product         string
}

func readAddonDescriptor(descriptorReader io.Reader, baseUrl string, validateDescriptor bool, product string) (map[string]interface{}, error) {
	vals := map[string]string{
		"BaseUrl": baseUrl,
	}

	temp, err := ioutil.ReadAll(descriptorReader)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("descriptor").Parse(string(temp))
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer

	err = tmpl.ExecuteTemplate(&buffer, "descriptor", vals)
	if err != nil {
		return nil, err
	}

	unbuffered := buffer.String()

	if validateDescriptor {
		bool, err, errA := ValidateAppDescriptor(product, unbuffered)
		LOG.Warnf("Validation was %t, %s, %+v", bool, err, errA)
	}

	descriptor := map[string]interface{}{}

	err = json.Unmarshal([]byte(unbuffered), &descriptor)
	if err != nil {
		return nil, err
	}

	return descriptor, nil
}

func NewAddon(configFile io.Reader, descriptorFile io.Reader) (*Addon, error) {
	LOG.Info("Initializing new Addon")

	LOG.Debug("Create new config object")
	config, currentProfile, product, err := NewConfig(configFile)
	if err != nil {
		LOG.Errorf("Could not create new config object: %s\n", err)
		return nil, err
	}

	LOG.Debug("Creating new store")
	store, err := store.New(config.Store.Type, config.Store.DatabaseUrl)
	if err != nil {
		LOG.Errorf("Could not create new store: %s\n", err)
		return nil, err
	}
	LOG.Debug("Reading AddonDescriptor")
	addonDescriptor, err := readAddonDescriptor(descriptorFile, config.BaseUrl, config.ValidateDescriptor, product)
	if err != nil {
		LOG.Errorf("Could not read AddonDescriptor: %s\n", err)
		return nil, err
	}

	name, ok := addonDescriptor["name"].(string)
	if !ok {
		return nil, errors.New("name could not be read from AddonDescriptor")
	}

	key, ok := addonDescriptor["key"].(string)
	if !ok {
		return nil, errors.New("key could not be read from AddonDescriptor")
	}

	addon := &Addon{
		Config:          config,
		Store:           store,
		Logger:          LOG,
		CurrentProfile:  currentProfile,
		AddonDescriptor: addonDescriptor,
		Name:            &name,
		Key:             &key,
		Product:         product,
	}

	LOG.Info("Addon successfully initialized!")
	return addon, nil
}
