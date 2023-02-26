package proxy

import (
	"net/http"
	"net/http/httputil"    
	"net/url"
	"log"
	// "io/ioutil"
	// "io"
	// "os"
	// "golang.org/x/crypto/bcrypt"

	// "../utils" 
)

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
			return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	// upgrade the request to websocket
	proxy.ModifyResponse = func(resp *http.Response) error {
		log.Println("[INFO] Response from backend: ", resp.Status)
		log.Println("[INFO] URL was ", resp.Request.URL)
		return nil
	}

	return proxy, nil
}

// ProxyRequestHandler handles the http request using proxy
func ProxyRequestHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
	}
}

func RouteTo(destination string) *httputil.ReverseProxy /*func(http.ResponseWriter, *http.Request)*/ {
	// initialize a reverse proxy and pass the actual backend server url here
	proxy, err := NewProxy(destination)
	if err != nil {
			panic(err)
	}

	// create a handler function which uses the reverse proxy
	return proxy //ProxyRequestHandler(proxy)
}