package cron

import (
	"net/http"
	"github.com/gorilla/websocket"
	"context"
	"log"

	"github.com/azukaar/cosmos-server/src/utils"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // adjust the origin checking to fit your requirements
	},
}

var clients = make(map[*websocket.Conn]bool) // connected clients

// listenJobs handles new WebSocket requests from clients
func listenJobs(w http.ResponseWriter, r *http.Request) {
	if utils.AdminOnly(w, r) != nil {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
			log.Println("Error during connection upgrade:", err)
			return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
			for {
					select {
					case <-ctx.Done():
							// Context canceled, client disconnected
							return
					default:
							// Your regular processing logic
					}

					// Example of reading from the connection
					_, message, err := conn.ReadMessage()
					if err != nil {
							log.Println("Error during message reading:", err)
							cancel() // Cancel the context on error
							break
					}
					log.Printf("Received: %s", message)
			}
	}()

	// The rest of your handler, like sending messages to the client
	// Make sure to check ctx.Done() regularly
}

// triggerJobUpdated sends an update to all connected WebSocket clients
func triggerJobUpdated(updateType string, jobName string, args ...string) {
	for client := range clients {
		err := client.WriteJSON(map[string]interface{}{
			"channel": "jobUpdated",
			"type": "updateType",
			"jobName": jobName,
			"args": args,
		})
		if err != nil {
			log.Printf("error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}
