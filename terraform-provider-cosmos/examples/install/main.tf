terraform {
  required_providers {
    cosmos = {
      source = "cosmos-cloud.io/azukaar/cosmos"
    }
  }
}

variable "vm_host" {
  description = "Public IP or DNS of the freshly provisioned VM."
  type        = string
}

variable "ssh_user" {
  description = "SSH user for the install."
  type        = string
  default     = "root"
}

variable "ssh_private_key_path" {
  description = "Path to the SSH private key for the VM."
  type        = string
}

variable "admin_nickname" {
  type    = string
  default = "admin"
}

variable "admin_password" {
  type      = string
  sensitive = true
}

variable "hostname" {
  description = "Public hostname Cosmos will serve."
  type        = string
}

variable "cosmos_licence" {
  description = "Cosmos Pro licence key."
  type        = string
  sensitive   = true
  default     = ""
}

# 1. The provider points at the VM but has no token yet — cosmos_install will
# bootstrap the server and emit one.
provider "cosmos" {
  # Pre-bootstrap server: still on plain HTTP (port 80). The configured alias
  # below switches to HTTPS once cosmos_install issues its Let's Encrypt cert.
  base_url = "http://${var.vm_host}"
  insecure = true
}

# 2. Install Cosmos itself by SSHing in and running the public installer.
resource "cosmos_remote_install" "vm" {
  host             = var.vm_host
  user             = var.ssh_user
  private_key_path = var.ssh_private_key_path
  pro              = true # use https://cosmos-cloud.io/get-pro.sh
  host_key_check   = false
}

# 3. Bootstrap the server config (db, https, admin) once it's reachable.
resource "cosmos_install" "main" {
  depends_on = [cosmos_remote_install.vm]

  mongodb_mode           = "Create"
  hostname               = var.hostname
  https_certificate_mode = "LETSENCRYPT"
  ssl_email              = "ops@example.com"
  nickname               = var.admin_nickname
  password               = var.admin_password
  licence                = var.cosmos_licence
}

# 4. Re-configure the provider to use the freshly issued admin token for
# subsequent resources.
provider "cosmos" {
  alias    = "configured"
  base_url = "https://${var.hostname}"
  token    = cosmos_install.main.admin_token
}

# 5. Manage cluster deployments declaratively.
resource "cosmos_deployment" "web" {
  provider = cosmos.configured

  name     = "web"
  replicas = 2
  strategy = "least-busy"
  tags     = ["edge"]

  compose = jsonencode({
    services = {
      web = {
        image = "nginx:1.27"
        ports = ["80:80"]
      }
    }
  })
}

output "admin_token" {
  value     = cosmos_install.main.admin_token
  sensitive = true
}
