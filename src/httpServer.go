package main

import (
    "net/http"
		"github.com/azukaar/cosmos-server/src/utils"
		"github.com/azukaar/cosmos-server/src/user"
		"github.com/azukaar/cosmos-server/src/configapi"
		"github.com/azukaar/cosmos-server/src/proxy"
		"github.com/azukaar/cosmos-server/src/docker"
		"github.com/gorilla/mux"
		"strconv"
		"time"
		"os"
		"strings"
		"github.com/go-chi/chi/middleware"
		"github.com/go-chi/httprate"
		"crypto/tls"
		spa "github.com/roberthodgen/spa-server"
		"github.com/foomo/simplecert"
		"github.com/foomo/tlsconfig"
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
	config  := utils.GetMainConfig()
	serverHostname := "0.0.0.0"

	cfg := simplecert.Default

	cfg.Domains = utils.GetAllHostnames()
	cfg.CacheDir = "/config/certificates"
	cfg.SSLEmail = config.HTTPConfig.SSLEmail
	cfg.HTTPAddress = serverHostname+":"+serverPortHTTP
	cfg.TLSAddress = serverHostname+":"+serverPortHTTPS
	
	if config.HTTPConfig.DNSChallengeProvider != "" {
		cfg.DNSProvider  = config.HTTPConfig.DNSChallengeProvider
	}

	cfg.FailedToRenewCertificate = func(err error) {
		utils.Error("Failed to renew certificate", err)
	}

	var certReloader *simplecert.CertReloader 
	var errSimCert error
	if(config.HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["LETSENCRYPT"]) {
		certReloader, errSimCert = simplecert.Init(cfg, nil)
		if errSimCert != nil {
			  // Temporary before we have a better way to handle this
				utils.Error("simplecert init failed, HTTPS wont renew", errSimCert)
		}
	}
		
	// redirect http to https
	go (func () {
		// err := http.ListenAndServe("0.0.0.0:" + serverPortHTTP, http.HandlerFunc(simplecert.Redirect))
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
			utils.Fatal("Listening to HTTP (Redirecting to HTTPS)", err)
		}
	})()

	utils.Log("Listening to HTTP on :" + serverPortHTTP)
	utils.Log("Listening to HTTPS on :" + serverPortHTTPS)

	utils.IsHTTPS = true

	tlsConf := tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)

	if(config.HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["LETSENCRYPT"]) {
		tlsConf.GetCertificate = certReloader.GetCertificateFunc()
	} else {
		cert, errCert := tls.X509KeyPair(([]byte)(tlsCert), ([]byte)(tlsKey))
		if errCert != nil {
			utils.Fatal("Getting Certificate pair", errCert)
		}

		tlsConf.Certificates = []tls.Certificate{cert}
	}
	
	server := http.Server{
		TLSConfig: tlsConf,
		Addr: serverHostname + ":" + serverPortHTTPS,
		ReadTimeout: 0,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout: 0,
		IdleTimeout: 30 * time.Second,
		Handler: router,
		DisableGeneralOptionsHandler: true,
	}

	// start https server
	errServ := server.ListenAndServeTLS("", "")

	if errServ != nil {
		utils.Fatal("Listening to HTTPS", errServ)
	}
}

func tokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Header.Del
		r.Header.Del("x-cosmos-user")
		r.Header.Del("x-cosmos-role")
		r.Header.Del("x-cosmos-mfa")

		u, err := user.RefreshUserToken(w, r)

		if err != nil {
			return
		}

		r.Header.Set("x-cosmos-user", u.Nickname)
		r.Header.Set("x-cosmos-role", strconv.Itoa((int)(u.Role)))
		r.Header.Set("x-cosmos-mfa", strconv.Itoa((int)(u.MFAState)))

		next.ServeHTTP(w, r)
	})
}

