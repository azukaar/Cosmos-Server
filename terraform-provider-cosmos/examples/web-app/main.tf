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

# API token for the app to talk back to Cosmos
resource "cosmos_api_token" "app" {
  name        = "web-app"
  description = "Token for the web application"
  read_only   = true
}

# Persistent storage
resource "cosmos_docker_volume" "app_data" {
  name = "web-app-data"
}

# The application container
resource "cosmos_docker_service" "app" {
  name  = "web-app"
  image = "nginx:latest"
  service_json = jsonencode({
    ports       = ["8080:80"]
    environment = ["COSMOS_TOKEN=${cosmos_api_token.app.token}"]
    volumes = [{
      source = cosmos_docker_volume.app_data.name
      target = "/usr/share/nginx/html"
      type   = "volume"
    }]
  })
}

# Route with smart shield protection
resource "cosmos_route" "app" {
  name              = "web-app"
  target            = "http://web-app:80"
  host              = "app.example.com"
  mode              = "PROXY"
  use_host          = true
  auth_enabled      = true
  block_common_bots = true
  block_api_abuse   = true

  smart_shield = {
    enabled                = true
    per_user_request_limit = 100
    policy_strictness      = 2
  }
}

# Nightly backup
resource "cosmos_backup" "app_data" {
  name       = "web-app-backup"
  repository = "/backups/web-app"
  source     = "/var/lib/docker/volumes/web-app-data/_data"
  crontab    = "0 2 * * *"
}

# CPU alert
resource "cosmos_alert" "high_cpu" {
  name            = "web-app-cpu"
  enabled         = true
  tracking_metric = "cosmos.system.docker.cpu.*"
  period          = "daily"
  severity        = "warn"

  condition = {
    operator = "gt"
    value    = 80
  }

  actions = [{
    type = "notification"
  }]
}
