# cosmos-cloud-sdk

JavaScript/TypeScript SDK for the [Cosmos Server](https://github.com/azukaar/cosmos-server) API.

## Install

```bash
npm install cosmos-cloud-sdk
```

## Quick Start

```js
const { createClient } = require('cosmos-cloud-sdk');
// or: import { createClient } from 'cosmos-cloud-sdk';

const cosmos = createClient({
  baseUrl: 'https://my-cosmos.example.com',
  token: 'cosmos_abc123...',
});

// List containers
const containers = await cosmos.docker.list();
console.log(containers.data);

// List users
const users = await cosmos.users.list();

// Get server status
const status = await cosmos.getStatus();
```

## API Reference

### `createClient({ baseUrl, token })`

Creates a Cosmos API client.

- `baseUrl` — Your Cosmos server URL (e.g. `https://my-cosmos.example.com`)
- `token` — API token (starts with `cosmos_`)

Returns an object with namespaced API methods:

### Namespaces

| Namespace | Description |
|-----------|-------------|
| `cosmos.docker` | Containers, volumes, networks, images |
| `cosmos.users` | User management, 2FA, notifications |
| `cosmos.config` | Server configuration, routes, DNS |
| `cosmos.storage` | Disks, mounts, RAID, SnapRAID |
| `cosmos.constellation` | VPN devices, tunnels |
| `cosmos.cron` | Job management |
| `cosmos.backups` | Backup & restore |
| `cosmos.metrics` | Metrics & events |
| `cosmos.market` | Marketplace |
| `cosmos.rclone` | Cloud storage (rclone) |
| `cosmos.apiTokens` | API token management |
| `cosmos.auth` | Login, logout, sudo |

### Top-level methods

| Method | Description |
|--------|-------------|
| `cosmos.getStatus()` | Server status |
| `cosmos.isOnline()` | Check connectivity |
| `cosmos.restartServer()` | Restart server |
| `cosmos.forceAutoUpdate()` | Force update |
| `cosmos.terminal(cmd?)` | Host terminal (WebSocket) |
| `cosmos.uploadImage(file, name)` | Upload image |
| `cosmos.checkHost(host)` | DNS check |
| `cosmos.getDNS(host)` | DNS lookup |

### Examples

```js
// Docker
const containers = await cosmos.docker.list();
await cosmos.docker.manageContainer('my-app', 'stop');
await cosmos.docker.manageContainer('my-app', 'start');
const logs = await cosmos.docker.getContainerLogs('my-app', 'error', 100);

// Streaming (image pull with progress)
await cosmos.docker.pullImage('nginx:latest', (line) => {
  console.log('Progress:', line);
});

// Users
await cosmos.users.create({ nickname: 'bob', password: '...' });
const user = await cosmos.users.get('bob');

// Config
const config = await cosmos.config.get();
await cosmos.config.set(config.data);
await cosmos.config.updateDNS({ dnsPort: '53' });

// API tokens
const tokens = await cosmos.apiTokens.list();
const newToken = await cosmos.apiTokens.create({
  name: 'my-automation',
  readOnly: true,
});

// Backups
const snapshots = await cosmos.backups.listSnapshots('daily');
await cosmos.backups.restoreBackup('daily', { snapshotId: 'abc', target: '/data' });

// WebSocket (terminal)
const ws = cosmos.docker.attachTerminal('my-container');
ws.onmessage = (e) => console.log(e.data);
```

## More Examples

See the [examples directory](./examples/index.md) for full, runnable scripts covering dashboards, CI/CD deployments, container management, and backups.

## Response Format

All methods return the raw API response:

```js
{
  status: 'OK',
  data: { /* endpoint-specific data */ },
  message: '...'
}
```

On error, a `CosmosError` is thrown with `message`, `status`, and `code` properties.

## Requirements

- Node.js 18+ (uses built-in `fetch`)
- Works in browsers too (bundled as ESM and CJS)
- WebSocket support requires `ws` package in Node.js < 22

## Version

The SDK version is synced with the Cosmos Server version.
