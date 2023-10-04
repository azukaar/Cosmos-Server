package constellation

// import (
// 	"net/http"
// 	"encoding/json"
// 	"math/rand"
// 	"time"
// 	"net"
	
// 	"github.com/azukaar/cosmos-server/src/utils" 
// )

// func DeviceConfig(w http.ResponseWriter, req *http.Request) {
// 	time.Sleep(time.Duration(rand.Float64()*2)*time.Second)

// 	if(req.Method == "GET") {

// 		ip, _, err := net.SplitHostPort(req.RemoteAddr)
// 		if err != nil {
// 			http.Error(w, "Invalid request", http.StatusBadRequest)
// 			return
// 		}

// 		// get authorization header
// 		auth := req.Header.Get("Authorization")
// 		if auth == "" {
// 			http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 			return
// 		}

// 		// remove "Bearer " from auth header
// 		auth = strings.Replace(auth, "Bearer ", "", 1)

		

// 	} else {
// 		utils.Error("DeviceConfig: Method not allowed" + req.Method, nil)
// 		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
// 		return
// 	}
// }