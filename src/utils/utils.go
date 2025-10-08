package utils

import (
	"encoding/json"
	"math/rand"
	"regexp"
	"net/http"
	"encoding/base64"
	"os"
	"io"
	"strconv"
	"strings"
	"io/ioutil"
	"fmt"
	"sync"
	"time"
	"errors"
	"path/filepath"
	"os/exec"
	"encoding/hex"
	"crypto/sha256"
	_url "net/url"
	osnet "net"
	
	
	"github.com/ory/fosite"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
	"golang.org/x/net/publicsuffix"
	"github.com/Masterminds/semver"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/mongo"
)

var ConfigLock sync.Mutex
var ConfigLockInternal sync.Mutex

var ProxyRClone = false
var ProxyRCloneUser = ""
var ProxyRClonePwd = ""

var BaseMainConfig Config
var MainConfig Config
var IsHTTPS = false
var NewVersionAvailable = false

var NeedsRestart = false

var IsHostNetwork = false

var UpdateAvailable = map[string]bool{}

var RestartHTTPServer = func() {}

// var ReBootstrapContainer func(string) error
var GetContainerIPByName func(string) (string, error)
var DoesContainerExist func(string) bool
var CheckDockerNetworkMode func() string
var WaitForAllJobs func()
var StopAllRCloneProcess func(bool)

var ResyncConstellationNodes = func() {}

var LetsEncryptErrors = []string{}

var CONFIGFOLDER = "/var/lib/cosmos/"

var ConstellationSlaveIPWarning = ""

var IsInsideContainer = false

var DefaultConfig = Config{
	LoggingLevel: "INFO",
	NewInstall:   true,
	AutoUpdate:	  true,
	BlockedCountries: []string{
	},
	HTTPConfig: HTTPConfig{
		HTTPSCertificateMode:    "DISABLED",
		GenerateMissingAuthCert: true,
		HTTPPort:                "80",
		HTTPSPort:               "443",
		Hostname:                "0.0.0.0",
		PublishMDNS:						 true,
		ProxyConfig: ProxyConfig{
			Routes: []ProxyRouteConfig{},
		},
	},
	DockerConfig: DockerConfig{
		DefaultDataPath: "/cosmos-storage",
	},
  MarketConfig: MarketConfig{
    Sources: []MarketSource{
		},
	},
  ConstellationConfig: ConstellationConfig{
    Enabled: false,
		DNSDisabled: false,
		DNSFallback: "8.8.8.8:53",
		DNSAdditionalBlocklists: []string{
			"https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt",
      "https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
      "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
      "https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/fakenews-only/hosts",
		},
	},
  MonitoringAlerts: map[string]Alert{
    "Anti Crypto-Miner": {
      Name: "Anti Crypto-Miner",
      Enabled: false,
      Period: "daily",
      TrackingMetric: "cosmos.system.docker.cpu.*",
			LastTriggered: time.Time{},
      Condition: AlertCondition {
        Operator: "gt",
        Value: 80,
        Percent: false,
      },
      Actions: []AlertAction {
				AlertAction {
          Type: "notification",
          Target: "",
        },
        AlertAction {
          Type: "email",
          Target: "",
        },
        AlertAction {
          Type: "stop",
          Target: "",
        },
			},
      Throttled: false,
      Severity: "warn",
    },
    "Anti Memory Leak": {
      Name: "Anti Memory Leak",
      Enabled: false,
      Period: "daily",
      TrackingMetric: "cosmos.system.docker.ram.*",
			LastTriggered: time.Time{},
      Condition: AlertCondition {
        Operator: "gt",
        Value: 80,
        Percent: true,
      },
      Actions: []AlertAction {
        {
          Type: "notification",
          Target: "",
        },
        {
          Type: "email",
          Target: "",
        },
        {
          Type: "stop",
          Target: "",
        },
      },
      Throttled: false,
      Severity: "warn",
    },
    "Disk Health": {
      Name: "Disk Health",
      Enabled: true,
      Period: "latest",
      TrackingMetric: "system.disk-health.temperature.*",
			LastTriggered: time.Time{},
      Condition: AlertCondition {
        Operator: "gt",
        Value: 50,
        Percent: false,
      },
      Actions: []AlertAction {
        {
          Type: "notification",
          Target: "",
        },
      },
      Throttled: true,
      Severity: "warn",
    },
    "Disk Full Notification": {
      Name: "Disk Full Notification",
      Enabled: true,
      Period: "latest",
      TrackingMetric: "cosmos.system.disk./",
			LastTriggered: time.Time{},
      Condition: AlertCondition {
        Operator: "gt",
        Value: 95,
        Percent: true,
      },
      Actions: []AlertAction {
        {
          Type: "notification",
          Target: "",
        },
			},
      Throttled: true,
      Severity: "warn",
    },
  },
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	Error("Reading file error: ", err)
	return false
}

