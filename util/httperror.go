package util

import (
	"net/http"
	"path/filepath"
	"runtime"

	gonnect "github.com/craftamap/atlas-gonnect"
)

func SendError(w http.ResponseWriter, addon *gonnect.Addon, errorCode int, message string) {
	w.WriteHeader(errorCode)
	w.Write([]byte(message))
	pc, file, no, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)

	if ok {
		addon.Logger.Errorf("%s:%d %s() %s", filepath.Base(file), no, details.Name(), message)
	} else {
		addon.Logger.Error(message)
	}
}
