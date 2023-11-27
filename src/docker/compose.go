package docker 

import (
	"bufio"
	"io"
	"os/exec"

	"github.com/azukaar/cosmos-server/src/utils"
)

const (
    Stdout = iota
    Stderr
)


// readOutput reads from the given reader and uses the callback function to handle the output.
func readOutput(r io.Reader, callback func(message string, outputType int), outputType int) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
			callback(scanner.Text(), outputType)
	}
}

// ExecCommand runs the specified command and uses a callback function to handle the output.
func ExecCommand(callback func(message string, outputType int), args ...string) error {
    cmd := exec.Command(args[0], args[1:]...)
		utils.Debug("Running command: " +  cmd.String())

		callback("Running command: " +  cmd.String(), Stdout)

    // Create pipes for stdout and stderr
    stdoutPipe, err := cmd.StdoutPipe()
    if err != nil {
        return err
    }

    stderrPipe, err := cmd.StderrPipe()
    if err != nil {
        return err
    }

    // Start the command
    if err := cmd.Start(); err != nil {
        return err
    }

    // Read from stdout and stderr in separate goroutines
    go readOutput(stdoutPipe, callback, Stdout)
    go readOutput(stderrPipe, callback, Stderr)

    // Wait for the command to finish
    return cmd.Wait()
}

func ComposeUp(filePath string, callback func(message string, outputType int)) error {
	return ExecCommand(callback, "docker", "compose", "--project-directory", filePath, "up", "--remove-orphans", "-d")
}