package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"fmt"
	"bufio"

	"github.com/gorilla/mux"
	"github.com/azukaar/cosmos-server/src/utils"
)

func PullImageIfMissing(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	imageName := utils.SanitizeSafe(req.URL.Query().Get("imageName"))

	if req.Method == "GET" {
		errD := Connect()
		if errD != nil {
			utils.Error("PullImageIfMissing", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "LN001")
			return
		}
	
		// Enable streaming of response by setting appropriate headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Transfer-Encoding", "chunked")

		flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
        return
    }
		
		_, _, errImage := DockerClient.ImageInspectWithRaw(DockerContext, imageName)
		if errImage != nil {
			utils.Log("PullImageIfMissing - Image not found, pulling " + imageName)
			fmt.Fprintf(w, "PullImageIfMissing - Image not found, pulling " + imageName + "\n")
			flusher.Flush()
			out, errPull := DockerPullImage(imageName)
			if errPull != nil {
				utils.Error("PullImageIfMissing - Image not found.", errPull)
				fmt.Fprintf(w, "[OPERATION FAILED] PullImageIfMissing - Image not found. " + errPull.Error() + "\n")
				flusher.Flush()
				return
			}
			defer out.Close()

			// wait for image pull to finish
			scanner := bufio.NewScanner(out)
			for scanner.Scan() {
				utils.Log(scanner.Text())
				fmt.Fprintf(w, scanner.Text() + "\n")
				flusher.Flush()
			}

			utils.Log("PullImageIfMissing - Image pulled " + imageName)
			fmt.Fprintf(w, "[OPERATION SUCCEEDED]")
			flusher.Flush()
			return
		}
		
		utils.Log("PullImageIfMissing - Image found, skipping " + imageName)
		fmt.Fprintf(w, "[OPERATION SUCCEEDED]")
		flusher.Flush()
	} else {
		utils.Error("PullImageIfMissing: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func PullImage(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	imageName := utils.SanitizeSafe(req.URL.Query().Get("imageName"))

	if req.Method == "GET" {
		errD := Connect()
		if errD != nil {
			utils.Error("PullImageIfMissing", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "LN001")
			return
		}
	
		// Enable streaming of response by setting appropriate headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Transfer-Encoding", "chunked")

		flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
        return
    }
		
		utils.Log("PullImageIfMissing - Image not found, pulling " + imageName)
		fmt.Fprintf(w, "PullImageIfMissing - Image not found, pulling " + imageName + "\n")
		flusher.Flush()
		out, errPull := DockerPullImage(imageName)
		if errPull != nil {
			utils.Error("PullImageIfMissing - Image not found.", errPull)
			fmt.Fprintf(w, "[OPERATION FAILED] PullImageIfMissing - Image not found. " + errPull.Error() + "\n")
			flusher.Flush()
			return
		}
		defer out.Close()

		// wait for image pull to finish
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			utils.Log(scanner.Text())
			fmt.Fprintf(w, scanner.Text() + "\n")
			flusher.Flush()
		}

		utils.Log("PullImageIfMissing - Image pulled " + imageName)
		fmt.Fprintf(w, "[OPERATION SUCCEEDED]")
		flusher.Flush()
		return
	} else {
		utils.Error("PullImageIfMissing: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func CanUpdateImageRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	vars := mux.Vars(req)
	containerId := vars["containerId"]

	if req.Method == "GET" {
		errD := Connect()
		if errD != nil {
			utils.Error("CanUpdateImageRoute", errD)
			utils.HTTPError(w, "Internal server error: "+errD.Error(), http.StatusInternalServerError, "LN001")
			return
		}

		// get Docker container
		container, err := DockerClient.ContainerInspect(context.Background(), containerId)
		if err != nil {
			utils.Error("CanUpdateImageRoute: Error while getting container", err)
			utils.HTTPError(w, "Container Get Error: " + err.Error(), http.StatusInternalServerError, "LN002")
			return
		}

		// check if the container's image can be updated
		canUpdate := false
		imageName := container.Image
		image, _, err := DockerClient.ImageInspectWithRaw(context.Background(), imageName)
		if err != nil {
			utils.Error("CanUpdateImageRoute: Error while inspecting image", err)
			utils.HTTPError(w, "Image Inspection Error: " + err.Error(), http.StatusInternalServerError, "LN003")
			return
		}
		if len(image.RepoDigests) > 0 {
			canUpdate = true
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   canUpdate,
		})
	} else {
		utils.Error("CanUpdateImageRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

