package utils

import (
	"os"
	"net/http"
	"encoding/json"
  "strconv"
	"strings"
	"math/rand"
)

var BaseMainConfig Config
var MainConfig Config
var IsHTTPS = false

var DefaultConfig = Config{
	LoggingLevel: "INFO",
	HTTPConfig: HTTPConfig{
		GenerateMissingTLSCert: true,
		GenerateMissingAuthCert: true,
		HTTPPort: "80",
		HTTPSPort: "443",
		Hostname: "0.0.0.0",
		ProxyConfig: ProxyConfig{
			Routes: []ProxyRouteConfig{},
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
	Status string `json:"status"`
	Message string `json:"message"`
	Code string `json:"code"`
}

func HTTPError(w http.ResponseWriter, message string, code int, userCode string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(HTTPErrorResult{
		Status: "error",
		Message: message,
		Code: userCode,
	})
	Error("HTTP Request returned Error " + strconv.Itoa(code) + " : " + message, nil)
}

func SetBaseMainConfig(config Config){
	BaseMainConfig = config
	SaveConfigTofile(config)
}

func LoadBaseMainConfig(config Config){
	BaseMainConfig = config
	MainConfig = config

	// use ENV to overwrite configs

	if os.Getenv("HTTP_PORT") != "" {
		MainConfig.HTTPConfig.HTTPPort = os.Getenv("HTTP_PORT")
	}
	if os.Getenv("HTTPS_PORT") != "" {
		MainConfig.HTTPConfig.HTTPSPort = os.Getenv("HTTPS_PORT")
	}
	if os.Getenv("HOSTNAME") != "" {
		MainConfig.HTTPConfig.Hostname = os.Getenv("HOSTNAME")
	}
	if os.Getenv("GENERATE_MISSING_TLS_CERT") != "" {
		MainConfig.HTTPConfig.GenerateMissingTLSCert = os.Getenv("GENERATE_MISSING_TLS_CERT") == "true"
	}
	if os.Getenv("GENERATE_MISSING_AUTH_CERT") != "" {
		MainConfig.HTTPConfig.GenerateMissingAuthCert = os.Getenv("GENERATE_MISSING_AUTH_CERT") == "true"
	}
	if os.Getenv("TLS_CERT") != "" {
		MainConfig.HTTPConfig.TLSCert = os.Getenv("TLS_CERT")
	}
	if os.Getenv("TLS_KEY") != "" {
		MainConfig.HTTPConfig.TLSKey = os.Getenv("TLS_KEY")
	}
	if os.Getenv("AUTH_PRIV_KEY") != "" {
		MainConfig.HTTPConfig.AuthPrivateKey = os.Getenv("AUTH_PRIVATE_KEY")
	}
	if os.Getenv("AUTH_PUBLIC_KEY") != "" {
		MainConfig.HTTPConfig.AuthPublicKey = os.Getenv("AUTH_PUBLIC_KEY")
	}
	if os.Getenv("LOG_LEVEL") != "" {
		MainConfig.LoggingLevel = (LoggingLevel)(os.Getenv("LOG_LEVEL"))
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

func GetConfigFileName() string {
	configFile := os.Getenv("CONFIG_FILE")
	
	if configFile == "" {
		configFile = "/config/cosmos.config.json"
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

	file, err := os.OpenFile(configFile, os.O_WRONLY, os.ModePerm)
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

	Log("Config file saved.");
}

func RestartServer() {
	Log("Restarting server...")
	os.Exit(0)
}