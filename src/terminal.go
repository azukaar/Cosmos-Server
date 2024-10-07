package main

import (
	// "context"
	"net/http"
	"os/exec"
	// "syscall"
	"time"
	"io"
	// "os"
	// "os/signal"

	"github.com/creack/pty"
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
	
	// Create arbitrary command.
	c := exec.Command("bash")
  // Set environment variables for better terminal emulation
	// env := os.Environ()
	env := []string{}
	env = append(env, "TERM=xterm-256color")
	env = append(env, "LINES=24", "COLUMNS=80")
	c.Env = env
	
	// Start the command with a pty.
	ptmx, err := pty.StartWithSize(c, &pty.Winsize{
		Rows: 24,
		Cols: 80,
	})

	if err != nil {
					return
	}
	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }() // Best effort.

	// Handle pty size.
	// ch := make(chan os.Signal, 1)
	// signal.Notify(ch, syscall.SIGWINCH)
	// go func() {
	// 				for range ch {
	// 								if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
	// 												utils.Error("error resizing pty", err)
	// 								}
	// 				}
	// }()
	// ch <- syscall.SIGWINCH // Initial resize.
	// defer func() { signal.Stop(ch); close(ch) }() // Cleanup signals when done.


	// Set stdin in raw mode.
	// oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	// if err != nil {
	// 				panic(err)
	// }
	// defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	// NOTE: The goroutine will keep reading until the next keystroke before returning.
	// go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	// _, _ = io.Copy(os.Stdout, ptmx)

	// Specify the UID you want to run the command as

	// uid := uint32(1000)  // Replace with the desired UID
	// gid := uint32(1000)  // Replace with the primary GID of the user
	// Set the user and group ID of the process
	// cmd.SysProcAttr = &syscall.SysProcAttr{
	// 		Credential: &syscall.Credential{
	// 				Uid: uid,
	// 				Gid: gid,
	// 		},
	// }
		
	// stdin, err := cmd.StdinPipe()
	// if err != nil {
	// 		utils.Error("Failed to create stdin pipe: ", err)
	// 		http.Error(w, "Failed to create stdin pipe: "+err.Error(), http.StatusInternalServerError)
	// 		return
	// }
	
	// stdout, err := cmd.StdoutPipe()
	// if err != nil {
	// 		utils.Error("Failed to create stdout pipe: ", err)
	// 		http.Error(w, "Failed to create stdout pipe: "+err.Error(), http.StatusInternalServerError)
	// 		return
	// }

	// stderr, err := cmd.StderrPipe()
	// if err != nil {
	// 	utils.Error("Failed to create stderr pipe: ", err)
	// 	http.Error(w, "Failed to create stderr pipe: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

    // Goroutine to handle pty output
    go func() {
			buf := make([]byte, 1024)
			for {
					n, err := ptmx.Read(buf)
					if err != nil {
							if err != io.EOF {
									utils.Error("Error reading from pty: ", err)
							}
							return
					}
					ws.SetWriteDeadline(time.Now().Add(timeoutDuration))
					ws.SetReadDeadline(time.Now().Add(timeoutDuration))
					if err := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
							utils.Error("Failed to write to websocket: ", err)
							return
					}
			}
	}()

	// Main loop to handle websocket input
	for {
			_, message, err := ws.ReadMessage()
			if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
							utils.Error("Websocket error: ", err)
					}
					break
			}

			ws.SetWriteDeadline(time.Now().Add(timeoutDuration))
			ws.SetReadDeadline(time.Now().Add(timeoutDuration))

			if string(message) == "_PING_" {
					ws.WriteMessage(websocket.TextMessage, []byte("_PONG_"))
			} else {
					_, err := ptmx.Write(message)
					if err != nil {
							utils.Error("Failed to write to pty: ", err)
							break
					}
			}
	}

	// Cleanup
	
	utils.Log("Terminal session ended")
}
