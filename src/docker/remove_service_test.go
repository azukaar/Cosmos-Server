package docker

import (
	"testing"

	"github.com/docker/docker/api/types"
)

// deploymentContainerAlive treats an exited lazy replica as alive, an exited non-lazy one as dead.
func TestDeploymentContainerAlive(t *testing.T) {
	cases := []struct {
		name string
		c    types.Container
		want bool
	}{
		{"running", types.Container{State: "running", Labels: map[string]string{DeploymentLabel: "d"}}, true},
		{"exited non-lazy is collectable", types.Container{State: "exited", Labels: map[string]string{DeploymentLabel: "d"}}, false},
		{"exited lazy stays alive", types.Container{State: "exited", Labels: map[string]string{DeploymentLabel: "d", LazyLabel: "true"}}, true},
		{"exited with lazy=false", types.Container{State: "exited", Labels: map[string]string{DeploymentLabel: "d", LazyLabel: "false"}}, false},
		{"no labels", types.Container{State: "exited"}, false},
	}
	for _, c := range cases {
		if got := deploymentContainerAlive(c.c); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
