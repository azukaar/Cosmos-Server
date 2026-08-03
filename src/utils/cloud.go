package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
	"github.com/golang-jwt/jwt"
	"github.com/shirou/gopsutil/v3/common"
	"github.com/shirou/gopsutil/v3/mem"

	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
)

type FirebaseApiSdk struct {
	BaseURL string
	LValid bool
	ServerToken string
	UserNumber int
	CosmosNodeNumber int
	AgentMode bool
}

var publicKeyPEM = []byte(`
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE8QolLbFdVfU3XPkC01NwsS94bv1W
Ijy+/SYjyHfakFQm7JDhKpbNPC5oc+e4uM6Y9UyC0686toqpTYBSzbgaQw==
-----END PUBLIC KEY-----
`)

func parseECPublicKeyFromPEM(publicKeyPEM []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ECDSA public key")
	}

	return ecdsaPub, nil
}

var FBL *FirebaseApiSdk

func InitFBL() {
	FBL = NewFirebaseApiSdk("https://api.cosmos-cloud.io")
	ProcessLicence()
	if FBL.UserNumber == 0 {
		FBL.UserNumber = 5
	}
	if FBL.CosmosNodeNumber == 0 {
		FBL.CosmosNodeNumber = 1
	}
}

func NewFirebaseApiSdk(baseURL string) *FirebaseApiSdk {
	return &FirebaseApiSdk{BaseURL: baseURL}
}