func StartServer() {
	baseMainConfig := utils.GetBaseMainConfig()
	config := utils.GetMainConfig()
	HTTPConfig := config.HTTPConfig
	serverPortHTTP = HTTPConfig.HTTPPort
	serverPortHTTPS = HTTPConfig.HTTPSPort

	var tlsCert = HTTPConfig.TLSCert
	var tlsKey= HTTPConfig.TLSKey

	domains := utils.GetAllHostnames()
	oldDomains := baseMainConfig.HTTPConfig.TLSKeyHostsCached

	NeedsRefresh := (tlsCert == "" || tlsKey == "") || !utils.StringArrayEquals(domains, oldDomains)

	if(NeedsRefresh && HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["SELFSIGNED"]) {
		utils.Log("Generating new TLS certificate")
		pub, priv := utils.GenerateRSAWebCertificates(domains)
		
		baseMainConfig.HTTPConfig.TLSCert = pub
		baseMainConfig.HTTPConfig.TLSKey = priv
		baseMainConfig.HTTPConfig.TLSKeyHostsCached = domains

		utils.SetBaseMainConfig(baseMainConfig)

		utils.Log("Saved new TLS certificate")

		tlsCert = pub
		tlsKey = priv
	}

	if ((HTTPConfig.AuthPublicKey == "" || HTTPConfig.AuthPrivateKey == "") && HTTPConfig.GenerateMissingAuthCert) {
		utils.Log("Generating new Auth ED25519 certificate")
		pub, priv := utils.GenerateEd25519Certificates()
		
		baseMainConfig.HTTPConfig.AuthPublicKey = pub
		baseMainConfig.HTTPConfig.AuthPrivateKey = priv
		utils.SetBaseMainConfig(baseMainConfig)

		utils.Log("Saved new Auth ED25519 certificate")
	}

	router := mux.NewRouter().StrictSlash(true)

	// need rewrite bc it catches too many things and prevent
	// client to be notified of the error
	// router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(utils.SetSecurityHeaders)

	if config.BlockedCountries != nil && len(config.BlockedCountries) > 0 {
		router.Use(utils.BlockByCountryMiddleware(config.BlockedCountries))
	}
	
	srapi := router.PathPrefix("/cosmos").Subrouter()

	srapi.HandleFunc("/api/status", StatusRoute)
	srapi.HandleFunc("/api/favicon", GetFavicon)
	srapi.HandleFunc("/api/ping", PingURL)
	srapi.HandleFunc("/api/newInstall", NewInstallRoute)
	srapi.HandleFunc("/api/login", user.UserLogin)
	srapi.HandleFunc("/api/logout", user.UserLogout)
	srapi.HandleFunc("/api/register", user.UserRegister)
	srapi.HandleFunc("/api/invite", user.UserResendInviteLink)
	srapi.HandleFunc("/api/me", user.Me)
	srapi.HandleFunc("/api/mfa", user.API2FA)
	srapi.HandleFunc("/api/password-reset", user.ResetPassword)
	srapi.HandleFunc("/api/config", configapi.ConfigRoute)
	srapi.HandleFunc("/api/restart", configapi.ConfigApiRestart)

	srapi.HandleFunc("/api/users/{nickname}", user.UsersIdRoute)
	srapi.HandleFunc("/api/users", user.UsersRoute)
	
	srapi.HandleFunc("/api/servapps/{containerId}/secure/{status}", docker.SecureContainerRoute)
	srapi.HandleFunc("/api/servapps", docker.ContainersRoute)

	srapi.Use(tokenMiddleware)
	srapi.Use(proxy.SmartShieldMiddleware(
		utils.SmartShieldPolicy{
			Enabled: true,
			PerUserSimultaneous: 3,
			MaxGlobalSimultaneous: 12,
			PolicyStrictness: 1,
			PerUserRequestLimit: 5000,
		},
	))
	srapi.Use(utils.MiddlewareTimeout(20 * time.Second))
	srapi.Use(utils.BlockPostWithoutReferer)
	srapi.Use(proxy.BotDetectionMiddleware)
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

	router = proxy.BuildFromConfig(router, HTTPConfig.ProxyConfig)
	
	router.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/ui", http.StatusMovedPermanently)
	}))

	if ((HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["SELFSIGNED"] || HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["PROVIDED"]) &&
			 tlsCert != "" && tlsKey != "") || (HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["LETSENCRYPT"]) {
		utils.Log("TLS certificate exist, starting HTTPS servers and redirecting HTTP to HTTPS")
		startHTTPSServer(router, tlsCert, tlsKey)
	} else {
		utils.Log("TLS certificates do not exists or are disabled, starting HTTP server only")
		startHTTPServer(router)
	}
}
