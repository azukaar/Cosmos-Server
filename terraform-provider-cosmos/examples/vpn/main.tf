terraform {
  required_providers {
    cosmos = {
      source = "cosmos-cloud.io/azukaar/cosmos"
    }
  }
}

provider "cosmos" {
  base_url = "https://cosmos.example.com"
  token    = var.cosmos_token
  insecure = true
}

variable "cosmos_token" {
  type      = string
  sensitive = true
}

# Initialize the VPN
resource "cosmos_constellation" "vpn" {
  device_name   = "server"
  hostname      = "vpn.example.com"
  ip_range      = "192.168.201.0/24"
  is_lighthouse = true
}

# Add a client device
resource "cosmos_constellation_device" "laptop" {
  device_name = "laptop"
  nickname    = "laptop"
  ip          = "192.168.201.2"
  depends_on  = [cosmos_constellation.vpn]
}

# The device config is needed to connect — only available after creation
output "laptop_config" {
  value     = cosmos_constellation_device.laptop.config
  sensitive = true
}

# Internal DNS for the VPN
resource "cosmos_constellation_dns" "app" {
  key        = "app.vpn"
  type       = "A"
  value      = "192.168.201.2"
  depends_on = [cosmos_constellation.vpn]
}
