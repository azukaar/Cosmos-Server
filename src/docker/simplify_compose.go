package docker

import (
	"strconv"

	yaml "gopkg.in/yaml.v2"

	"github.com/azukaar/cosmos-server/src/utils" 
)

func simplify_depends_on(data map[string]interface{}) map[string]interface{} {
  if services, ok := data["services"].(map[interface{}]interface{}); ok {
		for _, service := range services {
			if serviceMap, ok := service.(map[interface{}]interface{}); ok {
				if dependsOn, ok := serviceMap["depends_on"].(map[interface{}]interface{}); ok {
					isSimple := true
					dependsOnSimple := []string{}
					dependsOnComplex := map[string]map[interface{}]interface{}{}

					for _depKey, depValue := range dependsOn {
						depKey := _depKey.(string)

						if depMap, ok := depValue.(map[interface{}]interface{}); ok {
							isSimple = true

							// if condition exist and is not service_started
							if condition, ok := depMap["condition"].(string); ok {
								if condition != "service_started" {
									isSimple = false
								}
							}

							// if restart exist and is false
							if restart, ok := depMap["restart"].(bool); ok {
								if !restart {
									isSimple = false
								}
							}

							dependsOnSimple = append(dependsOnSimple, depKey)
							dependsOnComplex[depKey] = map[interface{}]interface{}{
								"condition": depMap["condition"],
								"restart": depMap["restart"],
							}
						}
					}

					if isSimple {
						serviceMap["depends_on"] = dependsOnSimple
					} else {
						serviceMap["depends_on"] = dependsOnComplex
					}
				}
			}
		}
	}

	return data
}

func simplify_environment(data map[string]interface{}) map[string]interface{} {
	if services, ok := data["services"].(map[interface{}]interface{}); ok {
		for _, service := range services {
			if serviceMap, ok := service.(map[interface{}]interface{}); ok {
				if environment, ok := serviceMap["environment"].(map[interface{}]interface{}); ok {
					environmentSimple := []string{}
					for envKey, envValue := range environment {
						key := envKey.(string)
						value := envValue.(string)
						environmentSimple = append(environmentSimple, key+"="+value)
					}
					serviceMap["environment"] = environmentSimple
				}
			}
		}
	}

	return data
}

func simplify_ports(data map[string]interface{}) map[string]interface{} {
	defaultPortOptions := map[interface{}]interface{}{
		"mode": "ingress",
		// Add other default values here if needed
	}

	if services, ok := data["services"].(map[interface{}]interface{}); ok {
		for _, service := range services {
			if serviceMap, ok := service.(map[interface{}]interface{}); ok {
				if ports, ok := serviceMap["ports"].([]interface{}); ok {
					canSimplify := true
					simplifiedPorts := []string{}

					for _, port := range ports {
						if portMap, ok := port.(map[interface{}]interface{}); ok {
							published := portMap["published"].(string)
							target := strconv.Itoa(portMap["target"].(int))
							protocol, protocolExists := portMap["protocol"].(string)

							// Check if non-protocol options are default
							for key, value := range portMap {
								if key != "published" && key != "target" && key != "protocol" {
									if defaultVal, ok := defaultPortOptions[key]; !ok || value != defaultVal {
										canSimplify = false
										break
									}
								}
							}

							tempPort := published + ":" + target
							if protocolExists && protocol != "tcp" { // Include protocol if it's not default 'tcp'
								tempPort += "/" + protocol
							}
							simplifiedPorts = append(simplifiedPorts, tempPort)
						} else {
							canSimplify = false
							break
						}
					}

					if canSimplify {
						serviceMap["ports"] = simplifiedPorts
					}
				}
			}
		}
	}

	return data
}

func simplify_volumes(data map[string]interface{}) map[string]interface{} {
	if services, ok := data["services"].(map[interface{}]interface{}); ok {
		for _, service := range services {
			if serviceMap, ok := service.(map[interface{}]interface{}); ok {
				if volumes, ok := serviceMap["volumes"].([]interface{}); ok {
					simplifiedVolumes := []interface{}{}

					for _, volume := range volumes {
						if volumeMap, ok := volume.(map[interface{}]interface{}); ok {
							// Check if volume has exactly 3 keys: source, target, type
							if len(volumeMap) == 4 && volumeMap["source"] != nil && volumeMap["target"] != nil && volumeMap["type"] != nil {
								// Check if the volume options are empty
								if volumeOptions, ok := volumeMap["volume"].(map[interface{}]interface{}); ok && len(volumeOptions) == 0 {
									source := volumeMap["source"].(string)
									target := volumeMap["target"].(string)
									simplifiedVolume := source + ":" + target
									simplifiedVolumes = append(simplifiedVolumes, simplifiedVolume)
									continue
								}
							}

							// Keep the detailed configuration if the conditions are not met
							simplifiedVolumes = append(simplifiedVolumes, volumeMap)
						}
					}

					serviceMap["volumes"] = simplifiedVolumes
				}
			}
		}
	}

	return data
}

func simplify_networks(data map[string]interface{}) map[string]interface{} {
	if services, ok := data["services"].(map[interface{}]interface{}); ok {
		for _, service := range services {
			if serviceMap, ok := service.(map[interface{}]interface{}); ok {
				if networks, ok := serviceMap["networks"].(map[interface{}]interface{}); ok {
					allEmpty := true
					networkNames := make([]string, 0, len(networks))

					for networkName, networkOptions := range networks {
						// Check if network options is an empty map
						if networkOptionsMap, _ := networkOptions.(map[interface{}]interface{}); len(networkOptionsMap) != 0 {
							allEmpty = false
							break
						}
						networkNames = append(networkNames, networkName.(string))
					}

					if allEmpty {
						// If all network options are empty maps, simplify to an array of strings
						serviceMap["networks"] = networkNames
					}
				}
			}
		}
	}

	return data
}

func SimplifyCompose(data []byte) []byte {
	utils.Log("Simplifying docker-compose.yml")

	var rawData map[string]interface{}

	err := yaml.Unmarshal(data, &rawData)
	if err != nil {
		panic(err)
	}

	rawData = simplify_depends_on(rawData)
	rawData = simplify_environment(rawData)
	rawData = simplify_ports(rawData)
	rawData = simplify_volumes(rawData)
	rawData = simplify_networks(rawData)

	data, err = yaml.Marshal(rawData)
	if err != nil {
		panic(err)
	}

	return data
}