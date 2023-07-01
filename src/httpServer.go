package main

import (
    "net/http"
		"github.com/azukaar/cosmos-server/src/utils"
		"github.com/azukaar/cosmos-server/src/user"
		"github.com/azukaar/cosmos-server/src/configapi"
		"github.com/azukaar/cosmos-server/src/proxy"
		"github.com/azukaar/cosmos-server/src/docker"
		"github.com/azukaar/cosmos-server/src/authorizationserver"
		"github.com/azukaar/cosmos-server/src/market"
		"github.com/gorilla/mux"
		"strconv"
		"time"
		"os"
		"strings"
		"github.com/go-chi/chi/middleware"
		"github.com/go-chi/httprate"
		"crypto/tls"
		spa "github.com/roberthodgen/spa-server"
		"github.com/foomo/tlsconfig"
		"context"
)

var serverPortHTTP = ""
var serverPortHTTPS = ""
var HTTPServer *http.Server

func startHTTPServer(router *mux.Router) {
	var errServ error
	for(errServ == http.ErrServerClosed || errServ == nil) {
		HTTPServer = &http.Server{
			Addr: "0.0.0.0:" + serverPortHTTP,
			ReadTimeout: 0,
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout: 0,
			IdleTimeout: 30 * time.Second,
			Handler: router,
			DisableGeneralOptionsHandler: true,
		}

		utils.Log("Listening to HTTP on : 0.0.0.0:" + serverPortHTTP)
	
		docker.CheckPorts()
	
		errServ = HTTPServer.ListenAndServe()

		if errServ != nil && errServ != http.ErrServerClosed {
			utils.Fatal("Listening to HTTPS", errServ)
		}

		utils.Log("HTTP Server closed. Restarting.")
		errServ = nil

		router = InitServer()
	}
}

func startHTTPSServer(router *mux.Router, tlsCert string, tlsKey string) {
	//config  := utils.GetMainConfig()

	utils.IsHTTPS = true
		
	// redirect http to https
	go (func () {
		httpRouter := mux.NewRouter()

		httpRouter.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// change port in host
			if strings.HasSuffix(r.Host, ":" + serverPortHTTP) {
				if serverPortHTTPS != "443" {
					r.Host = r.Host[:len(r.Host)-len(":" + serverPortHTTP)] + ":" + serverPortHTTPS
					} else {
					r.Host = r.Host[:len(r.Host)-len(":" + serverPortHTTP)]
				}
			}
			
			http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
		})

		err := http.ListenAndServe("0.0.0.0:" + serverPortHTTP, httpRouter)
		
		if err != nil {
			utils.Fatal("Listening to HTTP (Redirecting to HTTPS)", err)
		}
	})()

	utils.Log("Listening to HTTP on :" + serverPortHTTP)
	utils.Log("Listening to HTTPS on :" + serverPortHTTPS)

	tlsConf := tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)

	/*if(config.HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["LETSENCRYPT"]) {
		tlsConf.GetCertificate = certReloader.GetCertificateFunc()
	} else {*/
		cert, errCert := tls.X509KeyPair(([]byte)(tlsCert), ([]byte)(tlsKey))
		if errCert != nil {
			utils.Fatal("Getting Certificate pair", errCert)
		}

		tlsConf.Certificates = []tls.Certificate{cert}
	//}

	// start https server
	var errServ error

	for(errServ == http.ErrServerClosed || errServ == nil) {
		HTTPServer = &http.Server{
			TLSConfig: tlsConf,
			Addr: "0.0.0.0:" + serverPortHTTPS,
			ReadTimeout: 0,
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout: 0,
			IdleTimeout: 30 * time.Second,
			Handler: router,
			DisableGeneralOptionsHandler: true,
		}

		// Redirect ports 
		docker.CheckPorts()

		utils.Log("Now listening to HTTPS on :" + serverPortHTTPS)

		errServ = HTTPServer.ListenAndServeTLS("", "")

		if errServ != nil && errServ != http.ErrServerClosed {
			utils.Fatal("Listening to HTTPS", errServ)
		}

		utils.Log("HTTPS Server closed. Restarting.")
		errServ = nil

		router = InitServer()
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

func SecureAPI(userRouter *mux.Router, public bool) {
	if(!public) {
		userRouter.Use(tokenMiddleware)
	}
	userRouter.Use(proxy.SmartShieldMiddleware(
		utils.SmartShieldPolicy{
			Enabled: true,
			PolicyStrictness: 1,
			PerUserRequestLimit: 5000,
		},
	))
	userRouter.Use(utils.MiddlewareTimeout(45 * time.Second))
	userRouter.Use(utils.BlockPostWithoutReferer)
	userRouter.Use(proxy.BotDetectionMiddleware)
	userRouter.Use(httprate.Limit(120, 1*time.Minute, 
		httprate.WithKeyFuncs(httprate.KeyByIP),
    httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			utils.Error("Too many requests. Throttling", nil)
			utils.HTTPError(w, "Too many requests", 
				http.StatusTooManyRequests, "HTTP003")
			return 
		}),
	))
}

