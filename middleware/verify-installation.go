package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	gonnect "github.com/craftamap/atlas-gonnect"
	"github.com/gorilla/context"
)

type VerifyInstallationMiddleware struct {
	h     http.Handler
	addon *gonnect.Addon
}

func (h VerifyInstallationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Body == http.NoBody {
		w.WriteHeader(401)
		h.addon.Logger.Warn("No registration info provided")
		return
	}

	b := bytes.NewBuffer(make([]byte, 0))
	reader := io.TeeReader(r.Body, b)
	defer r.Body.Close()

	responseData := map[string]interface{}{}
	json.NewDecoder(reader).Decode(&responseData)

	r.Body = ioutil.NopCloser(b)

	// TODO: Add whitelist feature

	baseUrl, ok := responseData["baseUrl"]
	if !ok {
		w.WriteHeader(401)
		h.addon.Logger.Warn("No baseUrl provided for registration info")
		return
	}

	clientKey, ok := responseData["clientKey"]
	if !ok {
		w.WriteHeader(401)
		h.addon.Logger.Warnf("No clientKey provided for host %s", baseUrl)
		return
	}

	_, err := h.addon.Store.Get(clientKey.(string))
	if err != nil {
		// If err is set here, we serve the normal installation
		h.h.ServeHTTP(w, r)
	} else {
		authHandler := NewAuthenticationMiddleware(h.addon)
		authHandler(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			if context.Get(req, "clientKey") == clientKey {
				h.h.ServeHTTP(writer, req)
			} else {
				writer.WriteHeader(401)
				h.addon.Logger.Warn("clientKey in install payload did not match authenticated client")
				
			}
		})).ServeHTTP(w, r)
	}
}

func NewVerifyInstallationMiddleware(addon *gonnect.Addon) func(h http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return VerifyInstallationMiddleware{handler, addon}
	}
}
