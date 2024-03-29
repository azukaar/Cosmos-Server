package docker

import (
	"context"
	"time"

	"github.com/gorilla/mux"
	"github.com/docker/docker/api/types"
	"github.com/gorilla/websocket"
	"net/http"
	// "strings"
	// "encoding/json"

	conttype "github.com/docker/docker/api/types/container"
	"github.com/azukaar/cosmos-server/src/utils"
)

const timeoutDuration = 2 * time.Minute

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// func splitIntoChunks(input string, chunkSize int) [][]LogOutput {
// 	lines := strings.Split(input, "\n")
// 	var chunks [][]LogOutput

// 	for i := 0; i < len(lines); i += chunkSize {
// 			end := i + chunkSize

// 			// Avoid going over the end of the array
// 			if end > len(lines) {
// 					end = len(lines)
// 			}

// 			var chunk []LogOutput
// 			for j := i; j < end; j++ {
// 				chunk = append(chunk, ParseDockerLogHeader(([]byte)(
// 					lines[j],
// 				)))
// 			}
// 			chunks = append(chunks, chunk)
// 	}

// 	return chunks
// }

func splitIntoChunks(input string) []string {
	// split every 512 characters
	chunkSize := 512
	var chunks []string

	for i := 0; i < len(input); i += chunkSize {
		end := i + chunkSize

		// Avoid going over the end of the array
		if end > len(input) {
			end = len(input)
		}
		
		chunks = append(chunks, input[i:end])
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

	ws.SetReadDeadline(time.Now().Add(timeoutDuration))
	ws.SetWriteDeadline(time.Now().Add(timeoutDuration))

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

		// Start bash if it exists
		resp.Conn.Write([]byte("bash\n")) 
	
		utils.Log("Created new shell and attached to it in container " + containerID)
	} else {
		options := conttype.AttachOptions{
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
	go func(ctx context.Context) {
			defer close(WSChan) // Ensure channel is closed when goroutine exits
			for {
					select {
					case <-ctx.Done(): // Context cancellation
							return
					default:
							ws.SetReadDeadline(time.Now().Add(timeoutDuration))
							_, message, err := ws.ReadMessage()
							if err != nil {
									if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
											return
									}
									utils.Error("Failed to read from websocket: ", err)
									return
							}
							WSChan <- message
					}
			}
	}(ctx) // Pass the context


	// Start a goroutine to read from the container and write to our websocket
	go func(ctx context.Context) {
		defer close(DockerChan) // Ensure the channel is closed when the goroutine exits
	
		for {
			select {
			case <-ctx.Done(): // Check if the context is cancelled
				utils.Debug("Context cancelled, stopping read from container")
				return
			default:
				buf := make([]byte, 1024*1024*4)
				utils.Debug("Waiting for message from container")
				n, err := resp.Reader.Read(buf)
				utils.Debug("Got message from container")
	
				if err != nil {
					utils.Error("Failed to read from container: ", err)
					return
				}
	
				DockerChan <- buf[:n]
			}
		}
	}(ctx) // Pass the context to the goroutine

	for {
		select {
		case message, ok := <-WSChan:
			if !ok { // WSChan is closed
				utils.Debug("WSChan is closed, stopping writes to container")
				return
			}
			utils.Debug("Writing message to container " + string(message))
			_, err := resp.Conn.Write(message)
			if err != nil {
				utils.Error("Failed to write to container: ", err)
				return
			}
			utils.Debug("Wrote message to container")
	
		case message, ok := <-DockerChan:
			if !ok { // DockerChan is closed
				utils.Debug("DockerChan is closed, stopping writes to websocket")
				return
			}
			utils.Debug("Writing message to websocket " + string(message))
			
			messages := splitIntoChunks(string(message))

			ws.SetWriteDeadline(time.Now().Add(timeoutDuration))
			
			for _, messageSplit := range messages {
				err = ws.WriteMessage(websocket.TextMessage, []byte(messageSplit))
				if err != nil {
					utils.Error("Failed to write to websocket: ", err)
					return
				}
			}
		}
	}
	
}