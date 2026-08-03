terraform {
  required_providers {
    cosmos = {
      source  = "cosmos-cloud.io/azukaar/cosmos"
      version = "0.1.0"
    }
  }
}

provider "cosmos" {
  base_url = "http://localhost:8080"
  token    = "cosmos_LUQDv-Wfhd_CaKnAEA6Z48zHFo_5wCpHJof84EbTxvM"
  insecure = true
}

# resource "cosmos_route" "test" {
#   name   = "test-route"
#   target = "http://localhost:1234"
#   use_host = true
#   host   = "localhost:1235"
#   mode   = "PROXY"
# }

# resource "cosmos_api_token" "nginx_token" {
#   name        = "nginx-readonly"
#   description = "Token for nginx container"
#   read_only   = true
#   expiry_days = 30
# }

# resource "cosmos_docker_service" "nginx" {
#   name  = "nginx-test"
#   image = "nginx:latest"
#   service_json = jsonencode({
#     ports       = ["8888:80"]
#     environment = ["COSMOS_TOKEN=${cosmos_api_token.nginx_token.token}"]
#   })
# }

# Initialize Constellation VPN
resource "cosmos_constellation" "vpn" {
  device_name   = "server"
  hostname      = "10.255.255.254,192.168.1.102,169.254.83.107"
  ip_range      = "192.168.201.0/24"
  is_lighthouse = true
}

# Device without nickname
resource "cosmos_constellation_device" "cosmos-2" {
  device_name = "cosmos-2"
  ip          = "192.168.201.3"
  is_lighthouse = true
  public_hostname      = "192.168.1.204"
  cosmos_node = 1
  depends_on  = [cosmos_constellation.vpn]
} 
