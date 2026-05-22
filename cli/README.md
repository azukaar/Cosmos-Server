# Cosmos Server CLI

Official command-line interface for [Cosmos Server](https://github.com/azukaar/cosmos-server).
Manage containers, routes, users, backups, storage, and more from any terminal — locally or remotely.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/VHStark/Cosmos-Server/main/cli/scripts/install.sh | bash
```

Or build from source:

```bash
git clone https://github.com/VHStark/Cosmos-Server
cd Cosmos-Server/cli
go build -o cosmos .
```

## Configuration

Run the interactive setup wizard once:

```bash
cosmos configure
```

This saves the server URL to `~/.cosmos/config.yaml` and stores the API token
in the OS keychain (macOS Keychain, Linux libsecret, Windows Credential Manager).

### Creating an API Token

In Cosmos UI → Settings → API Tokens → **Create token**:

| Field | Description |
|-------|-------------|
| **Name** | Identifier for the token |
| **Read-only** | Leave unchecked for full CLI access |
| **IP Whitelist** | Set to `localhost,127.0.0.1,::1` if running on the same server |
| **Expires** | Optional expiration date |

### Multiple servers (profiles)

```bash
cosmos configure --profile homelab   # add a second server
cosmos configure --profile vps
cosmos configure list                # list all profiles
cosmos configure use homelab         # switch active profile
```

### CI / headless environments

No keychain on headless servers — use environment variables instead:

```bash
export COSMOS_URL=http://192.168.0.100
export COSMOS_TOKEN=your_api_token
cosmos containers list
```

## Usage

```bash
# Status
cosmos system status
cosmos system ping --url https://example.com

# Containers
cosmos containers list
cosmos containers inspect my-container
cosmos containers start my-container
cosmos containers stop my-container
cosmos containers restart my-container
cosmos containers recreate my-container
cosmos containers logs my-container --tail 100
cosmos containers check-update my-container
cosmos containers networks
cosmos containers volumes

# Routes (reverse proxy)
cosmos routes list
cosmos routes get my-route
cosmos routes create --data '{"Name":"myapp","Target":"http://localhost:3000","Mode":"PROXY"}'
cosmos routes update my-route --data '{"Target":"http://localhost:4000"}'
cosmos routes delete my-route

# Users
cosmos users list
cosmos users create --nickname alice --email alice@example.com
cosmos users delete alice

# Metrics
cosmos metrics alerts
cosmos metrics show
cosmos metrics events --from 2026-01-01T00:00:00Z

# Backups
cosmos backups list
cosmos backups snapshots mybackup
cosmos backups restore mybackup --snapshot abc123 --target /restore

# Cron
cosmos cron list
cosmos cron jobs-running
cosmos cron run myjob

# Storage
cosmos storage disks
cosmos storage mounts
cosmos storage snapraid-run myarray sync

# Constellation (VPN)
cosmos constellation devices
cosmos constellation tunnels
cosmos constellation dns

# API tokens
cosmos api-tokens list
cosmos api-tokens create --name mytoken

# OpenID Connect
cosmos openid list

# Profile management
cosmos configure list
cosmos configure use homelab
cosmos --profile vps containers list
```

## Command Groups

| Group | Description |
|-------|-------------|
| `configure` | Setup profiles, URL, and token in OS keychain |
| `system` | Server status, ping, DNS, restart, notifications |
| `containers` | Docker container lifecycle, networks, volumes |
| `routes` | Reverse proxy route CRUD |
| `users` | User management and invites |
| `api-tokens` | API token management |
| `openid` | OpenID Connect client management |
| `constellation` | Nebula VPN mesh — devices, DNS, tunnels |
| `backups` | Restic backup config, repositories, snapshots, restore |
| `cron` | CRON configuration and job control |
| `metrics` | Alerts, events, metrics data |
| `storage` | Disks, mounts, SnapRAID |

## Remote usage

The CLI connects to any Cosmos Server over HTTP or HTTPS:

```bash
# Remote server — one-off
cosmos --url https://cosmos.example.com --token my_token containers list

# Remote server — profile
cosmos configure --profile prod
cosmos --profile prod routes list
```

## Architecture

The CLI is a standalone Go module (`github.com/azukaar/cosmos-server/cli`) built
on top of the auto-generated [`go-sdk`](../go-sdk). It does not modify the server binary.

```
cli/
├── main.go                   # entrypoint
├── version.go                # version constant
├── cmd/                      # cobra command groups
│   ├── root.go               # root command + persistent flags
│   ├── configure.go          # profile wizard
│   ├── containers.go
│   ├── routes.go
│   ├── users.go
│   ├── system.go
│   ├── metrics.go
│   ├── backups.go
│   ├── cron.go
│   ├── storage.go
│   ├── constellation.go
│   ├── api_tokens.go
│   └── openid.go
├── internal/
│   ├── config/config.go      # profile resolution + OS keychain
│   ├── client/client.go      # go-sdk client factory
│   └── output/output.go      # table and JSON helpers
└── scripts/
    └── install.sh            # one-liner installer
