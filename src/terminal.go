package main

import (
	"context"
	"io"
	"net/http"
	"os/exec"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/azukaar/cosmos-server/src/utils"
)

const timeoutDuration = 2 * time.Minute

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func splitIntoChunks(input string) []string {
	chunkSize := 512
	var chunks []string

	for i := 0; i < len(input); i += chunkSize {
		end := i + chunkSize
		if end > len(input) {
			end = len(input)
		}
		chunks = append(chunks, input[i:end])
	}

	return chunks
}

func HostTerminalRoute(w http.ResponseWriter, r *http.Request) {
	if utils.AdminOnly(w, r) != nil {
		return
	}
	utils.Log("Attempting to attach to host terminal")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.Error("Failed to set websocket upgrade: ", err)
		http.Error(w, "Failed to set websocket upgrade: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer ws.Close()

	ws.SetReadDeadline(time.Now().Add(timeoutDuration))
	ws.SetWriteDeadline(time.Now().Add(timeoutDuration))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Specify the UID you want to run the command as
	uid := uint32(1000)  // Replace with the desired UID
	gid := uint32(1000)  // Replace with the primary GID of the user
	
	cmd := exec.CommandContext(ctx, "/bin/sh")
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
			utils.Error("Failed to create stdin pipe: ", err)
			http.Error(w, "Failed to create stdin pipe: "+err.Error(), http.StatusInternalServerError)
			return
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
			utils.Error("Failed to create stdout pipe: ", err)
			http.Error(w, "Failed to create stdout pipe: "+err.Error(), http.StatusInternalServerError)
			return
	}
	
	// Set the user and group ID of the process
	cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
					Uid: uid,
					Gid: gid,
			},
	}
	
	// Start the command
	err = cmd.Start()
	if err != nil {
			utils.Error("Failed to start command: ", err)
			http.Error(w, "Failed to start command: "+err.Error(), http.StatusInternalServerError)
			return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		utils.Error("Failed to create stderr pipe: ", err)
		http.Error(w, "Failed to create stderr pipe: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := cmd.Start(); err != nil {
		utils.Error("Failed to start command: ", err)
		http.Error(w, "Failed to start command: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Log("Attached to host terminal")

	go func() {
		defer cmd.Process.Kill()
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					return
				}
				utils.Error("Failed to read from websocket: ", err)
				return
			}
			if string(message) == "_PING_" {
				ws.SetWriteDeadline(time.Now().Add(timeoutDuration))
				ws.WriteMessage(websocket.TextMessage, []byte("_PONG_"))
			} else {
				stdin.Write(message)
			}
		}
	}()

	go func() {
		combined := io.MultiReader(stdout, stderr)
		buf := make([]byte, 1024)
		for {
			n, err := combined.Read(buf)
			if err != nil {
				if err != io.EOF {
					utils.Error("Failed to read from command output: ", err)
				}
				return
			}
			messages := splitIntoChunks(string(buf[:n]))
			for _, messageSplit := range messages {
				ws.SetWriteDeadline(time.Now().Add(timeoutDuration))
				if err := ws.WriteMessage(websocket.TextMessage, []byte(messageSplit)); err != nil {
					utils.Error("Failed to write to websocket: ", err)
					return
				}
			}
		}
	}()

	cmd.Wait()
}
