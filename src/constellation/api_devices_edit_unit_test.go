package constellation

import (
	"reflect"
	"testing"
)

func TestUnitNormalizeDeviceTags(t *testing.T) {
	tests := []struct {
		name    string
		in      []string
		want    []string
		wantErr bool
	}{
		{"nil", nil, []string{}, false},
		{"trims", []string{"  object-storage  "}, []string{"object-storage"}, false},
		{"drops empties", []string{"a", "", "   "}, []string{"a"}, false},
		{"dedupes preserving order", []string{"b", "a", "b"}, []string{"b", "a"}, false},
		{"comma rejected", []string{"a,b"}, nil, true},
		{"comma rejected after trim", []string{" ok ", "no,pe"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeDeviceTags(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeDeviceTags(%v) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalizeDeviceTags(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// Role flags are baked into every node's nebula.yml, so a remote tags-only edit must not write them.
func TestUnitDeviceEditFieldsRemoteIsTagsOnly(t *testing.T) {
	request := DeviceEditRequestJSON{DeviceName: "node-1"}
	got := deviceEditFields(request, []string{"object-storage"}, true)

	want := map[string]interface{}{"Tags": []string{"object-storage"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("remote deviceEditFields = %v, want %v", got, want)
	}

	for _, forbidden := range []string{"IsLighthouse", "IsRelay", "IsExitNode", "IsLoadBalancer"} {
		if _, present := got[forbidden]; present {
			t.Errorf("remote edit writes %s — it would clobber the device's topology", forbidden)
		}
	}
}

func TestUnitDeviceEditFieldsSelf(t *testing.T) {
	tests := []struct {
		name    string
		request DeviceEditRequestJSON
		want    map[string]interface{}
	}{
		{
			"lighthouse keeps its flags",
			DeviceEditRequestJSON{IsLighthouse: true, IsRelay: true, IsExitNode: true, IsLoadBalancer: true},
			map[string]interface{}{
				"IsLighthouse": true, "IsRelay": true, "IsExitNode": true, "IsLoadBalancer": true,
				"Tags": []string{"a"},
			},
		},
		{
			// Non-lighthouses cannot be relay, exit, or load balancer.
			"non-lighthouse flags forced off",
			DeviceEditRequestJSON{IsLighthouse: false, IsRelay: true, IsExitNode: true, IsLoadBalancer: true},
			map[string]interface{}{
				"IsLighthouse": false, "IsRelay": false, "IsExitNode": false, "IsLoadBalancer": false,
				"Tags": []string{"a"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deviceEditFields(tt.request, []string{"a"}, false)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deviceEditFields = %v, want %v", got, tt.want)
			}
		})
	}
}
