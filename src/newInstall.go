package main

import (
	"net/http"
	"encoding/json"
	"time"
	"os"
	"golang.org/x/crypto/bcrypt"	

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/docker"
)

func waitForDB() {
	time.Sleep(1 * time.Second)
	err := utils.DB()
	if err != nil {
		utils.Warn("DB Not ready yet")
		waitForDB()
	}
}

type NewInstallJSON struct {
	MongoDBMode string `json:"mongodbMode"`
	MongoDB string `json:"mongodb"`
	HTTPSCertificateMode string `json:"httpsCertificateMode"`
	TLSCert string `json:"tlsCert"`
	TLSKey string `json:"tlsKey"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
	Email string `json:"omitempty,email"`
	Hostname string `json:"hostname"`
	Step string `json:"step"`
	SSLEmail string `json:"sslEmail",validate:"omitempty,email"`
	UseWildcardCertificate bool `json:"useWildcardCertificate",validate:"omitempty"`
	DNSChallengeProvider string `json:"dnsChallengeProvider",validate:"omitempty"`
}

type AdminJSON struct {
	Nickname string `validate:"required,min=3,max=32,alphanum"`
	Password string `validate:"required,min=9,max=128,containsany=!@#$%^&*()_+,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789"`
}

func NewInstallRoute(w http.ResponseWriter, req *http.Request) {
	if !utils.GetMainConfig().NewInstall {
		utils.Error("Status: not a new New install", nil)
		utils.HTTPError(w, "New install", http.StatusForbidden, "NI001")
		return
	}

	if(req.Method == "POST") {
		var request NewInstallJSON
		err1 := json.NewDecoder(req.Body).Decode(&request)
		if err1 != nil {
			utils.Error("NewInstall: Invalid User Request", err1)
			utils.HTTPError(w, "New Install: Invalid User Request" + err1.Error(), 
				http.StatusInternalServerError, "NI001")
			return 
		}

		errV := utils.Validate.Struct(request)
		if errV != nil {
			utils.Error("NewInstall: Invalid User Request", errV)
			utils.HTTPError(w, "New Install: Invalid User Request " + errV.Error(),
				http.StatusInternalServerError, "NI001")
			return 
		}

		newConfig := utils.GetBaseMainConfig()

		if(request.Step == "2") {
			utils.Log("NewInstall: Step Database")
			// User Management & Mongo DB
			if(request.MongoDBMode == "DisableUserManagement") {
				utils.Log("NewInstall: Disable User Management")
				newConfig.DisableUserManagement = true
				utils.SaveConfigTofile(newConfig)
				utils.LoadBaseMainConfig(newConfig)
			} else if (request.MongoDBMode == "Provided") {
				utils.Log("NewInstall: DB Provided")
				newConfig.DisableUserManagement = false
				newConfig.MongoDB = request.MongoDB
				utils.SaveConfigTofile(newConfig)
				utils.LoadBaseMainConfig(newConfig)
			} else if (request.MongoDBMode == "Create") {
				utils.Log("NewInstall: Create DB")
				newConfig.DisableUserManagement = false

				strco, err := docker.NewDB(w, req)
				if err != nil {
					utils.Error("NewInstall: Error creating MongoDB", err)
					return 
				}			
						
				newConfig.MongoDB = strco
				utils.SaveConfigTofile(newConfig)
				utils.LoadBaseMainConfig(newConfig)
				utils.Log("NewInstall: MongoDB created, waiting for it to be ready")
				waitForDB()
			} else {
				utils.Log("NewInstall: Invalid MongoDBMode")
				utils.Error("NewInstall: Invalid MongoDBMode", nil)
				utils.HTTPError(w, "New Install: Invalid MongoDBMode",
					http.StatusInternalServerError, "NI001")
				return 
			}
		} else if (request.Step == "3") {
			// HTTPS Certificate Mode & Certs & Let's Encrypt
			newConfig.HTTPConfig.HTTPSCertificateMode = request.HTTPSCertificateMode
			newConfig.HTTPConfig.SSLEmail = request.SSLEmail
			newConfig.HTTPConfig.UseWildcardCertificate = request.UseWildcardCertificate
			newConfig.HTTPConfig.DNSChallengeProvider = request.DNSChallengeProvider
			newConfig.HTTPConfig.TLSCert = request.TLSCert
			newConfig.HTTPConfig.TLSKey = request.TLSKey

			// Hostname
			newConfig.HTTPConfig.Hostname = request.Hostname

			utils.SaveConfigTofile(newConfig)
			utils.LoadBaseMainConfig(newConfig)
		} else if (request.Step == "4") {
			
			adminObj := AdminJSON{
				Nickname: request.Nickname,
				Password: request.Password,
			}		

			errV2 := utils.Validate.Struct(adminObj)
			if errV2 != nil {
				utils.Error("NewInstall: Invalid User Request", errV2)
				utils.HTTPError(w, errV2.Error(), http.StatusInternalServerError, "UL001")
				return
			}

			// Admin User
			c, errCo := utils.GetCollection(utils.GetRootAppId(), "users")
			if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return
			}

			nickname := utils.Sanitize(request.Nickname)
			hashedPassword, err2 := bcrypt.GenerateFromPassword([]byte(request.Password), 14)

			if err2 != nil {
				utils.Error("NewInstall: Error hashing password", err2)
				utils.HTTPError(w, "New Install: Error hashing password " + err2.Error(),
					http.StatusInternalServerError, "NI001")
				return
			}

			// pre-remove every users
			_, err4 := c.DeleteMany(nil, map[string]interface{}{})
			if err4 != nil {
				utils.Error("NewInstall: Error deleting users", err4)
				utils.HTTPError(w, "New Install: Error deleting users " + err4.Error(),
					http.StatusInternalServerError, "NI001")
				return
			}

			_, err3 := c.InsertOne(nil, map[string]interface{}{
				"Nickname": nickname,
				"Email": request.Email,
				"Password": hashedPassword,
				"Role": utils.ADMIN,
				"PasswordCycle": 0,
				"CreatedAt": time.Now(),
				"RegisteredAt": time.Now(),
			})

			if err3 != nil {
				utils.Error("NewInstall: Error creating admin user", err3)
				utils.HTTPError(w, "New Install: Error creating admin user " + err3.Error(),
					http.StatusInternalServerError, "NI001")
				return
			}
		} else if (request.Step == "5") {
			newConfig.NewInstall = false
			utils.SaveConfigTofile(newConfig)
			os.Exit(0)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
		})
	} else {
		utils.Error("UserList: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}