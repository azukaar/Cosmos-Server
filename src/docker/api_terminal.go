package docker

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/docker/docker/api/types"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
	"encoding/json"

	"github.com/azukaar/cosmos-server/src/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func splitIntoChunks(input string, chunkSize int) [][]LogOutput {
	lines := strings.Split(input, "\n")
	var chunks [][]LogOutput

	for i := 0; i < len(lines); i += chunkSize {
			end := i + chunkSize

			// Avoid going over the end of the array
			if end > len(lines) {
					end = len(lines)
			}

			var chunk []LogOutput
			for j := i; j < end; j++ {
				chunk = append(chunk, ParseDockerLogHeader(([]byte)(
					lines[j],
				)))
			}
			chunks = append(chunks, chunk)
	}

	return chunks
}

func TerminalRoute(w http.ResponseWriter, r *http.Request) {
	if utils.AdminOnly(w, r) != nil {
		return
	}
	utils.Log("Attempting to attach container")

	upgrader.ReadBufferSize = 1024 * 4  // Increase the buffer size as needed
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.Error("Failed to set websocket upgrade: ", err)
		http.Error(w, "Failed to set websocket upgrade: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer ws.Close()

	ctx := context.Background()
	errD := Connect()
	if errD != nil {
		utils.Error("ManageContainer", errD)
		utils.HTTPError(w, "Internal server error: " + errD.Error(), http.StatusInternalServerError, "DS002")
		return
	}

	vars := mux.Vars(r)
	containerID := utils.SanitizeSafe(vars["containerId"])

	if containerID == "" {
		utils.Error("containerID is required: ", nil)
		http.Error(w, "containerID is required", http.StatusBadRequest)
		return
	}
	
	action := utils.Sanitize(vars["action"])

	utils.Log("Attaching container " + containerID + " to websocket")

	var resp types.HijackedResponse

	if action == "new" {
		execConfig := types.ExecConfig{
			Tty:    true,
			AttachStdin: true,
			AttachStdout: true,
			AttachStderr: true,
			Cmd: []string{"/bin/sh"},
		}

		execStart := types.ExecStartCheck{
			Tty: true,
		}
	
		execResp, errExec := DockerClient.ContainerExecCreate(ctx, containerID, execConfig)
		if errExec != nil {
			utils.Error("ContainerExecCreate failed: ", errExec)
			http.Error(w, "ContainerExecCreate failed: "+errExec.Error(), http.StatusInternalServerError)
			return
		}
	
		resp, err = DockerClient.ContainerExecAttach(ctx, execResp.ID, execStart)
		if err != nil {
			utils.Error("ContainerExecAttach failed: ", err)
			http.Error(w, "ContainerExecAttach failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
	
		utils.Log("Created new shell and attached to it in container " + containerID)
	} else {
		options := types.ContainerAttachOptions{
			Stream: true,
			Stdin:  true,
			Stdout: true,
			Stderr: true,
		}
	
		// Attach to the container
		resp, err = DockerClient.ContainerAttach(ctx, containerID, options)
		if err != nil {
			utils.Error("ContainerAttach failed: ", err)
			http.Error(w, "ContainerAttach failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
	
		utils.Log("Attached to existing process in container " + containerID)
	}
	defer resp.Close()

	utils.Log("Attached container " + containerID + " to websocket")

	var WSChan = make(chan []byte, 1024*1024*4)
	var DockerChan = make(chan []byte, 1024*1024*4)

	// Start a goroutine to read from our websocket and write to the container
	go (func() {
		for {
			utils.Debug("Waiting for message from websocket")
			_, message, err := ws.ReadMessage()
			utils.Debug("Got message from websocket")
			if err != nil {
				utils.Error("Failed to read from websocket: ", err)
				break
			}
			WSChan <- []byte((string)(message) + "\n")
		}
	})()

	// Start a goroutine to read from the container and write to our websocket
	go (func() {
		for {
			buf := make([]byte, 1024*1024*4)
			utils.Debug("Waiting for message from container")
			n, err := resp.Reader.Read(buf)
			utils.Debug("Got message from container")
			if err != nil {
				utils.Error("Failed to read from container: ", err)
				break
			}
			DockerChan <- buf[:n]
		}
	})()

	for {
		select {
		case message := <-WSChan:
			utils.Debug("Writing message to container")
			_, err := resp.Conn.Write(message)
			if err != nil {
				utils.Error("Failed to write to container: ", err)
				return
			}
			utils.Debug("Wrote message to container")
		case message := <-DockerChan:
			utils.Debug("Writing message to websocket")
			
			messages := splitIntoChunks(string(message), 5)
			for _, messageSplit := range messages {
				messageJSON, err := json.Marshal(messageSplit)
				if err != nil {
					utils.Error("Failed to marshal message: ", err)
					return
				}
				err = ws.WriteMessage(websocket.TextMessage, messageJSON)
				if err != nil {
					utils.Error("Failed to write to websocket: ", err)
					return
				}
			}
			utils.Debug("Wrote message to websocket")
		}
	}
}