func GetRootAppId() string {
	return "COSMOS"
}

func GetPrivateAuthKey() string {
	return MainConfig.HTTPConfig.AuthPrivateKey
}

func GetPublicAuthKey() string {
	return MainConfig.HTTPConfig.AuthPublicKey
}

var AlphaNumRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenerateRandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = AlphaNumRunes[rand.Intn(len(AlphaNumRunes))]
	}
	return string(b)
}

type HTTPErrorResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func HTTPError(w http.ResponseWriter, message string, code int, userCode string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(HTTPErrorResult{
		Status:  "error",
		Message: message,
		Code:    userCode,
	})
	Error("HTTP Request returned Error "+strconv.Itoa(code)+" : "+message, nil)
}

func SetBaseMainConfig(config Config) {
	SaveConfigTofile(config)
	LoadBaseMainConfig(config)
}

func ReadConfigFromFile() Config {
	ConfigLockInternal.Lock()
	defer ConfigLockInternal.Unlock()

	configFile := GetConfigFileName()
	Log("Using config file: " + configFile)
	if CreateDefaultConfigFileIfNecessary() {
		LoadBaseMainConfig(DefaultConfig)
		return DefaultConfig
	}

	file, err := os.Open(configFile)
	if err != nil {
		Fatal("Opening Config File: ", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)

	// check file is not empty
	if err != nil {
		// check error is not empty 
		if err.Error() == "EOF" {
			Fatal("Reading Config File: File is empty.", err)
		}

		// get error string 
		errString := err.Error()

		// replace string in error
		m1 := regexp.MustCompile(`json: cannot unmarshal ([A-Za-z\.]+) into Go struct field ([A-Za-z\.]+) of type ([A-Za-z\.]+)`)
		errString = m1.ReplaceAllString(errString, "Invalid JSON in config file.\n > Field $2 is wrong.\n > Type is $1 Should be $3")
		Fatal("Reading Config File: " + errString, err)
	}

	return config
}

func LoadBaseMainConfig(config Config) {
	BaseMainConfig = config
	MainConfig = config

	// use ENV to overwrite configs

	if os.Getenv("COSMOS_HTTP_PORT") != "" {
		MainConfig.HTTPConfig.HTTPPort = os.Getenv("COSMOS_HTTP_PORT")
	}
	if os.Getenv("COSMOS_HTTPS_PORT") != "" {
		MainConfig.HTTPConfig.HTTPSPort = os.Getenv("COSMOS_HTTPS_PORT")
	}
	if os.Getenv("COSMOS_HOSTNAME") != "" {
		MainConfig.HTTPConfig.Hostname = os.Getenv("COSMOS_HOSTNAME")
	}
	if os.Getenv("COSMOS_HTTPS_MODE") != "" {
		MainConfig.HTTPConfig.HTTPSCertificateMode = os.Getenv("COSMOS_HTTPS_MODE")
	}
	if os.Getenv("COSMOS_GENERATE_MISSING_AUTH_CERT") != "" {
		MainConfig.HTTPConfig.GenerateMissingAuthCert = os.Getenv("COSMOS_GENERATE_MISSING_AUTH_CERT") == "true"
	}
	if os.Getenv("COSMOS_TLS_CERT") != "" {
		MainConfig.HTTPConfig.TLSCert = os.Getenv("COSMOS_TLS_CERT")
	}
	if os.Getenv("COSMOS_TLS_KEY") != "" {
		MainConfig.HTTPConfig.TLSKey = os.Getenv("COSMOS_TLS_KEY")
	}
	if os.Getenv("COSMOS_AUTH_PRIV_KEY") != "" {
		MainConfig.HTTPConfig.AuthPrivateKey = os.Getenv("COSMOS_AUTH_PRIVATE_KEY")
	}
	if os.Getenv("COSMOS_AUTH_PUBLIC_KEY") != "" {
		MainConfig.HTTPConfig.AuthPublicKey = os.Getenv("COSMOS_AUTH_PUBLIC_KEY")
	}
	if os.Getenv("COSMOS_LOG_LEVEL") != "" {
		MainConfig.LoggingLevel = (LoggingLevel)(os.Getenv("COSMOS_LOG_LEVEL"))
	}
	if os.Getenv("COSMOS_MONGODB") != "" {
		MainConfig.MongoDB = os.Getenv("COSMOS_MONGODB")
	}
	if os.Getenv("COSMOS_SERVER_COUNTRY") != "" {
		MainConfig.ServerCountry = os.Getenv("COSMOS_SERVER_COUNTRY")
	}
	
	if MainConfig.DockerConfig.DefaultDataPath == "" {
		MainConfig.DockerConfig.DefaultDataPath = "/usr"
	}
}

func GetMainConfig() Config {
	return MainConfig
}

func GetBaseMainConfig() Config {
	return BaseMainConfig
}

func Sanitize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func SanitizeSafe(s string) string {
	return strings.TrimSpace(s)
}

func SanitizeNoSpace(s string) string {
	s1 := strings.ToLower(strings.TrimSpace(s))
	s1 = strings.ReplaceAll(s1, " ", "-")
	s1 = strings.ReplaceAll(s1, "@", "-")
	return s1
}

func GetConfigFileName() string {	
	configFile := os.Getenv("CONFIG_FILE")

	if configFile == "" {
		configFile = CONFIGFOLDER + "cosmos.config.json"
	}

	return configFile
}

func CreateDefaultConfigFileIfNecessary() bool {
	configFile := GetConfigFileName()

	// get folder path
	folderPath := strings.Split(configFile, "/")
	folderPath = folderPath[:len(folderPath)-1]
	folderPathString := strings.Join(folderPath, "/")
	os.MkdirAll(folderPathString, os.ModePerm)

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		Log("Config file does not exist. Creating default config file.")
		file, err := os.Create(configFile)
		if err != nil {
			Fatal("Creating Default Config File", err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(DefaultConfig)
		if err != nil {
			Fatal("Writing Default Config File", err)
		}

		return true
	}
	return false
}

func SaveConfigTofile(config Config) {
	ConfigLockInternal.Lock()
	defer ConfigLockInternal.Unlock()

	configFile := GetConfigFileName()
	CreateDefaultConfigFileIfNecessary()

	file, err := os.OpenFile(configFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		Fatal("Opening Config File", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(config)
	if err != nil {
		Fatal("Writing Config File", err)
	}

	Log("Config file saved.")
}

func RestartServer() {
	Log("Restarting server...")
	WaitForAllJobs() 
	StopAllRCloneProcess(false)
	os.Exit(0)
}

func LetsEncryptValidOnly(hostnames []string, acceptWildcard bool) []string {
	wrongPattern := `^(localhost(:\d+)?|(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(:\d+)?|.*\.local(:\d+)?)$`
	
	re, _ := regexp.Compile(wrongPattern)

	var validDomains []string
	for _, domain := range hostnames {
		if !re.MatchString(domain) && (acceptWildcard || !strings.Contains(domain, "*")) && !strings.Contains(domain, " ")  && !strings.Contains(domain, "::") && !strings.Contains(domain, ",") {
			validDomains = append(validDomains, domain)
		} else {
			Warn("Invalid domain found in URLs: " + domain + " it was removed from the certificate to not break Let's Encrypt")
		}
	}

	return validDomains
}

func RemoveStringFromSlice(slice []string, s string) []string {
	for i, v := range slice {
		if v == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func filterHostnamesByWildcard(hostnames []string, wildcards []string) []string {
	finalHostnames := make([]string, len(hostnames))
	copy(finalHostnames, hostnames)

	for _, wildcard := range wildcards {
		for _, hostname := range hostnames {
			if strings.HasSuffix(hostname, wildcard[1:]) && hostname != wildcard[2:] {
				// remove hostname
				finalHostnames = RemoveStringFromSlice(finalHostnames, hostname)
			}
		}
	}

	return finalHostnames
}

var nonWildcardableDomains = []string{
	"localhost",
	".local",
	".invalid",
	".test",
	".example",
	".onion",
}

// canDomainWildcard checks if a domain can be wildcarded
func canDomainWildcard(domain string) bool {
	// Normalize domain (convert to lowercase)
	domain = strings.ToLower(strings.TrimSpace(domain))

	// check if IP
	if osnet.ParseIP(domain) != nil {
		return false
	}

	// Direct match for localhost
	if domain == "localhost" {
		return false
	}

	// Check if the domain ends with any non-wildcardable suffix
	for _, suffix := range nonWildcardableDomains {
		if strings.HasSuffix(domain, suffix) {
			return false
		}
	}

	// Domains without a dot cannot be wildcarded (e.g., single-word hostnames)
	if !strings.Contains(domain, ".") {
		return false
	}

	// Valid wildcardable domain
	return true
}

func GetAllHostnames(applyWildCard bool, removePorts bool) []string {
	mainHostname := GetMainConfig().HTTPConfig.Hostname
	OverrideWildcardDomains := GetMainConfig().HTTPConfig.OverrideWildcardDomains

	if removePorts {
		mainHostname = strings.Split(mainHostname, ":")[0]
	}

	hostnames := []string{
		mainHostname,
	}

	proxies := GetMainConfig().ConstellationConfig.Tunnels
	proxies = append(proxies, GetMainConfig().HTTPConfig.ProxyConfig.Routes...)
	
	for _, proxy := range proxies {
		if proxy.UseHost && proxy.Host != "" && !strings.Contains(proxy.Host, ",") && !strings.Contains(proxy.Host, " ") {
			if removePorts {
				hostnames = append(hostnames, strings.Split(proxy.Host, ":")[0])
			} else {
				hostnames = append(hostnames, proxy.Host)
			}
		}
	}

	// remove doubles
	seen := make(map[string]bool)
	uniqueHostnames := []string{}
	for _, hostname := range hostnames {
		if _, ok := seen[hostname]; !ok {
			seen[hostname] = true
			uniqueHostnames = append(uniqueHostnames, hostname)
		}
	}

	bareMainHostname, _ := publicsuffix.EffectiveTLDPlusOne(mainHostname)
	canWildcard := canDomainWildcard(bareMainHostname) || (OverrideWildcardDomains != "")

	if applyWildCard && MainConfig.HTTPConfig.UseWildcardCertificate && canWildcard {

		Debug("bareMainHostname: " + bareMainHostname)

		filteredHostnames := []string{
			bareMainHostname,
			"*." + bareMainHostname,
		}

		if(OverrideWildcardDomains != "") {
			filteredHostnames = strings.Split(OverrideWildcardDomains, ",")
		}

		wildcards := []string{}
		othersHostname := []string{}
		for _, hostname := range append(uniqueHostnames, filteredHostnames...) {
			if strings.HasPrefix(hostname, "*.") {
				wildcards = append(wildcards, hostname)
			} else {
				othersHostname = append(othersHostname, hostname)
			}
		}

		tempUniqueHostnames := append(wildcards, filterHostnamesByWildcard(othersHostname, wildcards)...)

		// hardcode wildcard for local domains
		// if(MainConfig.HTTPConfig.HTTPSCertificateMode == HTTPSCertModeList["SELFSIGNED"]) {
		// 	for _, hostname := range tempUniqueHostnames {
		// 		if strings.HasSuffix(hostname, ".local") {
		// 			tempUniqueHostnames = append(tempUniqueHostnames, "*.local")
		// 			break
		// 		}
		// 	}
		// }

		// dedupe
		seen = make(map[string]bool)
		uniqueHostnames = []string{}
		for _, hostname := range tempUniqueHostnames {
			if hostname != "" {
				if _, ok := seen[hostname]; !ok {
					seen[hostname] = true
					uniqueHostnames = append(uniqueHostnames, hostname)
				}
			}
		}
	}

	return uniqueHostnames
}

// TODO
func GetAllTunnelHostnames() map[string]string {
	config := GetMainConfig()
	tunnels := config.HTTPConfig.ProxyConfig.Routes
	results := map[string]string{}
	
	for _, tunnel := range tunnels {
		if tunnel.TunnelVia != "" && tunnel.TunneledHost != "" {
			results[strings.Split(tunnel.TunneledHost, ":")[0]] = tunnel.TunnelVia
		}
	}

	Debug("Tunnel Hostnames: " + fmt.Sprint(results))

	return results
}

func GetAvailableRAM() uint64 {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}

	// Use total available memory as an approximation
	return vmStat.Available
}

func StringArrayEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for _, value := range a {
		if !StringArrayContains(b, value) {
			return false
		}
	}

	for _, value := range b {
		if !StringArrayContains(a, value) {
			return false
		}
	}

	return true
}

func HasAnyNewItem(after []string, before []string) bool {
	for _, value := range after {
		if !StringArrayContains(before, value) {
			return true
		}
	}
	return false
}

func StringArrayContains(a []string, b string) bool {
	for _, value := range a {
		if value == b {
			return true
		}
	}
	return false
}

func GetServerURL(overwriteHostname string) string {
	ServerURL := ""

	if IsHTTPS {
		ServerURL += "https://"
	} else {
		ServerURL += "http://"
	}

	if overwriteHostname != "" {
		ServerURL += overwriteHostname
	} else {
		ServerURL += MainConfig.HTTPConfig.Hostname
	}
	
	if IsHTTPS && MainConfig.HTTPConfig.HTTPSPort != "443" {
		ServerURL += ":" + MainConfig.HTTPConfig.HTTPSPort
	}
	if !IsHTTPS && MainConfig.HTTPConfig.HTTPPort != "80" {
		ServerURL += ":" + MainConfig.HTTPConfig.HTTPPort
	}

	return ServerURL + "/"
}

func GetServerPort() string {
	if IsHTTPS {
		return MainConfig.HTTPConfig.HTTPSPort
	}
	return MainConfig.HTTPConfig.HTTPPort
}

func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func ImageToBase64(path string) (string, error) {
	imageFile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer imageFile.Close()

	imageData, err := ioutil.ReadAll(imageFile)
	if err != nil {
		return "", err
	}

	encodedData := base64.StdEncoding.EncodeToString(imageData)

	fileExt := strings.ToLower(filepath.Ext(path))
	var mimeType string
	switch fileExt {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".bmp":
		mimeType = "image/bmp"
	default:
		return "", fmt.Errorf("unsupported file format: %s", fileExt)
	}

	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encodedData)
	return dataURI, nil
}

func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func GetCPUUsage() ([]float64) {
	percentages, _ := cpu.Percent(time.Second, false)
	return percentages
}

func GetRAMUsage() (uint64) {
	v, _ := mem.VirtualMemory()
	return v.Used
}

type DiskStatus struct {
	Path       string
	TotalBytes uint64
	UsedBytes  uint64
}

func GetDiskUsage() []DiskStatus {
	partitions, err := disk.Partitions(false)
	if err != nil {
		Error("Error getting disk partitions", err)
		return nil
	}

	var diskStatuses []DiskStatus

	for _, partition := range partitions {
		usageStat, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			Error("Error getting disk usage", err)
			return nil
		}
		diskStatus := DiskStatus{
			Path:       partition.Mountpoint,
			TotalBytes: usageStat.Total,
			UsedBytes:  usageStat.Used,
		}
		diskStatuses = append(diskStatuses, diskStatus)
	}

	return diskStatuses
}

type NetworkStatus struct {
	BytesSent uint64
	BytesRecv  uint64
}

func GetNetworkUsage() NetworkStatus {
	initialStat, err := net.IOCounters(true)
	if err != nil {
		Error("Error getting network usage", err)
		return NetworkStatus{}
	}

	time.Sleep(1 * time.Second)

	finalStat, err := net.IOCounters(true)
	if err != nil {
		Error("Error getting network usage", err)
		return NetworkStatus{}
	}

	res := NetworkStatus{
		BytesSent: 0,
		BytesRecv: 0,
	}
	
	for i := range initialStat {
		res.BytesSent += finalStat[i].BytesSent - initialStat[i].BytesSent
		res.BytesRecv += finalStat[i].BytesRecv - initialStat[i].BytesRecv
	}

	return NetworkStatus{}
}

func DownloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	return string(body), nil
}

func DownloadFileToLocation(path, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func GetClientIP(req *http.Request) string {
	/*ip := req.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = req.RemoteAddr
	}*/
	return req.RemoteAddr
}

func IsDomain(domain string) bool {
	// contains . and at least a letter and no special characters invalid in a domain
	if strings.Contains(domain, ".") && strings.ContainsAny(domain, "abcdefghijklmnopqrstuvwxyz") && !strings.ContainsAny(domain, " !@#$%^&*()+=[]{}\\|;:'\",/<>?") {
		return true
	}
	return false
}

func CheckHostNetwork() {
  IsHostNetwork =	CheckDockerNetworkMode() == "host"
	Log("Cosmos IsHostNetwork: " + strconv.FormatBool(IsHostNetwork))
}

// compareSemver compares two semantic version strings.
// Returns:
//   0 if v1 == v2
//   1 if v1 > v2
//  -1 if v1 < v2
//   error if there's a problem parsing either version string
func CompareSemver(v1, v2 string) (int, error) {
	ver1, err := semver.NewVersion(v1)
	if err != nil {
		Error("compareSemver 1 " + v1, err)
		return 0, err
	}

	ver2, err := semver.NewVersion(v2)
	if err != nil {
		Error("compareSemver 2 " + v2, err)
		return 0, err
	}

	return ver1.Compare(ver2), nil
}


func CheckPassword(nickname, password string) error {
	time.Sleep(time.Duration(rand.Float64()*1)*time.Second)
	
	c, closeDb, errCo := GetEmbeddedCollection(GetRootAppId(), "users")
	defer closeDb()
	if errCo != nil {
			return errCo
	}

	user := User{}

	err3 := c.FindOne(nil, map[string]interface{}{
		"Nickname": nickname,
	}).Decode(&user)

	
	if err3 == mongo.ErrNoDocuments {
		bcrypt.CompareHashAndPassword([]byte("$2a$14$4nzsVwEnR3.jEbMTME7kqeCo4gMgR/Tuk7ivNExvXjr73nKvLgHka"), []byte("dummyPassword"))
		return err3
	} else if err3 != nil {
		bcrypt.CompareHashAndPassword([]byte("$2a$14$4nzsVwEnR3.jEbMTME7kqeCo4gMgR/Tuk7ivNExvXjr73nKvLgHka"), []byte("dummyPassword"))
		return err3
	} else if user.Password == "" {
		return errors.New("User not registered")
	} else {
		err2 := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

		if err2 != nil {
			return err2
		}

		return nil
	}
}

func Values[M ~map[K]V, K comparable, V any](m M) []V {
	r := make([]V, 0, len(m))
	for _, v := range m {
			r = append(r, v)
	}
	return r
}

func Exec(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	Debug("Executing command: " + cmd + " " + strings.Join(args, " "))
	errF := err
	if err != nil {
		errF = errors.New(err.Error() + ": " + string(out))
	}
	return string(out), errF
}

func IsLocalIP(ip string) bool {
	// explicit localhost
	if ip == "localhost" {
		return true
	}
	parsed := osnet.ParseIP(ip)
	// check for loopback or private address space using go std lib
	if parsed.IsLoopback() || parsed.IsPrivate() {
		return true
	}
	// more private IPv4 address ranges
	if ip4 := parsed.To4(); ip4 != nil {
		// https://en.wikipedia.org/wiki/Reserved_IP_addresses
		//     100.64.0.0      -   100.127.255.255 (100.64/10 prefix)
		//     192.0.0.0       -   192.0.0.255     (192.0.0.0/24 prefix)
		return (ip4[0] == 100 && ip4[1]&0x40 == 64) ||
		    (ip4[0] == 192 && ip4[1] == 0 && ip4[2] == 0)
	}
	// Handling cases where IPv6 might be enclosed in brackets
	if strings.HasPrefix(ip, "[") && strings.HasSuffix(ip, "]") {
		return IsLocalIP(strings.TrimSuffix(strings.TrimPrefix(ip, "["), "]"))
	}
	return false
}

func IsConstellationIP(ip string) bool {
	if strings.HasPrefix(ip, "192.168.201.") || strings.HasPrefix(ip, "192.168.202.") {
		return true
	}

	return false 
}

func SplitIP(ipPort string) (string, string) {
	host, port, err := osnet.SplitHostPort(ipPort)
	if err != nil {
			// If there was an error splitting host and port, try parsing as IP only
			if ip := osnet.ParseIP(ipPort); ip != nil {
					// If it's a valid IP, return it with an empty port
					return ip.String(), ""
			}
			// Otherwise, return an empty IP and port (indicating an invalid input)
			return "", ""
	}
	return host, port
}

func ListInterfaces(skipNebula bool) ([]string, error) {
	result := []string{}

	// Get a list of all network interfaces.
	interfaces, err := osnet.Interfaces()
	if err != nil {
		return []string{}, err
	}

	// Iterate over all interfaces.
	for _, iface := range interfaces {
		// skip nebula1 interface 
		if skipNebula && strings.HasPrefix(iface.Name, "nebula") {
			continue
		}

		// skip docker interfaces
		if strings.HasPrefix(iface.Name, "docker") || strings.HasPrefix(iface.Name, "br-") || strings.HasPrefix(iface.Name, "veth") || strings.HasPrefix(iface.Name, "virbr") {
			continue
		}

		result = append(result, iface.Name)
	}

	return result, nil
}

func ListIps(skipNebula bool) ([]string, error) {
	// Get a list of all network interfaces.
	interfaces, err := osnet.Interfaces()
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	// Iterate over all interfaces.
	for _, iface := range interfaces {
			// skip nebula1 interface 
			if skipNebula && strings.HasPrefix(iface.Name, "nebula") {
				continue
			}

			// skip docker interfaces
			if strings.HasPrefix(iface.Name, "docker") || strings.HasPrefix(iface.Name, "br-") || strings.HasPrefix(iface.Name, "veth") || strings.HasPrefix(iface.Name, "virbr") {
				continue
			}
			
			// Get a list of addresses associated with the interface.
			addrs, err := iface.Addrs()
			if err != nil {
				Warn("Error getting addresses for interface " + iface.Name + ": " + err.Error())
					continue
			}

			// Iterate over all addresses.
			for _, addr := range addrs {
					// Check if the address is an IP address and not a mask.
					if ipnet, ok := addr.(*osnet.IPNet); ok && !ipnet.IP.IsLoopback() {
							if ipnet.IP.To4() != nil {
								// if not duplicate
								if !StringArrayContains(result, ipnet.IP.String()) {
									result = append(result, ipnet.IP.String())
								}
							} else if ipnet.IP.To16() != nil {
								// ignore for now
							}
					}
			}
	}

	// TODO sort, local first
	// ...

	return result, nil
}

func RemovePIDFile() {
	if _, err := os.Stat(CONFIGFOLDER + "nebula.pid"); err == nil {
		os.Remove(CONFIGFOLDER + "nebula.pid")
	}
}

func CheckInternet() {
	_, err := http.Get("https://www.google.com")
	if err != nil {
		MajorError("Your server has no internet connection!", err)
	}
}

func GetNextAvailableLocalPort(startPort int) (string, error) {
	for port := startPort; port < 65535; port++ {
			addr := fmt.Sprintf("127.0.0.1:%d", port)
			listener, err := osnet.Listen("tcp", addr)
			if err == nil {
					listener.Close()
					return addr, nil
			}
	}
	return "", fmt.Errorf("no available ports")
}

func GetProxyOIDCredentials(route ProxyRouteConfig, hashSecret bool) *fosite.DefaultClient {
	// Generate deterministic client ID and secret from route info
	config := GetMainConfig()
	HTTPPort := config.HTTPConfig.HTTPPort
	HTTPSPort := config.HTTPConfig.HTTPSPort
	h := sha256.New()
	h.Write([]byte(route.Host))
	h.Write([]byte(config.HTTPConfig.AuthPrivateKey)) // Use existing key as salt
	hash := h.Sum(nil)
	fullhost := route.Host

	// parse host
	parsedURL, err := _url.Parse("http://" + route.Host)
	if err == nil {
			host, port := parsedURL.Hostname(), parsedURL.Port()
			if port == "" {
					if IsHTTPS && HTTPSPort != "443" {
							fullhost = host + ":" + HTTPSPort
					} else if !IsHTTPS && HTTPPort != "80" {
							fullhost = host + ":" + HTTPPort
					}
			}
	} else {
			Error("Error parsing host: " + fullhost, err)
	}

	clientID := route.Name //SanitizeNoSpace(Route.Name) + "_" + hex.EncodeToString(hash[:8])
	plainSecret := hex.EncodeToString(hash[8:24])


	secret := plainSecret
	if hashSecret {
		// Hash the secret with bcrypt
		hashedSecret, err := bcrypt.GenerateFromPassword([]byte(plainSecret), bcrypt.DefaultCost)
		if err != nil {
				Error("Failed to hash client secret", err)
				// Return with unhashed secret as fallback
				hashedSecret = []byte(plainSecret)
		}

		secret = string(hashedSecret)
	}

	callbackURL := fmt.Sprintf("https://%s/cosmos/oauth2/detect-callback", route.Host)
	callbackURLClient := fmt.Sprintf("https://%s/oauth2/callback", route.Host)

	redURls := []string{callbackURL, callbackURLClient}

	if IsHTTPS && config.HTTPConfig.AllowHTTPLocalIPAccess {
		callbackURL2 := fmt.Sprintf("http://%s/cosmos/oauth2/detect-callback", route.Host)
		redURls = append(redURls, callbackURL2)
	}
	
	return &fosite.DefaultClient{
			ID:            clientID,
			Secret: 			 []byte(secret),
			RedirectURIs:  redURls,
			Scopes:        []string{"openid", "email", "profile", "offline"},
			ResponseTypes: []string{"code"},
			GrantTypes:    []string{"authorization_code"},
	}
}

func FindRouteByReqHost(hostname string) (string, *ProxyRouteConfig) {
	config := GetMainConfig()
	HTTPPort := config.HTTPConfig.HTTPPort
	HTTPSPort := config.HTTPConfig.HTTPSPort

	parsedURL, err := _url.Parse("http://" + hostname)
	
	if err != nil {
			Error("Failed to parse URL", err)
			return hostname, nil
	}
	
	host, port := parsedURL.Hostname(), parsedURL.Port()
	
	if port != "" {
			if IsHTTPS && port == HTTPSPort {
					hostname = host
			} else if !IsHTTPS && port == HTTPPort {
					hostname = host 
			}
	}
	
	var proxyRoute *ProxyRouteConfig
	for _, route := range config.HTTPConfig.ProxyConfig.Routes {
			if route.Host == hostname {
					proxyRoute = &route
					break
			}
	}

	return hostname, proxyRoute
}
