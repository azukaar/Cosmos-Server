package utils

import (
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"time"
	"crypto/rand"
	"crypto/rsa"
	"crypto/ed25519"
	"crypto/x509/pkix"
	"encoding/asn1"
)

func GenerateRSAWebCertificates() (string, string) {
	// generate self signed certificate

	Log("Generating RSA Web Certificates for " + GetMainConfig().HTTPConfig.Hostname)

	// generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		Fatal("Generating RSA Key", err)
	}

	// generate public key
	publicKey := &privateKey.PublicKey

	// random SerialNumber
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		Fatal("Generating Serial Number", err)
	}

	// generate certificate
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Cosmos Personal Server"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,

		DNSNames: []string{GetMainConfig().HTTPConfig.Hostname},

		// IPAddresses: []net.IP{},

		SubjectKeyId: []byte{1, 2, 3, 4, 6},

		AuthorityKeyId: []byte{1, 2, 3, 4, 5},

		PermittedDNSDomainsCritical: false,

		PermittedDNSDomains: []string{GetMainConfig().HTTPConfig.Hostname},

		// PermittedIPRanges: ,

		ExcludedDNSDomains: []string{},

		// ExcludedIPRanges:,

		PermittedURIDomains: []string{},

		ExcludedURIDomains: []string{},

		CRLDistributionPoints: []string{},

		PolicyIdentifiers: []asn1.ObjectIdentifier{},

	}

	// create certificate

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey, privateKey)
	if err != nil {
		Fatal("Creating RSA Key", err)
	}

	// return private , and public key

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})), string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}))
}

func GenerateEd25519Certificates() (string, string) {
	// generate self signed certificate

	// generate private key
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		Fatal("Generating ed25519 Key", err)
	}

	bpriv, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		Fatal("Generating ed25519 private Key", err)
	}
	
	bpub, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		Fatal("Generating ed25519 public Key", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: bpub})), string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: bpriv}))
}