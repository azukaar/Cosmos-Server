package utils

import (
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"time"
	"crypto/rand"
	"crypto"
	"crypto/rsa"
	"crypto/ed25519"
	"crypto/x509/pkix"
	"crypto/ecdsa"
	"crypto/elliptic"
	"os"
	"strings"
	"net"
	
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/go-acme/lego/v4/providers/dns"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

type CAConfig struct {
	Certificate *x509.Certificate
	PrivateKey  *rsa.PrivateKey
}

// generateCA creates a new CA certificate if it doesn't exist
func generateCA() (*CAConfig, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	id := GenerateRandomString(5)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Cosmos Personal Server CA " + id},
			CommonName:   "Cosmos Root CA " + id,
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(100 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, err
	}

	return &CAConfig{
		Certificate: cert,
		PrivateKey:  privateKey,
	}, nil
}

func loadCA() (*CAConfig, error) {
	config := GetMainConfig()
	CACert := config.HTTPConfig.CACert
	PAPrivateKey := config.HTTPConfig.CAPrivateKey
	
	if CACert == "" || PAPrivateKey == "" {
		return nil, nil
	}

	certBlock, _ := pem.Decode([]byte(CACert))

	if certBlock == nil {
		return nil, nil
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)

	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode([]byte(PAPrivateKey))

	if keyBlock == nil {
		return nil, nil
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)

	if err != nil {
		return nil, err
	}

	return &CAConfig{
		Certificate: cert,
		PrivateKey:  key,
	}, nil
}

func saveCA(ca *CAConfig) error {
	baseMainConfig := GetBaseMainConfig()

	baseMainConfig.HTTPConfig.CACert = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Certificate.Raw}))
	baseMainConfig.HTTPConfig.CAPrivateKey = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(ca.PrivateKey)}))

	SetBaseMainConfig(baseMainConfig)
	Log("Saved new CA certificate and private key")

	return nil
}

func GenerateRSAWebCertificates(domains []string) (string, string, error) {	
	// Try to load existing CA or generate new one
	ca, err := loadCA()
	if err != nil || ca == nil || ca.Certificate == nil || ca.PrivateKey == nil {
		// If CA doesn't exist, generate and save it
		ca, err = generateCA()
		if err != nil {
			return "", "", err
		}
		if err := saveCA(ca); err != nil {
			return "", "", err
		}
	}

	// Generate web server private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Cosmos Personal Server"},
			CommonName:   domains[0], // Use first domain as CN
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().AddDate(0, 0, 364),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:             domains,
		SubjectKeyId:         []byte{1, 2, 3, 4, 6},
		AuthorityKeyId:       ca.Certificate.SubjectKeyId,
	}

	// Create certificate signed by CA
	cert, err := x509.CreateCertificate(rand.Reader, &template, ca.Certificate, &privateKey.PublicKey, ca.PrivateKey)
	if err != nil {
		return "", "", err
	}

	// Return certificate and private key PEM encoded
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return string(certPEM), string(keyPEM), nil
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

type CertUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}
func (u *CertUser) GetEmail() string {
	return u.Email
}

func (u CertUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *CertUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}


func DoLetsEncrypt() (string, string) {
	config := GetMainConfig()
	LetsEncryptErrors = []string{}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		Error("LETSENCRYPT_ECDSA", err)
		LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
		return "", ""
	}

	myUser := CertUser{
		Email: config.HTTPConfig.SSLEmail,
		key:   privateKey,
	}

	certConfig := lego.NewConfig(&myUser)
	
	domains := GetAllHostnames(true, true)

	if os.Getenv("ACME_STAGING") == "true" {
		certConfig.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	} else {
		certConfig.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	}

	certConfig.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(certConfig)
	if err != nil {
		Error("LETSENCRYPT_NEW", err)
		LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
		return "", ""
	}

	if config.HTTPConfig.DNSChallengeProvider != "" {
		provider, err := dns.NewDNSChallengeProviderByName(config.HTTPConfig.DNSChallengeProvider)
		if err != nil {
			Error("LETSENCRYPT_DNS", err)
			LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
			return "", ""
		}

		// use the authoritative nameservers
		resolvers := []string{}

		if config.HTTPConfig.DNSChallengeResolvers == "" {
			processedDomains := map[string]bool{}

			for _, domain := range domains {
				levels := strings.Split(domain, ".")
				if len(levels) >= 2 {
					tld := strings.Join(levels[len(levels)-2:], ".")
					if processedDomains[tld] {
						continue
					}

					nameservers, err := net.LookupNS(tld)
					
					if err != nil {
						continue
					}
					
					for _, ns := range nameservers {
						resolvers = append(resolvers, ns.Host)
					}

					processedDomains[tld] = true
				}
			}

			// append the default resolvers
			resolvers = append(resolvers, "8.8.8.8", "1.1.1.1")
		} else {
			resolvers = strings.Split(config.HTTPConfig.DNSChallengeResolvers, ",")
			// trim
			for i, r := range resolvers {
				resolvers[i] = strings.TrimSpace(r)
			}
		}

		var PropagationWait = time.Duration(config.HTTPConfig.DNSChallengePropagationWait) * time.Second

		err = client.Challenge.SetDNS01Provider(provider,
			dns01.AddRecursiveNameservers(resolvers),
			dns01.CondOption(config.HTTPConfig.DisablePropagationChecks,
				dns01.PropagationWait(PropagationWait, true)),
		)
	} else {
		err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", config.HTTPConfig.HTTPPort))
		if err != nil {
			Error("LETSENCRYPT_HTTP01", err)
			LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
			return "", ""
		}

		err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", config.HTTPConfig.HTTPSPort))
		if err != nil {
			Error("LETSENCRYPT_TLS01", err)
			LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
			return "", ""
		}
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		Error("LETSENCRYPT_REGISTER", err)
		LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
		return "", ""
	}
	myUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: LetsEncryptValidOnly(domains, config.HTTPConfig.DNSChallengeProvider != ""),
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		Error("LETSENCRYPT_OBTAIN", err)
		LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
		return "", ""
	}

	// return cert and key
	return string(certificates.Certificate), string(certificates.PrivateKey)
}

// You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}