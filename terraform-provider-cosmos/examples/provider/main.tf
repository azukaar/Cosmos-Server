terraform {
  required_providers {
    cosmos = {
      source = "cosmos-cloud.io/azukaar/cosmos"
    }
  }
}

provider "cosmos" {
  base_url = "https://cosmos.example.com" # or COSMOS_BASE_URL env
  token    = var.cosmos_token              # or COSMOS_TOKEN env
  insecure = false                         # or COSMOS_INSECURE env
}

variable "cosmos_token" {
  type      = string
  sensitive = true
}
