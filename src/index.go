package main

import (
    "net/http"
		"./utils"
		"./file"
		"./user"
		"github.com/gorilla/mux"
		"log"
		"os"
		"strings"
)

var tlsCert = "localcert.crt"
var tlsKey= "localcert.key"
var serverPortHTTP = os.Getenv("HTTP_PORT") 
var serverPortHTTPS = os.Getenv("HTTPS_PORT")

func startHTTPServer(router *mux.Router) {
	log.Println("Listening to HTTP on :" + serverPortHTTP)

	err := http.ListenAndServe("0.0.0.0:" + serverPortHTTP, router)

	if err != nil {
		log.Fatal(err)
	}
}

func startHTTPSServer(router *mux.Router) {
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

func main() {
	  log.Println("Starting...")

		if serverPortHTTP == "" {
			serverPortHTTP = "80"
		}

		if serverPortHTTPS == "" {
			serverPortHTTPS = "443"
		}

		router := mux.NewRouter().StrictSlash(true)

		utils.DB()
		defer utils.Disconnect()

		router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    
		router.HandleFunc("/file/list", file.FileList)
    router.HandleFunc("/file/get", file.FileGet)
		router.HandleFunc("/file/delete", file.FileDelete)
		router.HandleFunc("/file/copy", file.FileCopy)
		router.HandleFunc("/file/move", file.FileMove)
		
		router.HandleFunc("/user/login", user.UserLogin)
		// router.HandleFunc("/user/register", user.UserRegister)
		// router.HandleFunc("/user/edit", )
		// router.HandleFunc("/user/delete", )
		
		// router.HandleFunc("/config/get", )
		// router.HandleFunc("/config/set", )
	
		// router.HandleFunc("/db", )
		
		if utils.FileExists(tlsCert) && utils.FileExists(tlsKey) {
			log.Println("TLS certificate found, starting HTTPS servers and redirecting HTTP to HTTPS")
			startHTTPSServer(router)
		} else {
			log.Println("No TLS certificate found, starting HTTP server only")
			startHTTPServer(router)
		}
}