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
		"github.com/azukaar/cosmos-server/src/constellation"
		"github.com/azukaar/cosmos-server/src/metrics"
		"github.com/azukaar/cosmos-server/src/cron"
		"github.com/azukaar/cosmos-server/src/storage"
		"github.com/azukaar/cosmos-server/src/backups"
		"github.com/gorilla/mux"
		"strconv"
		"time"
		"os"
		"net"
		"strings"
		"github.com/go-chi/httprate"
		"crypto/tls"
		"github.com/foomo/tlsconfig"
		"context"
    "net/http/pprof"
)

var serverPortHTTP = ""
var serverPortHTTPS = ""
var HTTPServer *http.Server
var HTTPServer2 *http.Server

func startHTTPServer(router *mux.Router) error {
	HTTPServer2 = nil
	HTTPServer = &http.Server{
		Addr: "0.0.0.0:" + serverPortHTTP,
		ReadTimeout: 0,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout: 0,
		IdleTimeout: 30 * time.Second,
		Handler: router,
		DisableGeneralOptionsHandler: true,
	}
	
	if utils.IsInsideContainer && !utils.IsHostNetwork {
		docker.CheckPorts()
	} else {
		proxy.InitInternalSocketProxy()
	}
	
	// Publish mDNS 
	if utils.GetMainConfig().HTTPConfig.PublishMDNS {
		proxy.PublishAllMDNSFromConfig()
	}

	utils.Log("Listening to HTTP on : 0.0.0.0:" + serverPortHTTP)

	return HTTPServer.ListenAndServe()
}

var primaryCert *tls.Certificate
var secondaryCert *tls.Certificate

// GetCertificate returns the appropriate certificate based on the client hello info
func GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	config := utils.GetMainConfig()
	
	if config.HTTPConfig.HTTPSCertificateMode == "PROVIDED" {
		return primaryCert, nil
	} else if config.HTTPConfig.HTTPSCertificateMode == "SELFSIGNED" {
		return secondaryCert, nil
	}

	host := clientHello.ServerName
	if host == "" {
			// If SNI is not used, check the connection's IP
			if clientHello.Conn != nil {
					host = clientHello.Conn.RemoteAddr().String()
					// Extract just the IP from the address
					if h, _, err := net.SplitHostPort(host); err == nil {
							host = h
					}
			}
	}

	// Check if we should use self-signed certificate
	shouldUseSelfSigned := false

	// Check for IP address
	if ip := net.ParseIP(host); ip != nil {
			shouldUseSelfSigned = true
	}

	// Check for .local domain
	if strings.HasSuffix(host, ".local") {
			shouldUseSelfSigned = true
	}

	// Check for localhost
	if host == "localhost" {
			shouldUseSelfSigned = true
	}

	if shouldUseSelfSigned {
			utils.Debug("Using self-signed certificate for: " + host)
			return secondaryCert, nil
	}

	return primaryCert, nil
}

