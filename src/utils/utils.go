package utils

import (
	"encoding/json"
	"math/rand"
	"regexp"
	"net/http"
	"encoding/base64"
	"os"
	"strconv"
	"strings"
	"io/ioutil"
	"fmt"
	"sync"
	"time"
	"path/filepath"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
	"golang.org/x/net/publicsuffix"
)

var ConfigLock sync.Mutex

var BaseMainConfig Config
var MainConfig Config
var IsHTTPS = false
var NewVersionAvailable = false

var NeedsRestart = false

var UpdateAvailable = map[string]bool{}

var RestartHTTPServer func()
var ReBootstrapContainer func(string) error

var LetsEncryptErrors = []string{}

var CONFIGFOLDER = "/config/"

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
		ProxyConfig: ProxyConfig{
			Routes: []ProxyRouteConfig{},
		},
	},
	DockerConfig: DockerConfig{
		DefaultDataPath: "/usr",
	},
  MarketConfig: MarketConfig{
    Sources: []MarketSource{
		},
	},
  ConstellationConfig: ConstellationConfig{
    Enabled: false,
		DNS: true,
		DNSFallback: "8.8.8.8:53",
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
	if os.Getenv("COSMOS_CONFIG_FOLDER") != "" {
		Log("Overwriting config folder with " + os.Getenv("COSMOS_CONFIG_FOLDER"))
		CONFIGFOLDER = os.Getenv("COSMOS_CONFIG_FOLDER")
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
	os.Exit(0)
}

func LetsEncryptValidOnly(hostnames []string, acceptWildcard bool) []string {
	wrongPattern := `^(localhost|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|.*\.local)$`
	re, _ := regexp.Compile(wrongPattern)

	var validDomains []string
	for _, domain := range hostnames {
		if !re.MatchString(domain) && (acceptWildcard || !strings.Contains(domain, "*")) && !strings.Contains(domain, " ") && !strings.Contains(domain, ",") {
			validDomains = append(validDomains, domain)
		} else {
			Error("Invalid domain found in URLs: " + domain + " it was removed from the certificate to not break Let's Encrypt", nil)
		}
	}

	return validDomains
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

	proxies := GetMainConfig().HTTPConfig.ProxyConfig.Routes
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

	if applyWildCard && MainConfig.HTTPConfig.UseWildcardCertificate {
		bareMainHostname, _ := publicsuffix.EffectiveTLDPlusOne(mainHostname)

		Debug("bareMainHostname: " + bareMainHostname)

		filteredHostnames := []string{
			bareMainHostname,
			"*." + bareMainHostname,
		}

		if(OverrideWildcardDomains != "") {
			filteredHostnames = strings.Split(OverrideWildcardDomains, ",")
		}

		for _, hostname := range uniqueHostnames {
			if hostname != bareMainHostname && !strings.HasSuffix(hostname, "." + bareMainHostname) {
				filteredHostnames = append(filteredHostnames, hostname)
			}
		}

		uniqueHostnames = filteredHostnames
	}

	Debug("Hostnames are " + strings.Join(uniqueHostnames, ", "))
	return uniqueHostnames
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

func GetServerURL() string {
	ServerURL := ""

	if IsHTTPS {
		ServerURL += "https://"
	} else {
		ServerURL += "http://"
	}

	ServerURL += MainConfig.HTTPConfig.Hostname

	if IsHTTPS && MainConfig.HTTPConfig.HTTPSPort != "443" {
		ServerURL += ":" + MainConfig.HTTPConfig.HTTPSPort
	}
	if !IsHTTPS && MainConfig.HTTPConfig.HTTPPort != "80" {
		ServerURL += ":" + MainConfig.HTTPConfig.HTTPPort
	}

	return ServerURL + "/"
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

func GetClientIP(req *http.Request) string {
	/*ip := req.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = req.RemoteAddr
	}*/
	return req.RemoteAddr
}