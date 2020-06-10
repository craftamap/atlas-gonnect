package routes

import (
	"encoding/json"
	"net/http"

	"github.com/craftamap/atlas-gonnect/middleware"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	gonnect "github.com/craftamap/atlas-gonnect"
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
	Addon *gonnect.Addon
}

func (h AtlassianConnectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.Addon.AddonDescriptor)
}

func NewAtlassianConnectHandler(addon *gonnect.Addon) http.Handler {
	return AtlassianConnectHandler{addon}
}

type InstalledHandler struct {
	Addon *gonnect.Addon
}

func (h InstalledHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenant, err := gonnect.NewTenantFromReader(r.Body)
	if err != nil {
		w.WriteHeader(501) //TODO: figure out right server error response codes
		return
	}
	_, err = h.Addon.Store.Set(tenant)
	if err != nil {
		w.WriteHeader(501) //TODO: figure out right server error response codes
		return
	}
	h.Addon.Logger.Infof("installed new tenant %s\n", tenant.BaseURL)
	//TODO: Figure out what to response - like with my girlfriend <3
	w.Write([]byte("OK"))
}

func NewInstalledHandler(addon *gonnect.Addon) http.Handler {
	return InstalledHandler{addon}
}

type UninstalledHandler struct {
	Addon *gonnect.Addon
}

func (h UninstalledHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenant, err := gonnect.NewTenantFromReader(r.Body)
	if err != nil {
		w.WriteHeader(501) //TODO: figure out right server error response codes
		return
	}
	_, err = h.Addon.Store.Set(tenant)
	if err != nil {
		w.WriteHeader(501) //TODO: figure out right server error response codes
		return
	}
	h.Addon.Logger.Infof("uninstalled tenant %s\n", tenant.BaseURL)
	//TODO: Figure out what to response
	w.Write([]byte("OK"))
}

func NewUninstalledHandler(addon *gonnect.Addon) http.Handler {
	return InstalledHandler{addon}
}

func RegisterRoutes(addon *gonnect.Addon, mux *mux.Router) {
	mux.Handle("/", NewRootHandler())
	mux.Handle("/atlassian-connect.json", NewAtlassianConnectHandler(addon))
	mux.Handle("/installed", handlers.MethodHandler{"POST": middleware.NewVerifyInstallationMiddleware(addon)(NewInstalledHandler(addon))})
	mux.Handle("/uninstalled", handlers.MethodHandler{"POST": NewUninstalledHandler(addon)})
}
