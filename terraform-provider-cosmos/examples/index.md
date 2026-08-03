# Examples

Each subdirectory is a self-contained Terraform module. Pick the one that matches your scenario.

## Greenfield bootstrap (machine + Cosmos)

| Example | What you get |
|---------|--------------|
| [install](./install/index.md) | Single Cosmos node from one already-provisioned VM (any cloud / bare metal). Uses `cosmos_remote_install` + `cosmos_install` + one `cosmos_deployment`. |
| [aws-3node](./aws-3node/index.md) | 3× EC2 instances → equal-footing Cosmos cluster (founder + 2 peers, all `cosmos_node = 2`) + one cluster-replicated deployment. |
| [digitalocean-3node](./digitalocean-3node/index.md) | Same shape as `aws-3node` on DigitalOcean Droplets. |

## Resource-only examples

These assume you already have a configured Cosmos server and an admin token; they only exercise specific resources.

| Example | Resources |
|---------|-----------|
| [provider](./provider/main.tf) | Bare provider configuration (`base_url` + `token`). Copy as a starter. |
| [vpn](./vpn/main.tf) | `cosmos_constellation` + `cosmos_constellation_device` + `cosmos_constellation_dns` for setting up a VPN on an existing Cosmos server. |
| [web-app](./web-app/main.tf) | `cosmos_api_token`, `cosmos_docker_volume`, `cosmos_docker_service`, `cosmos_route`, `cosmos_backup`, `cosmos_alert` — typical app stack on a configured server. |
| [backup](./backup/main.tf) | `cosmos_backup` standalone. **Destroying this resource permanently deletes its backups.** |

## How to run any example

```bash
# 1. Make sure the provider is locally available — either install it from the
#    registry, or use a dev_overrides block in ~/.terraformrc pointing to a
#    locally-built binary:
#
#    provider_installation {
#      dev_overrides {
#        "cosmos-cloud.io/azukaar/cosmos" = "/path/to/built/binary/dir"
#      }
#      direct {}
#    }

cd examples/<name>

terraform init   # skip if using dev_overrides
terraform plan -var '…' -var '…'
terraform apply -var '…'
```

Each per-example `index.md` lists the variables that example expects.

## Conventions used across examples

- All cluster/peer registrations use **`cosmos_node = 2`** — the device joins as a full Cosmos server peer, not an agent.
- "Founder" refers only to the node that generates the CA; it has no privileged role afterwards. Deployments live in NATS KV and replicate cluster-wide, so any node URL works as the entry point.
- `cosmos_install` always emits an `admin_token` (Sensitive Computed). Chain a second `provider "cosmos" { alias = ... token = cosmos_install.x.admin_token }` to authenticate downstream resources.
- `cosmos_remote_install.host_key_check = false` is convenient for cloud-provisioned VMs where the host key isn't known at plan time. Set `host_key` for production.
