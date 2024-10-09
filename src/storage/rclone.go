package storage 

import (
	"os"
	"os/exec"
	"io"
	"bufio"
	"bytes"
	
	"github.com/rclone/rclone/cmd"
	
	_ "github.com/rclone/rclone/backend/all" // import all backends
	_ "github.com/rclone/rclone/cmd/all"    // import all commands
	_ "github.com/rclone/rclone/lib/plugin" // import plugins

	"github.com/azukaar/cosmos-server/src/utils"
)


func RunRClone(args []string) {
	cmd.Root.SetArgs(args)
	cmd.Main()
}


func RunRCloneCommand(command []string) (*exec.Cmd, io.WriteCloser, *bytes.Buffer, *bytes.Buffer) {
	utils.Log("[RemoteStorage] Initializing remote storage")
	args := []string{"rclone"}
	args = append(args, command...)
	cmd := exec.Command(os.Args[0], args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
			utils.Error("[RemoteStorage] Error creating stdin pipe", err)
			return nil, nil, nil, nil
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
			utils.Error("[RemoteStorage] Error creating stdout pipe", err)
			return nil, nil, nil, nil
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
			utils.Error("[RemoteStorage] Error creating stderr pipe", err)
			return nil, nil, nil, nil
	}

	err = cmd.Start()
	if err != nil {
			utils.Error("[RemoteStorage] Error starting rclone command", err)
			return nil, nil, nil, nil
	}

	go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
					line := scanner.Text()
					utils.Debug("[RemoteStorage] " + line)
					stdoutBuf.WriteString(line + "\n")
			}
	}()

	go func() {
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
					line := scanner.Text()
					utils.Error("[RemoteStorage] " + line, nil)
					stderrBuf.WriteString(line + "\n")
			}
	}()

	return cmd, stdin, &stdoutBuf, &stderrBuf
}


func InitRemoteStorage() bool {
	configLocation := utils.CONFIGFOLDER + "rclone.conf"
	utils.ProxyRCloneUser = utils.GenerateRandomString(8)
	utils.ProxyRClonePwd = utils.GenerateRandomString(16)
	
	if _, err := os.Stat(configLocation); os.IsNotExist(err) {
			utils.Log("[RemoteStorage] Creating rclone config file")
			file, err := os.Create(configLocation)
			if err != nil {
					utils.Error("[RemoteStorage] Error creating rclone config file", err)
					return false
			}
			file.Close()
	}

	utils.Log("[RemoteStorage] Initializing remote storage")
	RunRCloneCommand([]string{"rcd", "--rc-user=" + utils.ProxyRCloneUser, "--rc-pass=" + utils.ProxyRClonePwd, "--config=" + configLocation, "--rc-baseurl=/cosmos/rclone"})

	return true
}