func startHTTPSServer(router *mux.Router) error {
	//config  := utils.GetMainConfig()

	utils.IsHTTPS = true
		
	// redirect http to https
	go (func () {
		httpRouter := mux.NewRouter()

		httpRouter.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// if requested hostanme is 192.168.201.1 and path is /cosmos/api/constellation/config-sync
			if r.Host == "192.168.201.1" && (r.URL.Path == "/cosmos/api/constellation/config-sync" || r.URL.Path == "/cosmos/api/constellation_webhook_sync") && utils.IsConstellationIP(r.RemoteAddr) {
				router.ServeHTTP(w, r)
			} else if utils.GetMainConfig().HTTPConfig.AllowHTTPLocalIPAccess && utils.IsLocalIP(r.RemoteAddr)  {
				// use router 
				router.ServeHTTP(w, r)
			} else {
				// change port in host
				if strings.HasSuffix(r.Host, ":" + serverPortHTTP) {
					if serverPortHTTPS != "443" {
						r.Host = r.Host[:len(r.Host)-len(":" + serverPortHTTP)] + ":" + serverPortHTTPS
						} else {
						r.Host = r.Host[:len(r.Host)-len(":" + serverPortHTTP)]
					}
				}

				http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
			}
		})

		HTTPServer2 = &http.Server{
			Addr: "0.0.0.0:" + serverPortHTTP,
			ReadTimeout: 0,
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout: 0,
			IdleTimeout: 30 * time.Second,
			Handler: httpRouter,
			DisableGeneralOptionsHandler: true,
		}

		err := HTTPServer2.ListenAndServe()
		
		if err != nil && err != http.ErrServerClosed {
			utils.Fatal("Listening to HTTP (Redirecting to HTTPS)", err)
		}
	})()

	utils.Log("Listening to HTTP on :" + serverPortHTTP)
	utils.Log("Listening to HTTPS on :" + serverPortHTTPS)

	config := utils.GetMainConfig()
	HTTPConfig := config.HTTPConfig

	var tlsCert = HTTPConfig.TLSCert
	var tlsKey= HTTPConfig.TLSKey
	
	if (HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["SELFSIGNED"]) {
		tlsCert = HTTPConfig.SelfTLSCert
		tlsKey = HTTPConfig.SelfTLSKey
	}
	
	tlsConf := tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)

	cert, errCert := tls.X509KeyPair(([]byte)(tlsCert), ([]byte)(tlsKey))
	if errCert != nil {
		config.HTTPConfig.ForceHTTPSCertificateRenewal = true
		utils.SetBaseMainConfig(config)
		utils.Fatal("Getting Certificate pair", errCert)
	}

	tlsConf.Certificates = []tls.Certificate{cert}

	_primaryCert, errCert := tls.X509KeyPair(([]byte)(HTTPConfig.TLSCert), ([]byte)(HTTPConfig.TLSKey))
	if errCert != nil {
		config.HTTPConfig.ForceHTTPSCertificateRenewal = true
		utils.SetBaseMainConfig(config)
		utils.Fatal("Getting Certificate pair", errCert)
	}

	_secondaryCert, errCert := tls.X509KeyPair(([]byte)(HTTPConfig.SelfTLSCert), ([]byte)(HTTPConfig.SelfTLSKey))
	if errCert != nil {
		config.HTTPConfig.ForceHTTPSCertificateRenewal = true
		utils.SetBaseMainConfig(config)
		utils.Fatal("Getting Certificate pair", errCert)
	}

	primaryCert = &_primaryCert
	secondaryCert = &_secondaryCert

	HTTPServer = &http.Server{
    TLSConfig: &tls.Config{
			GetCertificate: GetCertificate,
		},
		Addr: "0.0.0.0:" + serverPortHTTPS,
		ReadTimeout: 0,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout: 0,
		IdleTimeout: 30 * time.Second,
		Handler: router,
		DisableGeneralOptionsHandler: true,
	}

	// Redirect ports 
	if utils.IsInsideContainer && !utils.IsHostNetwork {
		docker.CheckPorts()
	} else {
		proxy.InitInternalSocketProxy()
	}

	// Publish mDNS 
	if config.HTTPConfig.PublishMDNS {
		proxy.PublishAllMDNSFromConfig()
	}

	utils.Log("Now listening to HTTPS on :" + serverPortHTTPS)

	return HTTPServer.ListenAndServeTLS("", "")
}

func tokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Header.Del
		r.Header.Del("x-cosmos-user")
		r.Header.Del("x-cosmos-role")
		r.Header.Del("x-cosmos-user-role")
		r.Header.Del("x-cosmos-mfa")

		role, u, err := user.RefreshUserToken(w, r)

		if err != nil {
			return
		}

		r.Header.Set("x-cosmos-user", u.Nickname)
		r.Header.Set("x-cosmos-role", strconv.Itoa((int)(role)))
		r.Header.Set("x-cosmos-user-role", strconv.Itoa((int)(u.Role)))
		r.Header.Set("x-cosmos-mfa", strconv.Itoa((int)(u.MFAState)))

		next.ServeHTTP(w, r)
	})
}

