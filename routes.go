package gonnect

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Panicln(err)
	}
	log.Printf("%s\n", b)
}

func NewInstalledHandler(addon *Addon) http.Handler {
	return InstalledHandler{addon}
}

func RegisterRoutes(addon *Addon, mux *mux.Router) {
	mux.Handle("/", NewRootHandler())
	mux.Handle("/atlassian-connect.json", NewAtlassianConnectHandler(addon))
	mux.Handle("/installed", handlers.MethodHandler{"POST": NewInstalledHandler(addon)})
}
