package main

import (
	"os"

	"github.com/azukaar/cosmos-server/cli/cmd"
)

func main() {
	if err := cmd.Execute(Version); err != nil {
		os.Exit(1)
	}
}
