package proxy

import (
	"net/http"
	"net/http/httputil"  
	"github.com/gorilla/mux"
	// "log"
	// "io/ioutil"
	// "io"
	// "os"
	// "golang.org/x/crypto/bcrypt"

	// "../utils" 
)

type Route struct {
	UseHost bool
  Host string
	UsePathPrefix bool
	PathPrefix string
}

func RouterGen(route Route, router *mux.Router, destination *httputil.ReverseProxy) *mux.Route {
	var realDestination http.Handler
	realDestination = destination

	origin := router.Methods("GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD")

	if(route.UseHost) {
		origin = origin.Host(route.Host)
	}

	if(route.UsePathPrefix) {
		origin = origin.PathPrefix(route.PathPrefix)
		realDestination = http.StripPrefix(route.PathPrefix, destination)
	}

	origin.Handler(realDestination)

	return origin
}