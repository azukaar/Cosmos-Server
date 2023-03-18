package main

import (
    "net/http"
		"./utils"
		"./user"
		"./configapi"
		"./proxy"
		"github.com/gorilla/mux"
		"strings"
		"strconv"
		"time"
    "encoding/json"
		"os"
		"github.com/go-chi/chi/middleware"
		"github.com/go-chi/httprate"
		"crypto/tls"
		spa "github.com/roberthodgen/spa-server"
)

var serverPortHTTP = ""
var serverPortHTTPS = ""

func startHTTPServer(router *mux.Router) {
	utils.Log("Listening to HTTP on :" + serverPortHTTP)

	err := http.ListenAndServe("0.0.0.0:" + serverPortHTTP, router)

	if err != nil {
		utils.Fatal("Listening to HTTP", err)
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
				utils.Fatal("Listening to HTTP (Red)", err)
			}
		})()

		utils.Log("Listening to HTTP on :" + serverPortHTTP)
		utils.Log("Listening to HTTPS on :" + serverPortHTTPS)

		utils.IsHTTPS = true

		cert, errCert := tls.X509KeyPair(([]byte)(tlsCert), ([]byte)(tlsKey))
		if errCert != nil {
			utils.Fatal("Getting Certificate pair", errCert)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			// Other options
		}
		
		server := http.Server{
			TLSConfig: tlsConfig,
			Addr: utils.GetMainConfig().HTTPConfig.Hostname + ":" + serverPortHTTPS,
			ReadTimeout: 0,
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout: 0,
			IdleTimeout: 30 * time.Second,
			Handler: router,
			DisableGeneralOptionsHandler: true,
		}

		// start https server
		err := server.ListenAndServeTLS("", "")

		if err != nil {
			utils.Fatal("Listening to HTTPS", err)
		}
}

func tokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("x-cosmos-user", "")
		r.Header.Set("x-cosmos-role", "")

		u, err := user.RefreshUserToken(w, r)

		if err != nil {
			return
		}

		r.Header.Set("x-cosmos-user", u.Nickname)
		r.Header.Set("x-cosmos-role", strconv.Itoa((int)(u.Role)))

		next.ServeHTTP(w, r)
	})
}

func StartServer() {
	baseMainConfig := utils.GetBaseMainConfig()
	config := utils.GetMainConfig().HTTPConfig
	serverPortHTTP = config.HTTPPort
	serverPortHTTPS = config.HTTPSPort

	var tlsCert = config.TLSCert
	var tlsKey= config.TLSKey

	configJson, _ := json.MarshalIndent(config, "", "  ")
	utils.Debug("Configuration" + (string)(configJson))

	if((tlsCert == "" || tlsKey == "") && config.GenerateMissingTLSCert) {
		utils.Log("Generating new TLS certificate")
		pub, priv := utils.GenerateRSAWebCertificates()
		
		baseMainConfig.HTTPConfig.TLSCert = pub
		baseMainConfig.HTTPConfig.TLSKey = priv
		utils.SetBaseMainConfig(baseMainConfig)

		utils.Log("Saved new TLS certificate")

		tlsCert = pub
		tlsKey = priv
	}

	if ((config.AuthPublicKey == "" || config.AuthPrivateKey == "") && config.GenerateMissingAuthCert) {
		utils.Log("Generating new Auth ED25519 certificate")
		pub, priv := utils.GenerateEd25519Certificates()
		
		baseMainConfig.HTTPConfig.AuthPublicKey = pub
		baseMainConfig.HTTPConfig.AuthPrivateKey = priv
		utils.SetBaseMainConfig(baseMainConfig)

		utils.Log("Saved new Auth ED25519 certificate")
	}

	router := mux.NewRouter().StrictSlash(true)

	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(utils.SetSecurityHeaders)

	router.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/ui", http.StatusMovedPermanently)
	}))
	
	srapi := router.PathPrefix("/cosmos").Subrouter()

	srapi.HandleFunc("/api/login", user.UserLogin)
	srapi.HandleFunc("/api/logout", user.UserLogout)
	srapi.HandleFunc("/api/register", user.UserRegister)
	srapi.HandleFunc("/api/invite", user.UserResendInviteLink)
	srapi.HandleFunc("/api/me", user.Me)
	srapi.HandleFunc("/api/config", configapi.ConfigRoute)
	srapi.HandleFunc("/api/restart", configapi.ConfigApiRestart)

	srapi.HandleFunc("/api/users/{nickname}", user.UsersIdRoute)
	srapi.HandleFunc("/api/users", user.UsersRoute)

	// srapi.Use(utils.AcceptHeader("*/*"))
	srapi.Use(tokenMiddleware)
	srapi.Use(utils.CORSHeader(utils.GetMainConfig().HTTPConfig.Hostname))
	srapi.Use(utils.MiddlewareTimeout(20 * time.Second))
	srapi.Use(httprate.Limit(60, 1*time.Minute, 
		httprate.WithKeyFuncs(httprate.KeyByIP),
    httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			utils.Error("Too many requests. Throttling", nil)
			utils.HTTPError(w, "Too many requests", 
				http.StatusTooManyRequests, "HTTP003")
			return 
		}),
	))
	
	pwd,_ := os.Getwd()
	utils.Log("Starting in " + pwd)
	if _, err := os.Stat(pwd + "/static"); os.IsNotExist(err) {
		utils.Fatal("Static folder not found at " + pwd + "/static", err)
	}

	fs  := spa.SpaHandler(pwd + "/static", "index.html")
	router.PathPrefix("/ui").Handler(http.StripPrefix("/ui", fs))

	router = proxy.BuildFromConfig(router, config.ProxyConfig)

	if tlsCert != "" && tlsKey != "" {
		utils.Log("TLS certificate exist, starting HTTPS servers and redirecting HTTP to HTTPS")
		startHTTPSServer(router, tlsCert, tlsKey)
	} else {
		utils.Log("TLS certificate does not exist, starting HTTP server only")
		startHTTPServer(router)
	}
}
