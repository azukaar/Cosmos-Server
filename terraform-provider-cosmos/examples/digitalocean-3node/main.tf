terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
    cosmos = {
      source = "cosmos-cloud.io/azukaar/cosmos"
    }
  }
}

# ─── Inputs ─────────────────────────────────────────────────────────────────

variable "do_token" {
  type      = string
  sensitive = true
}

variable "ssh_key_fingerprint" {
  description = "Fingerprint of an SSH key already registered on DigitalOcean."
  type        = string
}

variable "ssh_private_key_path" {
  description = "Local path to the matching private key (used to SSH and run the installer)."
  type        = string
}

variable "region" {
  type    = string
  default = "fra1"
}

variable "size" {
  type    = string
  default = "s-2vcpu-2gb"
}

variable "image" {
  type    = string
  default = "ubuntu-24-04-x64"
}

variable "node_hostnames" {
  description = "Three DNS names, one per node. Each must resolve to the matching digitalocean_droplet.node[i].ipv4_address before apply, so Let's Encrypt can issue per-node certs. node_hostnames[0] also doubles as the constellation VPN endpoint (UDP 4242)."
  type        = list(string)

  validation {
    condition     = length(var.node_hostnames) == 3
    error_message = "node_hostnames must contain exactly three entries."
  }
}

variable "admin_password" {
  type      = string
  sensitive = true
}

variable "cosmos_licence" {
  description = "Cosmos Pro licence key."
  type        = string
  sensitive   = true
}

# ─── DigitalOcean infrastructure ────────────────────────────────────────────

provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_droplet" "node" {
  count    = 3
  name     = "cosmos-node-${count.index}"
  region   = var.region
  size     = var.size
  image    = var.image
  ssh_keys = [var.ssh_key_fingerprint]
  # All three are equal Cosmos servers once joined; node 0 is just the
  # bootstrap point that owns CA generation.
  tags     = ["cosmos", count.index == 0 ? "bootstrap" : "peer"]
}

resource "digitalocean_firewall" "cosmos" {
  name        = "cosmos-cluster"
  droplet_ids = digitalocean_droplet.node[*].id

  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }
  inbound_rule {
    protocol         = "tcp"
    port_range       = "80"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }
  inbound_rule {
    protocol         = "tcp"
    port_range       = "443"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }
  inbound_rule {
    protocol         = "udp"
    port_range       = "4242"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }
  outbound_rule {
    protocol              = "tcp"
    port_range            = "all"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
  outbound_rule {
    protocol              = "udp"
    port_range            = "all"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}

# ─── Stage 1: install Cosmos on every droplet ──────────────────────────────

resource "cosmos_remote_install" "node" {
  count            = 3
  host             = digitalocean_droplet.node[count.index].ipv4_address
  user             = "root"
  private_key_path = var.ssh_private_key_path
  pro              = true
  host_key_check   = false
}

# ─── Stage 2: bootstrap node 0 (the constellation founder) ────────────────
# Node 0 isn't special after this stage — it just has to exist first because
# someone has to create the CA. All three nodes operate as equal Cosmos
# servers (cosmos_node = 2) once they've joined.

provider "cosmos" {
  # Pre-bootstrap server: still on plain HTTP (port 80). Once cosmos_install
  # finishes, the cluster alias below switches to HTTPS via the real DNS name.
  alias    = "founder_anon"
  base_url = "http://${digitalocean_droplet.node[0].ipv4_address}"
  insecure = true
}

resource "cosmos_install" "founder" {
  provider   = cosmos.founder_anon
  depends_on = [cosmos_remote_install.node]

  mongodb_mode           = "Create"
  hostname               = var.node_hostnames[0]
  https_certificate_mode = "LETSENCRYPT"
  ssl_email              = "ops@example.com"
  nickname               = "admin"
  password               = var.admin_password
  licence                = var.cosmos_licence
}

provider "cosmos" {
  alias    = "cluster"
  base_url = "https://${var.node_hostnames[0]}"
  token    = cosmos_install.founder.admin_token
}

resource "cosmos_constellation" "main" {
  provider      = cosmos.cluster
  device_name   = "node-0"
  hostname      = var.node_hostnames[0]
  is_lighthouse = true
  nats_replicas = 3
}

# ─── Stage 3: register the other two nodes as full Cosmos servers ─────────
# cosmos_node = 2 → the device joins as an equal-footing server peer (not
# an agent). The emitted .config is the constellation-join blob each peer
# feeds into its own cosmos_install.

resource "cosmos_constellation_device" "peer" {
  provider   = cosmos.cluster
  count      = 2
  depends_on = [cosmos_constellation.main]

  device_name = "node-${count.index + 1}"
  ip          = "192.168.201.${count.index + 2}"
  nickname    = "admin"
  cosmos_node = 2
}

# ─── Stage 4: bootstrap each peer, joining the founder's constellation ────

provider "cosmos" {
  alias    = "peer1_anon"
  base_url = "http://${digitalocean_droplet.node[1].ipv4_address}"
  insecure = true
}

provider "cosmos" {
  alias    = "peer2_anon"
  base_url = "http://${digitalocean_droplet.node[2].ipv4_address}"
  insecure = true
}

resource "cosmos_install" "peer1" {
  provider   = cosmos.peer1_anon
  depends_on = [cosmos_remote_install.node]

  mongodb_mode           =  "Create"
  hostname               = var.node_hostnames[1]
  https_certificate_mode = "LETSENCRYPT"
  ssl_email              = "ops@example.com"
  nickname               = "admin"
  password               = var.admin_password
  constellation_config   = cosmos_constellation_device.peer[0].config
}

resource "cosmos_install" "peer2" {
  provider   = cosmos.peer2_anon
  depends_on = [cosmos_remote_install.node]

  mongodb_mode           =  "Create"
  hostname               = var.node_hostnames[2]
  https_certificate_mode = "LETSENCRYPT"
  ssl_email              = "ops@example.com"
  nickname               = "admin"
  password               = var.admin_password
  constellation_config   = cosmos_constellation_device.peer[1].config
}

# ─── Stage 5: declarative deployments — replicated across the cluster ─────
# Deployments live in NATS KV, replicated cluster-wide, so this resource
# can target any of the 3 nodes. We use the founder's URL purely as an
# entry point.

resource "cosmos_deployment" "web" {
  provider   = cosmos.cluster
  depends_on = [cosmos_install.peer1, cosmos_install.peer2]

  name     = "web"
  replicas = 3
  strategy = "least-busy"

  compose = jsonencode({
    services = {
      web = {
        image   = "nginx:1.27"
        ports   = ["80:80"]
        restart = "unless-stopped"
      }
    }
  })
}

# ─── Outputs ────────────────────────────────────────────────────────────────

output "node_ips" {
  value = [for d in digitalocean_droplet.node : d.ipv4_address]
}

output "admin_token" {
  value     = cosmos_install.founder.admin_token
  sensitive = true
}