func (sdk *FirebaseApiSdk) CreateClientLicense(clientID string) (string, error) {
	// Check if token is expired or expiring within 45 days
	if isTokenExpiringWithin(sdk.ServerToken, 45*24*time.Hour) {
		Log("[Cloud] Server token is expired or expiring soon, renewing...")
		ProcessLicence()
		if FBL != nil {
			sdk.ServerToken = FBL.ServerToken
			sdk.LValid = FBL.LValid
			sdk.UserNumber = FBL.UserNumber
			sdk.CosmosNodeNumber = FBL.CosmosNodeNumber
		}
	}

	payload := map[string]string{
		"serverToken":	sdk.ServerToken,
		"clientId":    clientID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(sdk.BaseURL+"/createClientLicense", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create client license: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create client license: %s", string(body))
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return result.Token, nil
}

func (sdk *FirebaseApiSdk) RenewLicense(oldToken string) (string, int, error) {
	if oldToken == "" {
		return "", 0, fmt.Errorf("No server license found")
	}

	payload := map[string]string{
		"oldToken": oldToken,
		"agentMode": fmt.Sprintf("%t", FBL.AgentMode),
	}

	if IsPro() {
		ctx := context.Background()

		if IsInsideContainer {
			if _, err := os.Stat("/mnt/host"); err == nil {
				ctx = context.WithValue(context.Background(),
					common.EnvKey, common.EnvMap{
						common.HostProcEnvKey: "/mnt/host/proc",
						common.HostSysEnvKey:  "/mnt/host/sys",
						common.HostEtcEnvKey:  "/mnt/host/etc",
						common.HostVarEnvKey:  "/mnt/host/var",
						common.HostRunEnvKey:  "/mnt/host/run",
						common.HostDevEnvKey:  "/mnt/host/dev",
						common.HostRootEnvKey: "/mnt/host/",
					},
				)
			}
		}

		memInfo, err := mem.VirtualMemoryWithContext(ctx)
		if err != nil {
			Error("[Cloud] Error fetching RAM for license renewal", err)
		} else {
			payload["ram"] = strconv.FormatUint(memInfo.Total, 10)
		}

		if nbUsers, err := CountUsers(); err != nil {
			Error("[Cloud] Error counting users for license renewal", err)
		} else {
			payload["nbUsers"] = strconv.FormatInt(nbUsers, 10)
		}
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(sdk.BaseURL+"/renewLicense", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", 0, fmt.Errorf("failed to renew license: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", resp.StatusCode, fmt.Errorf("failed to renew license: %s", string(body))
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return result.Token, 0, nil
}

func isTokenOlderThan(token string, duration time.Duration) bool {
	Debug("[Cloud] isTokenOlderThan: checking token age")
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return true
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return true
	}

	iatRaw, exists := claims["iat"]
	if !exists {
		return true
	}

	var iat float64
	switch v := iatRaw.(type) {
	case float64:
		iat = v
	case json.Number:
		iat, err = v.Float64()
		if err != nil {
			return true
		}
	default:
		return true
	}

	issuedAt := time.Unix(int64(iat), 0)
	age := time.Since(issuedAt)
	result := age > duration
	Debug(fmt.Sprintf("[Cloud] isTokenOlderThan: issued=%v age=%v threshold=%v older=%v", issuedAt, age, duration, result))
	return result
}

func GetSubscriptionTypeFromToken(serverToken string) string {
	token, _, err := new(jwt.Parser).ParseUnverified(serverToken, jwt.MapClaims{})
	if err != nil {
		return ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}

	subType, _ := claims["subscriptionType"].(string)
	return subType
}

// IsTokenValidForProduct returns false when running the Pro build against a
// non-Pro subscription token — a Pro-only binary must refuse regular licences.
func IsTokenValidForProduct(serverToken string) bool {
	if !IsPro() {
		return true
	}
	return GetSubscriptionTypeFromToken(serverToken) == "pro"
}

func GetNumberUsersFromToken(serverToken string) (int, int) {
	Debug("[Cloud] GetNumberUsersFromToken")

	// decode the token
	token, _, err := new(jwt.Parser).ParseUnverified(serverToken, jwt.MapClaims{})
	if err != nil {
		Error("[Cloud] Could not parse server token", err)
		return 5, 1
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		Error("[Cloud] Could not parse server token", err)
		return 5, 1
	}

	nbUser := 20
	nbNodes := 1
	
	// get the number of users
	userNumber, ok := claims["nbUsers"].(float64)
	if !ok {
		Log("[Cloud] Could not get number of users from token, defaulting to 20")
	} else {
		nbUser = int(userNumber)
	}

	cosmosNodeNumber, ok := claims["nbCosmosNodes"].(float64)
	if !ok {
		Log("[Cloud] Could not get number of cosmos nodes from token, defaulting to 1")
	} else {
		nbNodes = int(cosmosNodeNumber)
	}

	Log("[Cloud] Number of users: " + fmt.Sprintf("%d", int(userNumber)))

	if int(nbUser) < 20 {
		nbUser = 20
	}

	return int(nbUser), nbNodes
}


func GetNumberUsers() int {
	// Pro is unlimited; non-Pro keeps the licensed seat count (or the
	// 5-seat free tier when no licence is loaded).
	if IsPro() {
		return math.MaxInt32
	}
	if FBL.LValid {
		return FBL.UserNumber
	}
	return 5
}

func GetNumberCosmosNode() int {
	// Pro is unlimited; non-Pro keeps the licensed node count (or the
	// 1-node free tier when no licence is loaded).
	if IsPro() {
		return math.MaxInt32
	}
	if FBL.LValid {
		return FBL.CosmosNodeNumber
	}
	return 1
}

func isTokenExpiringWithin(token string, duration time.Duration) bool {
	if token == "" {
		return true
	}

	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		Error("[Cloud] Could not parse token for expiration check", err)
		return true
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		Error("[Cloud] Could not parse token claims for expiration check", nil)
		return true
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		Log("[Cloud] Token does not have expiration claim, assuming expired")
		return true
	}

	expirationTime := time.Unix(int64(exp), 0)
	return time.Now().Add(duration).After(expirationTime)
}

func ProcessLicence() {
	if FBL == nil {
		return
	}

	config := ReadConfigFromFile()
	licence := config.Licence
	serverToken := config.ServerToken
	isAgent := config.AgentMode

	FBL.AgentMode = isAgent

	publicKey, err := parseECPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		MajorError("[Cloud] Failed to parse public key", err)
		return
	}

	if licence == "" {
		// licence removed by the user: flush any leftover server token
		if serverToken != "" {
			Log("[Cloud] No licence configured, flushing server token")
			config.ServerToken = ""
			SetBaseMainConfig(config)
		}
		FBL.ServerToken = ""
		FBL.LValid = false
		FBL.UserNumber = 5
		FBL.CosmosNodeNumber = 1
		return
	}

	// Verify token signature if we have one
	if serverToken != "" {
		token, err := jwt.Parse(serverToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				serverToken = ""
				Error("[Cloud] Server Token is not trustworthy", nil)
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		})

		if err == nil && token.Valid && !isTokenOlderThan(serverToken, 24*time.Hour) {
			if !IsTokenValidForProduct(serverToken) {
				MajorError("[Cloud] Licence subscription type is not valid for this product (Pro required)", nil)
				FBL.LValid = false
				return
			}
			Debug("[Cloud] Existing server token is valid and not too old, using it")
			FBL.ServerToken = serverToken
			FBL.LValid = true
			FBL.UserNumber, FBL.CosmosNodeNumber = GetNumberUsersFromToken(serverToken)
			InitPremiumFeatures()
			return
		}
	}

	// Attempt renewal
	newToken, _, err := FBL.RenewLicense(licence)
	if err != nil || newToken == "" {
		MajorError("[Cloud] Could not renew token", err)

		// Offline fallback: use existing token if still valid
		if serverToken != "" && !isTokenExpiringWithin(serverToken, 0) {
			if !IsTokenValidForProduct(serverToken) {
				MajorError("[Cloud] Licence subscription type is not valid for this product (Pro required)", nil)
				FBL.LValid = false
				return
			}
			Log("[Cloud] Keeping existing token (offline)")
			FBL.ServerToken = serverToken
			FBL.LValid = true
			FBL.UserNumber, FBL.CosmosNodeNumber = GetNumberUsersFromToken(serverToken)
			InitPremiumFeatures()
			return
		}

		MajorError("[Cloud] Token expired and renewal failed", err)
		FBL.LValid = false
		return
	}

	if !IsTokenValidForProduct(newToken) {
		MajorError("[Cloud] Licence subscription type is not valid for this product (Pro required)", nil)
		FBL.LValid = false
		return
	}

	config.ServerToken = newToken
	SetBaseMainConfig(config)
	FBL.ServerToken = newToken
	FBL.LValid = true
	FBL.UserNumber, FBL.CosmosNodeNumber = GetNumberUsersFromToken(newToken)
	InitPremiumFeatures()
	Log("[Cloud] Token renewed successfully")
}