func SecureAPI(userRouter *mux.Router, public bool, publicCors bool) {
	if(!public) {
		userRouter.Use(tokenMiddleware)
	}
	userRouter.Use(proxy.SmartShieldMiddleware(
		"__COSMOS",
		utils.ProxyRouteConfig{
			Name: "Cosmos-Internal",
			SmartShield: utils.SmartShieldPolicy{
				Enabled: true,
				PolicyStrictness: 1,
				PerUserRequestLimit: 12000,
			},
		},
	))

	if(publicCors || public) {
		userRouter.Use(utils.PublicCORS)
	}

	userRouter.Use(utils.MiddlewareTimeout(45 * time.Second))
	userRouter.Use(httprate.Limit(180, 1*time.Minute, 
		httprate.WithKeyFuncs(httprate.KeyByIP),
    httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			utils.Error("Too many requests. Throttling", nil)
			utils.HTTPError(w, "Too many requests", 
				http.StatusTooManyRequests, "HTTP003")
			return 
		}),
	))
}

func CertificateIsExpiredSoon(validUntil time.Time) bool {
	// allow 5 days of leeway
	isValid := time.Now().Add(45 * 24 * time.Hour).Before(validUntil)
	if !isValid {
		utils.TriggerEvent(
			"cosmos.proxy.certificate",
			"Cosmos Certificate Expire Soon",
			"warning",
			"",
			map[string]interface{}{
		})

		utils.Log("Certificate is not valid anymore. Needs refresh")
	}
	return isValid
}