func CertificateIsValid(validUntil time.Time) bool {
	// allow 5 days of leeway
	isValid := time.Now().Add(5 * 24 * time.Hour).Before(validUntil)
	if !isValid {
		utils.Log("Certificate is not valid anymore. Needs refresh")
	}
	return isValid
}

func InitServer() *mux.Router {
	utils.RestartHTTPServer = RestartServer

	baseMainConfig := utils.GetBaseMainConfig()
	config := utils.GetMainConfig()
	HTTPConfig := config.HTTPConfig
	serverPortHTTP = HTTPConfig.HTTPPort
	serverPortHTTPS = HTTPConfig.HTTPSPort

	var tlsCert = HTTPConfig.TLSCert
	var tlsKey= HTTPConfig.TLSKey

	domains := utils.GetAllHostnames(true, false)
	oldDomains := baseMainConfig.HTTPConfig.TLSKeyHostsCached
	falledBack := false

	NeedsRefresh := baseMainConfig.HTTPConfig.ForceHTTPSCertificateRenewal || (tlsCert == "" || tlsKey == "") || !utils.StringArrayEquals(domains, oldDomains) || !CertificateIsValid(baseMainConfig.HTTPConfig.TLSValidUntil)
	
	// If we have a certificate, we can fallback to it if necessary
	CanFallback := tlsCert != "" && tlsKey != "" && 
		len(config.HTTPConfig.TLSKeyHostsCached) > 0 && 
		config.HTTPConfig.TLSKeyHostsCached[0] == config.HTTPConfig.Hostname  &&
		CertificateIsValid(baseMainConfig.HTTPConfig.TLSValidUntil)

	if(NeedsRefresh && config.HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["LETSENCRYPT"]) {
		if(config.HTTPConfig.DNSChallengeProvider != "") {
			newEnv := config.HTTPConfig.DNSChallengeConfig
			for key, value := range newEnv {
				os.Setenv(key, value)
			}
		}

		// Get Certificates 
		pub, priv := utils.DoLetsEncrypt()

		if(pub == "" || priv == "") {
			if(!CanFallback) {
				utils.Error("Getting TLS certificate. Fallback to SELFSIGNED certificates", nil)
				HTTPConfig.HTTPSCertificateMode = utils.HTTPSCertModeList["SELFSIGNED"]
				falledBack = true
			} else {
				utils.Error("Getting TLS certificate. Fallback to previous certificate", nil)
			}
		} else {
			baseMainConfig.HTTPConfig.TLSCert = pub
			baseMainConfig.HTTPConfig.TLSKey = priv
			baseMainConfig.HTTPConfig.TLSKeyHostsCached = domains
			baseMainConfig.HTTPConfig.TLSValidUntil = time.Now().AddDate(0, 0, 90)
			baseMainConfig.HTTPConfig.ForceHTTPSCertificateRenewal = false
	
			utils.SetBaseMainConfig(baseMainConfig)

			utils.Log("Saved new LETSENCRYPT TLS certificate")
	
			tlsCert = pub
			tlsKey = priv
		}
	}
	
	if(NeedsRefresh && HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["SELFSIGNED"]) {
		utils.Log("Generating new TLS certificate for domains: " + strings.Join(domains, ", "))
		pub, priv := utils.GenerateRSAWebCertificates(domains)
		
		if !falledBack {
			baseMainConfig.HTTPConfig.TLSCert = pub
			baseMainConfig.HTTPConfig.TLSKey = priv
			baseMainConfig.HTTPConfig.TLSKeyHostsCached = domains
			baseMainConfig.HTTPConfig.TLSValidUntil = time.Now().AddDate(0, 0, 364)
			baseMainConfig.HTTPConfig.ForceHTTPSCertificateRenewal = false

			utils.SetBaseMainConfig(baseMainConfig)
			utils.Log("Saved new SELFISGNED TLS certificate")
		}

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

	utils.Log("Initialising HTTP(S) Router and all routes")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/logo", SendLogo)

	router.Use(middleware.Logger)
	router.Use(utils.SetSecurityHeaders)

	if config.BlockedCountries != nil && len(config.BlockedCountries) > 0 {
		router.Use(utils.BlockByCountryMiddleware(config.BlockedCountries))
	}
	
	srapi := router.PathPrefix("/cosmos").Subrouter()

	srapi.HandleFunc("/api/dns", GetDNSRoute)
	srapi.HandleFunc("/api/dns-check", CheckDNSRoute)
	
	srapi.HandleFunc("/api/status", StatusRoute)
	srapi.HandleFunc("/api/can-send-email", CanSendEmail)
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

	srapi.HandleFunc("/api/images/pull-if-missing", docker.PullImageIfMissing)
	srapi.HandleFunc("/api/images/pull", docker.PullImage)
	srapi.HandleFunc("/api/images", docker.InspectImageRoute)

	srapi.HandleFunc("/api/volume/{volumeName}", docker.DeleteVolumeRoute)
	srapi.HandleFunc("/api/volumes", docker.VolumesRoute)

	srapi.HandleFunc("/api/network/{networkID}", docker.DeleteNetworkRoute)
	srapi.HandleFunc("/api/networks", docker.NetworkRoutes)
	
	srapi.HandleFunc("/api/servapps/{containerId}/manage/{action}", docker.ManageContainerRoute)
	srapi.HandleFunc("/api/servapps/{containerId}/secure/{status}", docker.SecureContainerRoute)
	srapi.HandleFunc("/api/servapps/{containerId}/auto-update/{status}", docker.AutoUpdateContainerRoute)
	srapi.HandleFunc("/api/servapps/{containerId}/logs", docker.GetContainerLogsRoute)
	srapi.HandleFunc("/api/servapps/{containerId}/terminal/{action}", docker.TerminalRoute)
	srapi.HandleFunc("/api/servapps/{containerId}/update", docker.UpdateContainerRoute)
	srapi.HandleFunc("/api/servapps/{containerId}/", docker.GetContainerRoute)
	srapi.HandleFunc("/api/servapps/{containerId}/network/{networkId}", docker.NetworkContainerRoutes)
	srapi.HandleFunc("/api/servapps/{containerId}/networks", docker.NetworkContainerRoutes)
	srapi.HandleFunc("/api/servapps/{containerId}/check-update", docker.CanUpdateImageRoute)
	srapi.HandleFunc("/api/servapps", docker.ContainersRoute)
	srapi.HandleFunc("/api/docker-service", docker.CreateServiceRoute)
	
	srapi.HandleFunc("/api/markets", market.MarketGet)

	srapi.HandleFunc("/api/background", UploadBackground)
	srapi.HandleFunc("/api/background/{ext}", GetBackground)


	if(!config.HTTPConfig.AcceptAllInsecureHostname) {
		srapi.Use(utils.EnsureHostname)
	}

	SecureAPI(srapi, false)
	
	pwd,_ := os.Getwd()
	utils.Log("Starting in " + pwd)
	if _, err := os.Stat(pwd + "/static"); os.IsNotExist(err) {
		utils.Fatal("Static folder not found at " + pwd + "/static", err)
	}


	fs  := spa.SpaHandler(pwd + "/static", "index.html")
	
	if(!config.HTTPConfig.AcceptAllInsecureHostname) {
		fs = utils.EnsureHostname(fs)
	}

	router.PathPrefix("/cosmos-ui").Handler(http.StripPrefix("/cosmos-ui", fs))

	router = proxy.BuildFromConfig(router, HTTPConfig.ProxyConfig)
	
	router.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/cosmos-ui", http.StatusMovedPermanently)
	}))


	userRouter := router.PathPrefix("/oauth2").Subrouter()
	SecureAPI(userRouter, false)

	serverRouter := router.PathPrefix("/oauth2").Subrouter()
	SecureAPI(userRouter, true)

	wellKnownRouter := router.PathPrefix("/.well-known").Subrouter()
	SecureAPI(userRouter, true)

	authorizationserver.RegisterHandlers(wellKnownRouter, userRouter, serverRouter)

	return router
}

func StartServer() {
	config := utils.GetMainConfig()
	HTTPConfig := config.HTTPConfig

	var tlsCert = HTTPConfig.TLSCert
	var tlsKey= HTTPConfig.TLSKey

	router := InitServer()

	if (
		(
			HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["SELFSIGNED"] ||
			HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["PROVIDED"]  ||
			HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["LETSENCRYPT"]) &&
			tlsCert != "" && tlsKey != "") {
		utils.Log("TLS certificate exist, starting HTTPS servers and redirecting HTTP to HTTPS")
		startHTTPSServer(router, tlsCert, tlsKey)
	} else {
		utils.Log("TLS certificates do not exists or are disabled, starting HTTP server only")
		startHTTPServer(router)
	}
}

func RestartServer() {
	utils.Log("Restarting HTTP Server...")
	LoadConfig()

	go func() {
		HTTPServer.Shutdown(context.Background())
	}()

	utils.Log("HTTP Server stopped. Restarting...")
}
