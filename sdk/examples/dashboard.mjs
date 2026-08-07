// Dashboard Example
// Displays an overview of your Cosmos server: containers, metrics, and recent events.
//
// Usage:
//   COSMOS_URL=https://my-cosmos.example.com COSMOS_TOKEN=cosmos_xxx node dashboard.mjs

import { createClient } from 'cosmos-cloud-sdk';

const cosmos = createClient({
  baseUrl: process.env.COSMOS_URL,
  token: process.env.COSMOS_TOKEN,
});

// --- Server status ---
const status = await cosmos.getStatus();
console.log('=== Server Status ===');
console.log(`  Hostname: ${status.data.hostname}`);
console.log(`  Docker:   ${status.data.docker ? 'connected' : 'not connected'}`);
console.log();

// --- Running containers ---
const containers = await cosmos.docker.list();
console.log('=== Containers ===');
console.table(
  containers.data.map((c) => ({
    Name: (c.Names?.[0] || '').replace(/^\//, ''),
    Image: c.Image,
    Status: c.Status,
    Running: c.State === 'running' ? 'Yes' : 'No',
  }))
);

// --- Available metrics ---
const metricList = await cosmos.metrics.list();
const metricNames = Object.keys(metricList.data);
console.log(`\n=== Available Metrics (${metricNames.length}) ===`);
console.log(metricNames.join(', '));

// Fetch a few metrics if available
if (metricNames.length > 0) {
  const toFetch = metricNames.slice(0, 5);
  const metrics = await cosmos.metrics.get(toFetch);
  console.log('\n=== Metric Samples ===');
  for (const m of metrics.data) {
    const latest = Array.isArray(m.Values) && m.Values.length > 0
      ? JSON.stringify(m.Values[m.Values.length - 1])
      : '(no data)';
    console.log(`  ${m.Key}: ${latest}`);
  }
}

// --- Recent events ---
const now = new Date();
const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
const events = await cosmos.metrics.events(
  oneHourAgo.toISOString(),
  now.toISOString(),
);
console.log(`\n=== Recent Events (last hour) ===`);
const eventList = Array.isArray(events.data) ? events.data : [];
if (eventList.length === 0) {
  console.log('  No recent events.');
} else {
  eventList.slice(0, 10).forEach((e) => {
    console.log(`  [${e.Date || ''}] ${e.Label || e.Message || JSON.stringify(e)}`);
  });
}

// --- Container logs (first container) ---
if (containers.data.length > 0) {
  const first = containers.data[0];
  const name = (first.Names?.[0] || '').replace(/^\//, '');
  console.log(`\n=== Last Logs: ${name} ===`);
  const logs = await cosmos.docker.getContainerLogs(name, '', 50);
  const logData = logs.data;
  if (Array.isArray(logData)) {
    logData.slice(-10).forEach((entry) => console.log(`  ${entry.output}`));
  } else if (typeof logData === 'string') {
    logData.split('\n').filter(Boolean).slice(-10).forEach((l) => console.log(`  ${l}`));
  }
}
