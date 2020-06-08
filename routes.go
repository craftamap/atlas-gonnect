package gonnect

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type RootHandler struct {
}

func (h RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/atlassian-connect.json", http.StatusPermanentRedirect)
}

func NewRootHandler() http.Handler {
	return RootHandler{}
}

type AtlassianConnectHandler struct {
	Addon *Addon
}

func (h AtlassianConnectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.Addon.AddonDescriptor)
}

func NewAtlassianConnectHandler(addon *Addon) http.Handler {
	return AtlassianConnectHandler{addon}
}

type InstalledHandler struct {
	Addon *Addon
}

func (h InstalledHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenant, err := NewTenantFromReader(r.Body)
	if err != nil {
		w.WriteHeader(501) //TODO: figure out right server error response codes
		return
	}
	_, err = h.Addon.Store.set(tenant)
	if err != nil {
		w.WriteHeader(501) //TODO: figure out right server error response codes
		return
	}
	LOG.Infof("installed new tenant %s\n", tenant.BaseURL)
	//TODO: Figure out what to response - like with my girlfriend <3
	w.Write([]byte("OK"))
}

func NewInstalledHandler(addon *Addon) http.Handler {
	return InstalledHandler{addon}
}

type UninstalledHandler struct {
	Addon *Addon
}

func (h UninstalledHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenant, err := NewTenantFromReader(r.Body)
	if err != nil {
		w.WriteHeader(501) //TODO: figure out right server error response codes
		return
	}
	_, err = h.Addon.Store.set(tenant)
	if err != nil {
		w.WriteHeader(501) //TODO: figure out right server error response codes
		return
	}
	LOG.Infof("uninstalled tenant %s\n", tenant.BaseURL)
	//TODO: Figure out what to response
	w.Write([]byte("OK"))
}

func NewUninstalledHandler(addon *Addon) http.Handler {
	return InstalledHandler{addon}
}

func RegisterRoutes(addon *Addon, mux *mux.Router) {
	mux.Handle("/", NewRootHandler())
	mux.Handle("/atlassian-connect.json", NewAtlassianConnectHandler(addon))
	mux.Handle("/installed", handlers.MethodHandler{"POST": NewInstalledHandler(addon)})
	mux.Handle("/uninstalled", handlers.MethodHandler{"POST": NewUninstalledHandler(addon)})
}
