package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"bufio"
	"io"
	"strings"
	"encoding/binary"

	"github.com/docker/docker/api/types"
	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils"
)

type LogOutput struct {
	StreamType byte   `json:"streamType"`
	Size       uint32 `json:"size"`
	Output     string `json:"output"`
}

// ParseDockerLogHeader parses the first 8 bytes of a Docker log message
// and returns the stream type, size, and the rest of the message as output.
// It also checks if the message contains a log header and extracts the log message from it.
func ParseDockerLogHeader(data []byte) (LogOutput) {
	var logOutput LogOutput
	logOutput.StreamType = 1 // assume stdout if header not present
	logOutput.Size = uint32(len(data))
	logOutput.Output = string(data)

	if len(data) < 8 {
		return logOutput
	}

	// check if the output contains a log header
	hasHeader := true
	streamType := data[0]
	if(!(streamType >= 0 && streamType <= 2)) {
		hasHeader = false
	}
	if(data[1] != 0 || data[2] != 0 || data[3] != 0) {
		hasHeader = false
	}
	if hasHeader {
		sizeBytes := data[4:8]
		size := binary.BigEndian.Uint32(sizeBytes)

		output := string(data[8:])

		logOutput.StreamType = streamType
		logOutput.Size = size
		logOutput.Output = output
	}
		
	return logOutput
}

func FilterLogs(logReader io.Reader, searchQuery string, limit int) []LogOutput {
	scanner := bufio.NewScanner(logReader)
	logLines := make([]LogOutput, 0)

	// Read all logs into a slice
	for scanner.Scan() {
		line := scanner.Text()

		if len(searchQuery) > 3 && !strings.Contains(strings.ToUpper(line), strings.ToUpper(searchQuery)) {
			continue
		}

		logLines = append(logLines, ParseDockerLogHeader(([]byte)(line)))
	}

	from := utils.Max(len(logLines)-limit, 0)
	logLines = logLines[from:]

	return logLines
}

func GetContainerLogsRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	vars := mux.Vars(req)
	containerId := vars["containerId"]

	if req.Method == "GET" {
		errD := Connect()
		if errD != nil {
			utils.Error("GetContainerLogsRoute", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "LN001")
			return
		}

		query := req.URL.Query()
		limit := 100
		lastReceivedLogs := ""

		if query.Get("limit") != "" {
			limit, _ = strconv.Atoi(query.Get("limit"))
		}

		if query.Get("lastReceivedLogs") != "" {
			lastReceivedLogs = query.Get("lastReceivedLogs")
		}

		errorOnly := false
		if query.Get("errorOnly") != "" {
			errorOnly, _ = strconv.ParseBool(query.Get("errorOnly"))
		}

		options := types.ContainerLogsOptions{
			ShowStdout: !errorOnly,
			ShowStderr: true,
			Timestamps: true,
			Until:      lastReceivedLogs,
		}

		logReader, err := DockerClient.ContainerLogs(context.Background(), containerId, options)
		if err != nil {
			utils.Error("GetContainerLogsRoute: Error while getting container logs", err)
			utils.HTTPError(w, "Container Logs Error: "+err.Error(), http.StatusInternalServerError, "LN002")
			return
		}
		defer logReader.Close()

		lines := FilterLogs(logReader, query.Get("search"), limit)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   lines,
		})
	} else {
		utils.Error("GetContainerLogsRoute: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}