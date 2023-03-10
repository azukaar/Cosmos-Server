package proxy

import (
	"github.com/gorilla/mux"
	"net/http"
	"../utils"
)

func BuildFromConfig(config utils.ProxyConfig) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/_health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	router.PathPrefix("/_static").Handler(http.StripPrefix("/_static", http.FileServer(http.Dir("static"))))
	
	for i := len(config.Routes)-1; i >= 0; i-- {
		routeConfig := config.Routes[i]
		RouterGen(routeConfig.Routing, router, RouteTo(routeConfig.Target))
	}
	
	return router
}