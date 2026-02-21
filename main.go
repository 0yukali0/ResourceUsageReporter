package nodemonitor

import (
	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()
	for _, r := range monitorRoutes {
		router.Handler(r.Method, r.API, r.HandlerFunc)
	}
}
