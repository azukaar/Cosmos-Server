package main

import (
    "net/http"
		"./utils"
		// "./file"
		// "./user"
		"./proxy"
		"github.com/gorilla/mux"
		"log"
		"os"
		"strings"
)

type HTTPConfig struct {
	TLSCert string
	TLSKey string
	GenerateMissingTLSCert bool
	HTTPPort string
	HTTPSPort string
	ProxyConfig proxy.Config
} 

var serverPortHTTP = os.Getenv("HTTP_PORT") 
var serverPortHTTPS = os.Getenv("HTTPS_PORT")

func startHTTPServer(router *mux.Router) {
	log.Println("Listening to HTTP on :" + serverPortHTTP)

	err := http.ListenAndServe("0.0.0.0:" + serverPortHTTP, router)

	if err != nil {
		log.Fatal(err)
	}
}

func startHTTPSServer(router *mux.Router, tlsCert string, tlsKey string) {
		// redirect http to https
		go (func () {
			err := http.ListenAndServe("0.0.0.0:" + serverPortHTTP, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// change port in host
				if strings.HasSuffix(r.Host, ":" + serverPortHTTP) {
					if serverPortHTTPS != "443" {
						r.Host = r.Host[:len(r.Host)-len(":" + serverPortHTTP)] + ":" + serverPortHTTPS
						} else {
						r.Host = r.Host[:len(r.Host)-len(":" + serverPortHTTP)]
					}
				}
				
				http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
			}))
			if err != nil {
				log.Fatal(err)
			}
		})()

		log.Println("Listening to HTTP on :" + serverPortHTTP)
		log.Println("Listening to HTTPS on :" + serverPortHTTPS)

		// start https server
		err := http.ListenAndServeTLS("0.0.0.0:" + serverPortHTTPS, tlsCert, tlsKey, router)

		if err != nil {
			log.Fatal(err)
		}
}


func StartServer(config HTTPConfig) {
	var tlsCert = config.TLSCert
	var tlsKey= config.TLSKey

	if serverPortHTTP == "" {
		serverPortHTTP = config.HTTPPort
	}

	if serverPortHTTPS == "" {
		serverPortHTTPS = config.HTTPSPort
	}

	router := proxy.BuildFromConfig(config.ProxyConfig)

	if utils.FileExists(tlsCert) && utils.FileExists(tlsKey) {
		log.Println("TLS certificate found, starting HTTPS servers and redirecting HTTP to HTTPS")
		startHTTPSServer(router, tlsCert, tlsKey)
	} else {
		log.Println("No TLS certificate found, starting HTTP server only")
		startHTTPServer(router)
	}
}

func StopServer() {
}