func CertificateIsExpired(validUntil time.Time) bool {
	// allow 5 days of leeway
	isValid := time.Now().Before(validUntil)
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
	var tlsKey = HTTPConfig.TLSKey
	var selfTLSCert = HTTPConfig.SelfTLSCert
	var selfTLSKey = HTTPConfig.SelfTLSKey

	domains := utils.GetAllHostnames(true, true)
	oldDomains := baseMainConfig.HTTPConfig.TLSKeyHostsCached
	// falledBack := false

	// Check if self-signed certificate needs refresh
	selfSignedNeedsRefresh := baseMainConfig.HTTPConfig.ForceHTTPSCertificateRenewal ||
			(selfTLSCert == "" || selfTLSKey == "") ||
			utils.HasAnyNewItem(domains, baseMainConfig.HTTPConfig.SelfTLSKeyHostsCached) ||
			!CertificateIsExpiredSoon(baseMainConfig.HTTPConfig.SelfTLSValidUntil)
	
	// Always ensure we have a valid self-signed certificate
	if selfSignedNeedsRefresh {
			utils.Log("Generating new self-signed TLS certificate for domains: " + strings.Join(domains, ", "))
			pub, priv, err := utils.GenerateRSAWebCertificates(domains)
			baseMainConfig = utils.GetBaseMainConfig()

			if err != nil {
					utils.Error("Generating self-signed TLS certificate", err)
			} else {
					baseMainConfig.HTTPConfig.SelfTLSCert = pub
					baseMainConfig.HTTPConfig.SelfTLSKey = priv
					baseMainConfig.HTTPConfig.SelfTLSKeyHostsCached = domains
					baseMainConfig.HTTPConfig.SelfTLSValidUntil = time.Now().AddDate(0, 0, 364)
					utils.SetBaseMainConfig(baseMainConfig)
					utils.Log("Saved new self-signed TLS certificate")
					
					selfTLSCert = pub
					selfTLSKey = priv
			}
	}

	// Check if Let's Encrypt certificate needs refresh
	letsEncryptNeedsRefresh := baseMainConfig.HTTPConfig.ForceHTTPSCertificateRenewal || 
			(tlsCert == "" || tlsKey == "") || 
			utils.HasAnyNewItem(domains, oldDomains) || 
			!CertificateIsExpiredSoon(baseMainConfig.HTTPConfig.TLSValidUntil)

	// If we have a certificate, we can fallback to it if necessary
	CanFallback := tlsCert != "" && tlsKey != "" && 
		len(config.HTTPConfig.TLSKeyHostsCached) > 0 && 
		config.HTTPConfig.TLSKeyHostsCached[0] == config.HTTPConfig.Hostname  &&
		CertificateIsExpired(baseMainConfig.HTTPConfig.TLSValidUntil)

	if letsEncryptNeedsRefresh && config.HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["LETSENCRYPT"] {
			if config.HTTPConfig.DNSChallengeProvider != "" {
					newEnv := config.HTTPConfig.DNSChallengeConfig
					for key, value := range newEnv {
							os.Setenv(key, value)
					}
			}

			// Get Certificates
			pub, priv := utils.DoLetsEncrypt()

			if pub == "" || priv == "" {
				if(!CanFallback) {
					utils.MajorError("Couldn't get TLS certificate. Fallback to SELFSIGNED certificates since none exists", nil)
					HTTPConfig.HTTPSCertificateMode = utils.HTTPSCertModeList["SELFSIGNED"]
					// falledBack = true
				} else {
					utils.MajorError("Couldn't get TLS certificate. Fallback to previous certificate", nil)
				}
			} else {
					baseMainConfig.HTTPConfig.TLSCert = pub
					baseMainConfig.HTTPConfig.TLSKey = priv
					baseMainConfig.HTTPConfig.TLSKeyHostsCached = domains
					baseMainConfig.HTTPConfig.TLSValidUntil = time.Now().AddDate(0, 0, 90)
					baseMainConfig.HTTPConfig.ForceHTTPSCertificateRenewal = false
	
					utils.SetBaseMainConfig(baseMainConfig)
					utils.Log("Saved new LETSENCRYPT TLS certificate")
	
					utils.TriggerEvent(
							"cosmos.proxy.certificate",
							"Cosmos Certificate Renewed",
							"important",
							"",
							map[string]interface{}{
									"domains": domains,
					})

					utils.WriteNotification(utils.Notification{
							Recipient: "admin",
							Title: "header.notification.title.certificateRenewed",
							Message: "header.notification.message.certificateRenewed",
							Vars: strings.Join(domains, ", "),
							Level: "info",
					})

					tlsCert = pub
					tlsKey = priv
			}
	}
	
	// Only update serving certificates when in SELFSIGNED mode
	if config.HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["SELFSIGNED"] {
			// Use the self-signed certificates for serving
			tlsCert = selfTLSCert
			tlsKey = selfTLSKey
	}

	utils.Log("Initialising HTTP(S) Router and all routes")

	router := mux.NewRouter().StrictSlash(true)
	
	router.Use(utils.BlockBannedIPs)

	router.Use(utils.Logger)

	if config.BlockedCountries != nil && len(config.BlockedCountries) > 0 {
		router.Use(utils.BlockByCountryMiddleware(config.BlockedCountries, config.CountryBlacklistIsWhitelist))
	}

	// robots.txt
	if !config.HTTPConfig.AllowSearchEngine {
		router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("User-agent: *\nDisallow: /"))
		})
	}
	
	logoAPI := router.PathPrefix("/logo").Subrouter()
	SecureAPI(logoAPI, true, true)
	logoAPI.HandleFunc("/", SendLogo)
	
	
	srapi := router.PathPrefix("/cosmos").Subrouter()
	srapi.Use(utils.ContentTypeMiddleware("application/json"))
	
	srapi.HandleFunc("/api/login", user.UserLogin)
	srapi.HandleFunc("/api/sudo", user.UserSudo)
	srapi.HandleFunc("/api/password-reset", user.ResetPassword)
	srapi.HandleFunc("/api/mfa", user.API2FA)
	srapi.HandleFunc("/api/status", StatusRoute)
	srapi.HandleFunc("/api/can-send-email", CanSendEmail)
	srapi.HandleFunc("/api/newInstall", NewInstallRoute)
	srapi.HandleFunc("/api/logout", user.UserLogout)
	srapi.HandleFunc("/api/register", user.UserRegister)
	srapi.HandleFunc("/api/dns", GetDNSRoute)
	srapi.HandleFunc("/api/dns-check", CheckDNSRoute)
	srapi.HandleFunc("/api/favicon", GetFavicon)
	srapi.HandleFunc("/api/ping", PingURL)
	srapi.HandleFunc("/api/me", user.Me)

	srapi.HandleFunc("/api/terminal/{route}", HostTerminalRoute)
	
	srapiAdmin := router.PathPrefix("/cosmos").Subrouter()
	srapiAdmin.Use(utils.ContentTypeMiddleware("application/json"))

	srapiAdmin.HandleFunc("/api/restart-server", restartHostMachineRoute)
	srapiAdmin.HandleFunc("/_logs", LogsRoute)
	srapiAdmin.HandleFunc("/api/force-server-update", ForceUpdateRoute)
	srapiAdmin.HandleFunc("/api/config", configapi.ConfigRoute)
	srapiAdmin.HandleFunc("/api/_memory", MemStatusRoute)
	srapiAdmin.HandleFunc("/api/restart", configapi.ConfigApiRestart)
	
	srapiAdmin.HandleFunc("/api/invite", user.UserResendInviteLink)
	srapiAdmin.HandleFunc("/api/users/{nickname}", user.UsersIdRoute)
	srapiAdmin.HandleFunc("/api/users", user.UsersRoute)

	srapiAdmin.HandleFunc("/api/images/pull-if-missing", docker.PullImageIfMissing)
	srapiAdmin.HandleFunc("/api/images/pull", docker.PullImage)
	srapiAdmin.HandleFunc("/api/images", docker.InspectImageRoute)

	srapiAdmin.HandleFunc("/api/volume/{volumeName}", docker.DeleteVolumeRoute)
	srapiAdmin.HandleFunc("/api/volumes", docker.VolumesRoute)

	srapiAdmin.HandleFunc("/api/network/{networkID}", docker.DeleteNetworkRoute)
	srapiAdmin.HandleFunc("/api/networks", docker.NetworkRoutes)

	srapiAdmin.HandleFunc("/api/migrate-host", docker.MigrateToHostModeRoute)
	
	srapiAdmin.HandleFunc("/api/servapps/{containerId}/manage/{action}", docker.ManageContainerRoute)
	srapiAdmin.HandleFunc("/api/servapps/{containerId}/secure/{status}", docker.SecureContainerRoute)
	srapiAdmin.HandleFunc("/api/servapps/{containerId}/auto-update/{status}", docker.AutoUpdateContainerRoute)
	srapiAdmin.HandleFunc("/api/servapps/{containerId}/logs", docker.GetContainerLogsRoute)
	srapiAdmin.HandleFunc("/api/servapps/{containerId}/terminal/{action}", docker.TerminalRoute)
	srapiAdmin.HandleFunc("/api/servapps/{containerId}/update", docker.UpdateContainerRoute)
	srapiAdmin.HandleFunc("/api/servapps/{containerId}/export", docker.ExportContainerRoute)
	srapiAdmin.HandleFunc("/api/servapps/{containerId}/", docker.GetContainerRoute)
	srapiAdmin.HandleFunc("/api/servapps/{containerId}/network/{networkId}", docker.NetworkContainerRoutes)
	srapiAdmin.HandleFunc("/api/servapps/{containerId}/networks", docker.NetworkContainerRoutes)
	srapiAdmin.HandleFunc("/api/servapps/{containerId}/check-update", docker.CanUpdateImageRoute)
	srapiAdmin.HandleFunc("/api/servapps", docker.ContainersRoute)
	srapiAdmin.HandleFunc("/api/docker-service", docker.CreateServiceRoute)
	
	srapiAdmin.HandleFunc("/api/markets", market.MarketGet)

	srapiAdmin.HandleFunc("/api/upload/{name}", UploadImage)
	srapiAdmin.HandleFunc("/api/image/{name}", GetImage)

	srapiAdmin.HandleFunc("/api/get-backup", configapi.BackupFileApiGet)

	srapiAdmin.HandleFunc("/api/constellation/devices", constellation.ConstellationAPIDevices)
	srapiAdmin.HandleFunc("/api/constellation/restart", constellation.API_Restart)
	srapiAdmin.HandleFunc("/api/constellation/reset", constellation.API_Reset)
	srapiAdmin.HandleFunc("/api/constellation/connect", constellation.API_ConnectToExisting)
	srapiAdmin.HandleFunc("/api/constellation/config", constellation.API_GetConfig)
	srapiAdmin.HandleFunc("/api/constellation/logs", constellation.API_GetLogs)
	srapiAdmin.HandleFunc("/api/constellation/block", constellation.DeviceBlock)
	srapiAdmin.HandleFunc("/api/constellation/ping", constellation.API_Ping)
	// device request config
	srapiAdmin.HandleFunc("/api/constellation/config-sync", constellation.GetDeviceConfigSync)
	// user manually request constellation config for resync
	srapiAdmin.HandleFunc("/api/constellation/config-manual-sync", constellation.GetDeviceConfigManualSync)

	srapiAdmin.HandleFunc("/api/events", metrics.API_ListEvents)

	srapiAdmin.HandleFunc("/api/metrics", metrics.API_GetMetrics)
	srapiAdmin.HandleFunc("/api/reset-metrics", metrics.API_ResetMetrics)
	srapiAdmin.HandleFunc("/api/list-metrics", metrics.ListMetrics)

	srapiAdmin.HandleFunc("/api/notifications/read", utils.MarkAsRead)
	srapiAdmin.HandleFunc("/api/notifications", utils.NotifGet)

	srapiAdmin.HandleFunc("/api/listen=jobs", cron.ListJobs)
	srapiAdmin.HandleFunc("/api/jobs", cron.ListJobs)
	srapiAdmin.HandleFunc("/api/jobs/stop", cron.StopJobRoute)
	srapiAdmin.HandleFunc("/api/jobs/run", cron.RunJobRoute)
	srapiAdmin.HandleFunc("/api/jobs/get", cron.GetJobRoute)
	srapiAdmin.HandleFunc("/api/jobs/delete", cron.DeleteJobRoute)
	srapiAdmin.HandleFunc("/api/jobs/running", cron.GetRunningJobsRoute)

	srapiAdmin.HandleFunc("/api/smart-def", storage.ListSmartDef)
	srapiAdmin.HandleFunc("/api/disks", storage.ListDisksRoute)
	srapiAdmin.HandleFunc("/api/disks/format", storage.FormatDiskRoute)
	srapiAdmin.HandleFunc("/api/mounts", storage.ListMountsRoute)
	srapiAdmin.HandleFunc("/api/mount", storage.MountRoute)
	srapiAdmin.HandleFunc("/api/unmount", storage.UnmountRoute)
	srapiAdmin.HandleFunc("/api/merge", storage.MergeRoute)
	srapiAdmin.HandleFunc("/api/snapraid", storage.SNAPRaidCRUDRoute)
	srapiAdmin.HandleFunc("/api/snapraid/{name}", storage.SnapRAIDEditRoute)
	srapiAdmin.HandleFunc("/api/snapraid/{name}/{action}", storage.SnapRAIDRunRoute)
	srapiAdmin.HandleFunc("/api/rclone-restart", storage.API_Rclone_remountAll)
	srapiAdmin.HandleFunc("/api/list-dir", storage.ListDirectoryRoute)
	srapiAdmin.HandleFunc("/api/new-dir", storage.CreateFolderRoute)

	srapiAdmin.HandleFunc("/api/backups-repository", backups.ListRepos)
	srapiAdmin.HandleFunc("/api/backups-repository/{name}/snapshots", backups.ListSnapshotsRouteFromRepo)
	srapiAdmin.HandleFunc("/api/backups/{name}/snapshots", backups.ListSnapshotsRoute)
	srapiAdmin.HandleFunc("/api/backups/{name}/{snapshot}/folders", backups.ListFoldersRoute) 
	srapiAdmin.HandleFunc("/api/backups/{name}/restore", backups.RestoreBackupRoute)
	srapiAdmin.HandleFunc("/api/backups", backups.AddBackupRoute)
	srapiAdmin.HandleFunc("/api/backups/edit", backups.EditBackupRoute)
	srapiAdmin.HandleFunc("/api/backups/{name}", backups.RemoveBackupRoute)
	srapiAdmin.HandleFunc("/api/backups/{name}/{snapshot}/forget", backups.ForgetSnapshotRoute)
	srapiAdmin.HandleFunc("/api/backups/{name}/{snapshot}/subfolder-restore-size", backups.StatsRepositorySubfolderRoute)

	// srapiAdmin.HandleFunc("/api/storage/raid", storage.RaidListRoute).Methods("GET")
	// srapiAdmin.HandleFunc("/api/storage/raid", storage.RaidCreateRoute).Methods("POST")
	// srapiAdmin.HandleFunc("/api/storage/raid/{name}", storage.RaidDeleteRoute).Methods("DELETE")
	// srapiAdmin.HandleFunc("/api/storage/raid/{name}/status", storage.RaidStatusRoute).Methods("GET")
	// srapiAdmin.HandleFunc("/api/storage/raid/{name}/device", storage.RaidAddDeviceRoute).Methods("POST")
	// srapiAdmin.HandleFunc("/api/storage/raid/{name}/replace", storage.RaidReplaceDeviceRoute).Methods("POST")
	// srapiAdmin.HandleFunc("/api/storage/raid/{name}/resize", storage.RaidResizeRoute).Methods("POST")

	if utils.LoggingLevelLabels[utils.GetMainConfig().LoggingLevel] == utils.DEBUG {
		debugRouter := srapiAdmin.PathPrefix("/debug").Subrouter()
		debugRouter.Use(utils.AdminOnlyMiddleware)
		debugRouter.HandleFunc("/pprof/", pprof.Index)
		debugRouter.HandleFunc("/pprof/allocs", pprof.Cmdline)
		debugRouter.HandleFunc("/pprof/cmdline", pprof.Cmdline)
		debugRouter.HandleFunc("/pprof/profile", pprof.Profile)
		debugRouter.HandleFunc("/pprof/symbol", pprof.Symbol)
		debugRouter.HandleFunc("/pprof/trace", pprof.Trace)
		debugRouter.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
		debugRouter.Handle("/pprof/heap", pprof.Handler("heap"))
		debugRouter.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
		debugRouter.Handle("/pprof/block", pprof.Handler("block"))
	}

	srapiAdmin.Use(utils.Restrictions(config.AdminConstellationOnly, config.AdminWhitelistIPs))

	srapi.Use(utils.SetSecurityHeaders)
	srapiAdmin.Use(utils.SetSecurityHeaders)

	if(!config.HTTPConfig.AcceptAllInsecureHostname) {
		srapi.Use(utils.EnsureHostname)
		srapiAdmin.Use(utils.EnsureHostname)
	
		srapi.Use(utils.EnsureHostnameCosmosAPI)
		srapiAdmin.Use(utils.EnsureHostnameCosmosAPI)
	}

	SecureAPI(srapi, false, false)
	SecureAPI(srapiAdmin, false, false)
	
	pwd,_ := os.Getwd()
	utils.Log("Starting in " + pwd)
	if _, err := os.Stat(pwd + "/static"); os.IsNotExist(err) {
		utils.Fatal("Static folder not found at " + pwd + "/static", err)
	}

	// fs := http.FileServer(http.Dir(pwd + "/static"))
	uirouter := router.PathPrefix("/cosmos-ui").Subrouter()
	uirouter.Use(utils.SetSecurityHeaders)
	SecureAPI(uirouter, true, true)
	uirouter.PathPrefix("/").Handler(http.StripPrefix("/cosmos-ui", utils.SPAHandler(pwd + "/static")))
	
	if(!config.HTTPConfig.AcceptAllInsecureHostname) {
		uirouter.Use(utils.EnsureHostname)
	}

	OpenIDDetect := router.PathPrefix("/").Subrouter()
	SecureAPI(OpenIDDetect, true, true)
	authorizationserver.RegisterHandlersDetect(OpenIDDetect, srapi)

	router = proxy.BuildFromConfig(router, HTTPConfig.ProxyConfig)

	wellKnownRouter := router.PathPrefix("/").Subrouter()
	SecureAPI(wellKnownRouter, true, true)

	userRouter := router.PathPrefix("/oauth2").Subrouter()
	SecureAPI(userRouter, false, true)

	serverRouter := router.PathPrefix("/oauth2").Subrouter()
	SecureAPI(serverRouter, true, true)

	authorizationserver.RegisterHandlers(wellKnownRouter, userRouter, serverRouter)
	
	router.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/cosmos-ui/", http.StatusTemporaryRedirect)
	}))

	router.HandleFunc("/cosmos-ui", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/cosmos-ui/", http.StatusTemporaryRedirect)
	}))

	return router
}

