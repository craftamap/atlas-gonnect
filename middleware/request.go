package middleware

import (
	"net/http"
	"net/url"

	gonnect "github.com/craftamap/atlas-gonnect"
	"github.com/gorilla/context"
)

type RequestMiddleware struct {
	h              http.Handler
	addon          *gonnect.Addon
	verifiedParams map[string]string
}

func (h RequestMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	getParam := func(key string) string {
		value := r.URL.Query().Get(key)
		if value != "" {
			err := r.ParseForm()
			if err != nil {
				return ""
			}
			v := r.Form
			value = v.Get(key)
		}
		return value
	}

	getHostBaseUrlFromQueryParams := func() string {
		hostUrl := getParam("xdm_e")
		if hostUrl != "" {
			return hostUrl + getParam("cp")
		} else {
			return ""
		}

	}

	getHostResourceUrl := func(isDev bool, baseUrl string, ext string) *url.URL {
		// again, bb handling is missing here
		var resource string
		if isDev {
			resource = "all-debug." + ext
		} else {
			resource = "all." + ext
		}

		uri, err := url.Parse(baseUrl + "/atlassian-connect/" + resource)
		if err != nil {
			return &url.URL{}
		} else {
			return uri
		}
	}

	h.addon.Logger.Debug("Setting Context Variables in Request Middleware")
	//TODO: Better Logging in this middleware
	context.Set(r, "title", *h.addon.Name)
	context.Set(r, "addonKey", *h.addon.Key)
	context.Set(r, "license", getParam("lic"))
	context.Set(r, "localBaseUrl", h.addon.Config.BaseUrl)

	// if missing here: if isJira || isConfluence
	// Since this poc is for confluence only, this should be valid, for now
	context.Set(r, "hostBaseUrl", getHostBaseUrlFromQueryParams())

	if(len(h.verifiedParams) > 0){
		// TODO: verifiedParams Logic
	}

	context.Set(r, "baseUrl", getHostBaseUrlFromQueryParams())
	context.Set(r, "hostStylesheetUrl",
		getHostResourceUrl(true, context.Get(r, "hostBaseUrl").(string), "css"))
	context.Set(r, "hostScriptUrl", "https://connect-cdn.atl-paas.net/all.js")

	h.h.ServeHTTP(w, r)
}

func NewRequestMiddleware(addon *gonnect.Addon, verifiedParameters map[string]string) func(h http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return RequestMiddleware{handler, addon, verifiedParameters}
	}
}
