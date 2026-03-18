# Terraform Provider for Cosmos Server

Manage [Cosmos Server](https://github.com/azukaar/cosmos-server) resources with Terraform.

## Requirements

- Terraform >= 1.0
- Go >= 1.22 (for development)
- A running Cosmos Server instance with an API token

## Installation

```hcl
terraform {
  required_providers {
    cosmos = {
      source = "cosmos-cloud.io/azukaar/cosmos"
      # explicit version (for beta release for ex.)
      # version = "0.22.0-unstable08"
    }
  }
}

provider "cosmos" {
  base_url = "https://cosmos.example.com"  # or COSMOS_BASE_URL env
  token    = var.cosmos_token               # or COSMOS_TOKEN env
  insecure = true                           # or COSMOS_INSECURE env
}
```

`terraform init` will automatically download the provider for your platform.

### Manual installation

Alternatively, download the binary from [GitHub Releases](https://github.com/azukaar/cosmos-server/releases) and place it in:

```
~/.terraform.d/plugins/cosmos-cloud.io/azukaar/cosmos/<VERSION>/<OS>_<ARCH>/
```

## Provider Configuration

| Attribute  | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `base_url` | string | yes | Cosmos Server URL (e.g. `https://cosmos.example.com`) |
| `token`    | string | yes | API token (sensitive) |
| `insecure` | bool   | no  | Skip TLS verification. Default: `false` |

All attributes can be set via environment variables: `COSMOS_BASE_URL`, `COSMOS_TOKEN`, `COSMOS_INSECURE`.

## Resources

| Resource | Description |
|----------|-------------|
| `cosmos_route` | Proxy route with smart shield, auth, and filtering |
| `cosmos_docker_service` | Docker container (JSON blob config, recreate on change) |
| `cosmos_docker_volume` | Docker volume |
| `cosmos_docker_network` | Docker network |
| `cosmos_api_token` | API token (token value returned once on create) |
| `cosmos_user` | User account |
| `cosmos_backup` | Backup configuration (password managed server-side) |
| `cosmos_alert` | Alert with conditions and actions |
| `cosmos_cron_job` | Scheduled cron job |
| `cosmos_openid_client` | OpenID Connect client (secret generated server-side, returned once) |
| `cosmos_constellation` | Initialize Constellation VPN (destroy resets the VPN) |
| `cosmos_constellation_device` | VPN device (config YAML returned once on create) |
| `cosmos_constellation_dns` | Constellation DNS entry |
| `cosmos_snapraid` | SnapRAID array configuration |
| `cosmos_storage_mount` | Storage mount (all fields require replacement) |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `cosmos_status` | Server status (hostname, version, docker/constellation status) |
| `cosmos_containers` | List of running Docker containers |

## Examples

See the [examples/](examples/) directory for complete, working configurations:

- **[web-app](examples/web-app/)** — Token, volume, container, route with smart shield, backup, and alert
- **[vpn](examples/vpn/)** — Constellation VPN init, device, and DNS

## Development

### Build and install locally

```bash
cd terraform-provider-cosmos
make build    # compile
make install  # copy to ~/.terraform.d/plugins/
```

After installing, in your Terraform project:

```bash
rm .terraform.lock.hcl && terraform init
```

### Dev overrides (faster iteration)

Add to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "cosmos-cloud.io/azukaar/cosmos" = "/path/to/terraform-provider-cosmos"
  }
  direct {}
}
```

Then just `make build` and run `terraform plan` directly — no `terraform init` needed.

### Project structure

```
main.go                           # Entry point
internal/
  provider/provider.go            # Provider config
  client/client.go                # SDK wrapper, response parsing, retry logic
  resources/                      # All resource implementations
  datasources/                    # Data source implementations
examples/                         # Usage examples
```

### Adding a new resource

1. Create `internal/resources/my_resource.go`
2. Implement `resource.Resource` and `resource.ResourceWithImportState`
3. Register it in `internal/provider/provider.go` -> `Resources()`
4. Run `make build` to verify

### Common patterns

- **Response parsing**: `client.ParseResponse[T]()` for typed, `client.CheckResponse()` for status-only, `client.CheckStreamingResponse()` for streaming endpoints
- **404 handling**: `client.IsNotFound(err)` — Read methods should call `resp.State.RemoveResource(ctx)` on 404
- **Create-time secrets**: `Computed + Sensitive` with `UseStateForUnknown` — capture on create, preserve on read
- **Optional with server defaults**: `Optional + Computed + Default` to avoid phantom diffs
- **Nested objects**: Use `types.Object` (not Go struct pointers) for `SingleNestedAttribute`
