package docker

import "testing"

func TestStopTimeoutUnsetUsesDockerDefault(t *testing.T) {
	if got := stopTimeout(0); got != nil {
		t.Errorf("stopTimeout(0) = %d, want nil", *got)
	}
	if got := stopTimeout(30); got == nil || *got != 30 {
		t.Errorf("stopTimeout(30) = %v, want 30", got)
	}
}
