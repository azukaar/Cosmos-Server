package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/golang-jwt/jwt"
	
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
)

type FirebaseApiSdk struct {
	BaseURL string
	LValid bool
	ServerToken string
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
	FBL = _InitFBL()
}

func _InitFBL() *FirebaseApiSdk {
	res := NewFirebaseApiSdk("https://api.cosmos-cloud.io")
	config := ReadConfigFromFile()
	Licence := config.Licence
	ServerToken := config.ServerToken

	publicKey, err := parseECPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		Error("Failed to parse public key", err)
		return res
	}

	if Licence == "" {
		return res
	}
	
	_, statuscode1, err1 := res.RenewLicense(ServerToken)

	if err1 != nil {
		Error("Could not validate server token", err1)

		newToken, statuscode2, err2 := res.RenewLicense(Licence)
		if err2 != nil || newToken == "" {

			MajorError("Could not renew server token, check internet connection", err2)

			if (ServerToken == "") {
				MajorError("No server token. And could not get one.", err2)
				return res
			}

			if statuscode1 == 0 && statuscode2 == 0 {
				token, err := jwt.Parse(ServerToken, func(token *jwt.Token) (interface{}, error) {
					// Verify the signing method
					if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}
					return publicKey, nil
				})

				if (err != nil) || !token.Valid {
					MajorError("Invalid Server license.", err)
					return res
				}

				res.LValid = true
				res.ServerToken = ServerToken
				return res
			} else {
				MajorError("Invalid license please check your original license email.", err)
				return res
			}
		}

		// save the new token
		config.ServerToken = newToken

		SetBaseMainConfig(config)
		// constellation.RestartNebula()

		res.ServerToken = newToken
		res.LValid = true
		return res
	} else {
		res.ServerToken = ServerToken
		res.LValid = true
		return res
	}
}

func NewFirebaseApiSdk(baseURL string) *FirebaseApiSdk {
	return &FirebaseApiSdk{BaseURL: baseURL}
}

func (sdk *FirebaseApiSdk) CreateClientLicense(clientID string) (string, error) {
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