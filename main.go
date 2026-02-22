package main

import (
	"fmt"
	"net/http"
	"os"

	"nodeMonitor/pkg/webservice"

	"github.com/julienschmidt/httprouter"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	router := httprouter.New()
	for _, r := range webservice.MonitorRoutes {
		router.Handler(r.Method, r.API, r.HandlerFunc)
	}
	_ = http.ListenAndServe(fmt.Sprintf(":%s", port), router)
}
