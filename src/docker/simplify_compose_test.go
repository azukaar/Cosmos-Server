package docker

import (
	"reflect"
	"testing"

	yaml "gopkg.in/yaml.v2"

	// "github.com/azukaar/cosmos-server/src/utils"
)

// mockUtilsLog is a mock function for utils.Log
func mockUtilsLog(msg string) {
	// This function can be left empty or used to log messages during testing
}
func mockUtilsDebug(msg string) {
	// This function can be left empty or used to log messages during testing
}

// TestSimplifyCompose will test the SimplifyCompose function
func TestSimplifyCompose(t *testing.T) {
	// Replace utils.Log with the mock function
	// utils.Log = mockUtilsLog
	// utils.Debug = mockUtilsDebug

	// Define your test cases
	testCases := []struct {
		name     string
		input    string // Docker Compose YAML as a string
		expected string // Expected simplified YAML as a string
	}{
		{
			name: "Simple",
			input: `services:
  service1:
    depends_on:
      db:
        condition: service_started
        restart: true
        required: true
    environment:
      VAR1: value1
    ports:
      - target: 80
        published: "8080"
        protocol: tcp
      - target: 443
        published: "8443"
        protocol: udp
    networks:
      network1: null
    volumes:
    - source: nextcloud
      target: /var/www/html
      type: volume
      volume: {}
`,
			expected: `services:
  service1:
    depends_on:
      - db
    environment:
      - VAR1=value1
    networks:
      - network1
    ports:
      - 8080:80
      - 8443:443/udp
    volumes:
      - nextcloud:/var/www/html
`,
		},

		{
			name: "Complex",
			input: `services:
  service1:
    depends_on:
      db:
        condition: service_healthy
        restart: false
        required: true
    environment:
      VAR1: complex_value1
    networks:
      network1:
        aliases: stuff
    ports:
      - target: 80
        published: "8080"
        protocol: udp
        mode: host
    volumes:
      - source: nextcloud
        target: /var/www/html
        type: volume
        volume:
          nocopy: true
`,
			expected: `services:
  service1:
    depends_on:
      db:
        condition: service_healthy
        restart: false
    environment:
      - VAR1=complex_value1
    networks:
      network1:
        aliases: stuff
    ports:
      - target: 80
        published: "8080"
        protocol: udp
        mode: host
    volumes:
      - source: nextcloud
        target: /var/www/html
        type: volume
        volume:
          nocopy: true
`,
	},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := SimplifyCompose([]byte(tc.input))

			var expectedData map[string]interface{}
			err := yaml.Unmarshal([]byte(tc.expected), &expectedData)
			if err != nil {
				t.Fatalf("error unmarshalling expected data: %v", err)
			}
			expectedBytes, err := yaml.Marshal(expectedData)
			if err != nil {
				t.Fatalf("error marshalling expected data: %v", err)
			}

			if !reflect.DeepEqual(actual, expectedBytes) {
				t.Errorf("SimplifyCompose() = %v, want %v", string(actual), string(expectedBytes))
			}
		})
	}
}