func StartServer() {
	router := InitServer()

	// start https server
	var errServ error

	for(errServ == http.ErrServerClosed || errServ == nil) {
		config := utils.GetMainConfig()
		HTTPConfig := config.HTTPConfig

		var tlsCert = HTTPConfig.TLSCert
		var tlsKey= HTTPConfig.TLSKey

		if (HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["SELFSIGNED"]) {
			tlsCert = HTTPConfig.SelfTLSCert
			tlsKey = HTTPConfig.SelfTLSKey
		}

		if (
			(
				HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["SELFSIGNED"] ||
				HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["PROVIDED"]  ||
				HTTPConfig.HTTPSCertificateMode == utils.HTTPSCertModeList["LETSENCRYPT"]) &&
				tlsCert != "" && tlsKey != "") {
			utils.Log("TLS certificate exist, starting HTTPS servers and redirecting HTTP to HTTPS")
			errServ = startHTTPSServer(router)
		} else {
			utils.Log("TLS certificates do not exists or are disabled, starting HTTP server only")
			errServ = startHTTPServer(router)
		}
		if errServ != nil && errServ != http.ErrServerClosed {
			utils.Fatal("Listening to HTTPS", errServ)
		}

		utils.Log("HTTPS Server closed. Restarting.")
		errServ = nil
		router = InitServer()
	}
}

func RestartServer() {
	utils.LetsEncryptErrors = []string{}
	IconCache = map[string]CachedImage{}
	
	utils.Log("Restarting HTTP Server...")
	LoadConfig()

	authorizationserver.Init()

	docker.BootstrapAllContainersFromTags()

	go func() {
		if HTTPServer2 != nil {
			HTTPServer2.Shutdown(context.Background())
		}
		HTTPServer.Shutdown(context.Background())
	}()

	utils.Log("HTTPServer shutdown.")
}
