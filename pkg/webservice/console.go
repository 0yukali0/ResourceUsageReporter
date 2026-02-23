package webservice

import (
	"net/http"
	"os"
)

func ConsoleHandler(w http.ResponseWriter, r *http.Request) {
	os.Getenv("WEB_DIST")
	http.ServeFile(w, r, "./dist/index.html")
}
