// Container Management Example
// Demonstrates container lifecycle operations:
// list, inspect, stop/start/restart, log search, image update, and auto-update.
//
// Usage:
//   COSMOS_URL=https://my-cosmos.example.com COSMOS_TOKEN=cosmos_xxx \
//     node container-management.mjs [container-name]

import { createClient } from 'cosmos-cloud-sdk';

const cosmos = createClient({
  baseUrl: process.env.COSMOS_URL,
  token: process.env.COSMOS_TOKEN,
});

const TARGET = process.argv[2];

// --- List all containers ---
const containers = await cosmos.docker.list();
console.log('=== All Containers ===');
containers.data.forEach((c) => {
  const name = (c.Names?.[0] || '').replace(/^\//, '');
  const status = c.State === 'running' ? 'RUNNING' : 'STOPPED';
  console.log(`  [${status}] ${name} (${c.Image})`);
});

if (!TARGET) {
  console.log('\nPass a container name as argument to manage it:');
  console.log('  node container-management.mjs my-container');
  process.exit(0);
}

// --- Inspect ---
console.log(`\n=== Container: ${TARGET} ===`);
const info = await cosmos.docker.get(TARGET);
console.log(`  Image:   ${info.data.Config.Image}`);
console.log(`  Status:  ${info.data.State.Status}`);
console.log(`  Started: ${info.data.State.StartedAt}`);

// --- Stop / Start / Restart ---
// Available actions: start, stop, restart, kill, remove, pause, unpause, recreate
console.log('\nStopping...');
await cosmos.docker.manageContainer(TARGET, 'stop');
console.log('  Stopped.');

console.log('Starting...');
await cosmos.docker.manageContainer(TARGET, 'start');
console.log('  Started.');

console.log('Restarting...');
await cosmos.docker.manageContainer(TARGET, 'restart');
console.log('  Restarted.');

// --- Search logs ---
// Note: search query must be at least 3 characters
console.log('\n=== Recent Error Logs ===');
const logs = await cosmos.docker.getContainerLogs(TARGET, 'err', 100);
const logData = logs.data;
if (Array.isArray(logData) && logData.length > 0) {
  logData.slice(-10).forEach((entry) => console.log(`  ${entry.output}`));
} else if (typeof logData === 'string' && logData.length > 0) {
  logData.split('\n').filter(Boolean).slice(-10).forEach((l) => console.log(`  ${l}`));
} else {
  console.log('  No matching logs found.');
}

// --- Update container image ---
// This pulls the latest version of the current image and recreates the container.
console.log('\n=== Updating Container Image ===');
await cosmos.docker.updateContainerImage(TARGET, (line) => {
  console.log(`  ${line}`);
});
console.log('  Updated.');

// --- Toggle auto-update ---
console.log('\nEnabling auto-update...');
await cosmos.docker.autoUpdate(TARGET, 'true');
console.log('  Auto-update enabled.');
