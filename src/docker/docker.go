// Copyright 2023 Yann-Sebastien Tremblay-Desrochers, modified by Plurality
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"golang.org/x/net/context/ctxhttp"
)

var DockerClient *client.Client
var DockerContext context.Context

var ContainerMutex sync.Mutex
var ContainerMutexList sync.Mutex

// DockerClientInit initializes the Docker client used by the rest of the app
func DockerClientInit() {
	utils.Log("Initializing Docker client...")
	// See https://godoc.org/github.com/docker/docker/client#Client
	DockerClient, DockerContext = Client()

}