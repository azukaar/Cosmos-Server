package authorizationserver

import (
	"encoding/json"
	"net/http"
	"encoding/base64"
	"math/big"
	"crypto/rsa"
	
	"github.com/azukaar/cosmos-server/src/utils"
)

type JsonWebKey struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JsonWebKeySet struct {
	Keys []JsonWebKey `json:"keys"`
}

func jwksEndpoint(rw http.ResponseWriter, req *http.Request) {
	hostname := utils.GetMainConfig().HTTPConfig.Hostname

	if utils.IsHTTPS {
		hostname = "https://" + hostname
	} else {
		hostname = "http://" + hostname
	}

	// RSA Public Key from rsa.GenerateKey
	publicKey := AuthPrivateKey.Public().(*rsa.PublicKey)

	json.NewEncoder(rw).Encode(&JsonWebKeySet{
		Keys: []JsonWebKey{
			{
				Alg: "RS256",
				Kid: "1",
				Kty: "RSA",
				Use: "sig",
				N:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
				E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
			},
		},
	})
}
