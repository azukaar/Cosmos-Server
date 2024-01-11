package docker 

import (
	"bufio"
	"io"
	"os/exec"
    "os"
    "encoding/json"
    // "path/filepath"
    // "context"

	"github.com/azukaar/cosmos-server/src/utils"
    
	cli "github.com/compose-spec/compose-go/v2/cli"
	composeTypes "github.com/compose-spec/compose-go/v2/types"
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

    callback("$> " +  cmd.String(), Stdout)

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
	return ExecCommand(callback, "docker", "compose", "--project-directory", filePath, "up", "-d")

    // ctx := context.Background()

    // // Configure Compose options (adjust these to your needs)
    // composeFiles := []string{filePath}
    // projectDir := filepath.Dir(filePath)

    // // Create a new Docker CLI client
    // dockerCLI, err := compose.NewDockerCli()
    // if err != nil {
    //     callback(err.Error(), Stderr)
    //     return err
    // }

    // // Create a new Compose project
    // project, err := compose.NewProject(ctx, composeFiles, compose.WithWorkingDirectory(projectDir))
    // if err != nil {
    //     callback(err.Error(), Stderr)
    //     return err
    // }

    // // Create a new Compose service
    // service := project.Service("cosmos")

    // // Start up the services defined in the Docker Compose file
    // err = service.Up(ctx, api.UpOptions{
    //     Detach: true,
    //     RemoveOrphans: true,
    // })

    // if err != nil {
    //     callback(err.Error(), Stderr)
    //     return err
    // }

    // return nil
}

func ComposeDown(filePath string, callback func(message string, outputType int)) error {
	return ExecCommand(callback, "docker", "compose", "--project-directory", filePath, "down")
}

func LoadCompose(dcWorkingDir string) (*composeTypes.Project, error) {
	options, err := cli.NewProjectOptions(os.Args[1:],
		cli.WithWorkingDirectory(dcWorkingDir),
		cli.WithOsEnv,
		cli.WithDotEnv,
		cli.WithConfigFileEnv,
		cli.WithDefaultConfigPath,
	)
	
	if err != nil {
        return nil, err
	}

	project, err := cli.ProjectFromOptions(options)
	if err != nil {
        return nil, err
	}

    return project, nil
}

func LoadComposeFromName(containerName string) (*composeTypes.Project, error) {
    container, err := DockerClient.ContainerInspect(DockerContext, containerName)
    if err != nil {
        return nil, err
    }

	dcWorkingDir := container.Config.Labels["com.docker.compose.project.working_dir"]

    return LoadCompose(dcWorkingDir)
}

func SaveCompose(dcFile string, data []byte) error {
    utils.Log("SaveCompose: saving to file " + dcFile)

	data, err := SimplifyCompose(data)

	if err != nil {
		return err
	}
	
	// write docker compose file
	err = os.WriteFile(dcFile, data, 0644)
	if err != nil {
		return err
	}

    return nil
}

func SaveComposeFromName(containerName string, data []byte) error {
    container, err := DockerClient.ContainerInspect(DockerContext, containerName)
    if err != nil {
        return err
    }
    
    dcFile := container.Config.Labels["com.docker.compose.project.config_files"]
    
    return SaveCompose(dcFile, data)
}

type DockerPsOutput struct {
    Name    string `json:"name"`
    Command string `json:"command"`
    State   string `json:"state"`
    Ports   string `json:"ports"`
}

func DockerComposePs(filePath string) ([]DockerPsOutput, error) {
    cmd := exec.Command("docker", "compose", "-f", filePath, "ps", "--format", "json")
    utils.Debug("Running command: " +  cmd.String())
    output, err := cmd.CombinedOutput()

    if err != nil {
        return nil, err
    }

    var psOutput []DockerPsOutput
    err = json.Unmarshal(output, &psOutput)
    if err != nil {
        return nil, err
    }

    utils.Debug("DockerComposePs: " + string(output))

    return psOutput, nil
}
