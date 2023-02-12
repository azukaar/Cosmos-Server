package user

import (
	"net/http"
	"log"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	"../utils" 
)

func UserResendInviteLink(w http.ResponseWriter, req *http.Request) {
	utils.SetHeaders(w)

	if(req.Method == "POST") {
		id := req.FormValue("id")

		c := utils.GetCollection(utils.GetRootAppId(), "users")

		user := utils.User{}

		err := c.FindOne(nil, map[string]interface{}{
			"_id": id,
		}).Decode(&user)

		if err == mongo.ErrNoDocuments {
			log.Println("UserResend: User not found")
			http.Error(w, "User Resend Invite Error", http.StatusNotFound)
		} else if err != nil {
			log.Println("UserResend: Error while finding user")
			http.Error(w, "User Resend Invite Error", http.StatusInternalServerError)
		} else {
			RegisterKeyExp := time.Now().Add(time.Hour * 24 * 7)

			_, err := c.UpdateOne(nil, map[string]interface{}{
				"_id": id,
			}, map[string]interface{}{
				"$set": map[string]interface{}{
					"RegisterKeyExp": RegisterKeyExp,
				},
			})

			if err != nil {
				log.Println("UserResend: Error while updating user")
				http.Error(w, "User Resend Invite Error", http.StatusInternalServerError)
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"Status": "OK",
				"RegisterKey": user.RegisterKey,
				"RegisterKeyExp": RegisterKeyExp,
			})
		}
	} else {
		log.Println("UserResend: Method not allowed" + req.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}