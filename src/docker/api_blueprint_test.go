package docker

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestDependsOnField(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      DependsOnField
		either_output []string
	}{
		{
			name:          "Array input",
			input:         `["db", "redis"]`,
			expected:      DependsOnField{"db": nil, "redis": nil},
			either_output: []string{`["db","redis"]`, `["redis","db"]`},
		},
		{
			name:          "Object input without conditions",
			input:         `{"db": {}, "redis": {}}`,
			expected:      DependsOnField{"db": {}, "redis": {}},
			either_output: []string{`["db","redis"]`, `["redis","db"]`},
		},
		{
			name:  "Object input with conditions",
			input: `{"db": {"condition": "service_healthy"}, "redis": {"condition": "service_started"}}`,
			expected: DependsOnField{
				"db":    &DependsOnCondition{Condition: "service_healthy"},
				"redis": &DependsOnCondition{Condition: "service_started"},
			},
			either_output: []string{`{"db":{"condition":"service_healthy"},"redis":{"condition":"service_started"}}`},
		},
		{
			name:          "Mixed object input",
			input:         `{"db": {"condition": "service_healthy"}, "redis": {}}`,
			expected:      DependsOnField{"db": &DependsOnCondition{Condition: "service_healthy"}, "redis": {}},
			either_output: []string{`{"db":{"condition":"service_healthy"},"redis":{}}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got DependsOnField
			err := json.Unmarshal([]byte(tt.input), &got)
			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Unmarshal() got = %v, want %v", got, tt.expected)
			}

			output, err := json.Marshal(got)
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			found := false
			for _, o := range tt.either_output {
				if o == string(output) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Marshal() got = %v, want %v", string(output), tt.either_output)
			}
		})
	}
}

func TestDependsOnFieldInvalidInput(t *testing.T) {
	invalidInputs := []string{
		`"invalid"`,
		`123`,
		`[1, 2, 3]`,
		`{"key": "value"}`,
	}

	for _, input := range invalidInputs {
		var got DependsOnField
		err := json.Unmarshal([]byte(input), &got)
		if err == nil {
			t.Errorf("Expected error for input %s, but got none", input)
		}
	}
}

func TestReOrderServices(t *testing.T) {
	tests := []struct {
		name            string
		input           map[string]ContainerCreateRequestContainer
		either_expected [][]ContainerCreateRequestContainer
	}{
		{
			name: "No dependencies",
			input: map[string]ContainerCreateRequestContainer{
				"service1": {DependsOn: DependsOnField{}},
				"service2": {DependsOn: DependsOnField{}},
				"service3": {DependsOn: DependsOnField{}},
			},
			either_expected: [][]ContainerCreateRequestContainer{
				{
					{DependsOn: DependsOnField{}},
					{DependsOn: DependsOnField{}},
					{DependsOn: DependsOnField{}},
				},
			},
		},
		{
			name: "Single dependency",
			input: map[string]ContainerCreateRequestContainer{
				"service1": {DependsOn: DependsOnField{"service2": nil}},
				"service2": {DependsOn: DependsOnField{}},
				"service3": {DependsOn: DependsOnField{}},
			},
			either_expected: [][]ContainerCreateRequestContainer{
				{
					{DependsOn: DependsOnField{}},
					{DependsOn: DependsOnField{"service2": nil}},
					{DependsOn: DependsOnField{}},
				},
				{
					{DependsOn: DependsOnField{}},
					{DependsOn: DependsOnField{}},
					{DependsOn: DependsOnField{"service2": nil}},
				},
			},
		},
		{
			name: "Chained dependencies",
			input: map[string]ContainerCreateRequestContainer{
				"service1": {DependsOn: DependsOnField{"service2": nil}},
				"service2": {DependsOn: DependsOnField{"service3": nil}},
				"service3": {DependsOn: DependsOnField{}},
			},
			either_expected: [][]ContainerCreateRequestContainer{{
				{DependsOn: DependsOnField{}},
				{DependsOn: DependsOnField{"service3": nil}},
				{DependsOn: DependsOnField{"service2": nil}},
			}},
		},
		{
			name: "Multiple dependencies",
			input: map[string]ContainerCreateRequestContainer{
				"service1": {Name: "service1", DependsOn: DependsOnField{"service3": nil}},
				"service2": {Name: "service2", DependsOn: DependsOnField{
					"service1": nil,
					"service3": nil,
					"service4": nil,
				}},
				"service3": {Name: "service3", DependsOn: DependsOnField{}},
				"service4": {Name: "service4", DependsOn: DependsOnField{}},
			},
			either_expected: [][]ContainerCreateRequestContainer{
				{
					{Name: "service4", DependsOn: DependsOnField{}},
					{Name: "service3", DependsOn: DependsOnField{}},
					{Name: "service2", DependsOn: DependsOnField{
						"service1": nil,
						"service3": nil,
						"service4": nil,
					}},
					{Name: "service1", DependsOn: DependsOnField{"service3": nil}},
				},
				{
					{Name: "service3", DependsOn: DependsOnField{}},
					{Name: "service4", DependsOn: DependsOnField{}},
					{Name: "service1", DependsOn: DependsOnField{"service3": nil}},
					{Name: "service2", DependsOn: DependsOnField{
						"service1": nil,
						"service3": nil,
						"service4": nil,
					}},
				},
				{
					{Name: "service4", DependsOn: DependsOnField{}},
					{Name: "service3", DependsOn: DependsOnField{}},
					{Name: "service1", DependsOn: DependsOnField{"service3": nil}},
					{Name: "service2", DependsOn: DependsOnField{
						"service1": nil,
						"service3": nil,
						"service4": nil,
					}},
				},
				{
					{Name: "service3", DependsOn: DependsOnField{}},
					{Name: "service4", DependsOn: DependsOnField{}},
					{Name: "service2", DependsOn: DependsOnField{
						"service1": nil,
						"service3": nil,
						"service4": nil,
					}},
					{Name: "service1", DependsOn: DependsOnField{"service3": nil}},
				},
				{
					{Name: "service3", DependsOn: DependsOnField{}},
					{Name: "service1", DependsOn: DependsOnField{"service3": nil}},
					{Name: "service4", DependsOn: DependsOnField{}},
					{Name: "service2", DependsOn: DependsOnField{
						"service1": nil,
						"service3": nil,
						"service4": nil,
					}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReOrderServices(tt.input)

			if err != nil {
				t.Errorf("ReOrderServices() error = %v", err)
				return
			}

			found := false
			for _, e := range tt.either_expected {
				if reflect.DeepEqual(e, got) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Marshal() got = %v, want %v", got, tt.either_expected)
			}
		})
	}
}
