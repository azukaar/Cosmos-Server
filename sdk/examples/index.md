# Cosmos SDK Examples

Practical, runnable examples using the `cosmos-cloud-sdk` package. Each script reads `COSMOS_URL` and `COSMOS_TOKEN` from environment variables.

```bash
export COSMOS_URL=https://my-cosmos.example.com
export COSMOS_TOKEN=cosmos_xxx
```

## Examples

### [Dashboard](./dashboard.mjs)
Server overview: list running containers, read system metrics, fetch recent events, and display container logs.

```bash
node sdk/examples/dashboard.mjs
```

### [CI/CD Deploy](./cicd-deploy.mjs)
Deploy a compose-style service (images are pulled automatically) — designed for CI pipelines (GitHub Actions, GitLab CI, etc.).

```bash
DEPLOY_IMAGE=nginx:latest node sdk/examples/cicd-deploy.mjs
```

### [Container Management](./container-management.mjs)
Container lifecycle: stop, start, restart, rolling image update, log search, and auto-update toggle.

```bash
node sdk/examples/container-management.mjs my-container
```

### [Backup & Jobs](./backup-and-jobs.mjs)
List backup repositories and snapshots, trigger backups, and manage scheduled jobs.

```bash
node sdk/examples/backup-and-jobs.mjs
```
