// CI/CD Deploy Example
// Deploys a compose-style service through the Cosmos API.
// The createService endpoint automatically pulls images, creates networks/volumes,
// and starts containers in dependency order.
//
// Usage:
//   COSMOS_URL=https://my-cosmos.example.com COSMOS_TOKEN=cosmos_xxx \
//     node cicd-deploy.mjs
//
// Environment variables:
//   COSMOS_URL    - Your Cosmos server URL
//   COSMOS_TOKEN  - API token (starts with cosmos_)
//   DEPLOY_IMAGE  - Image to deploy (default: nginx:latest)

import { createClient } from 'cosmos-cloud-sdk';

const cosmos = createClient({
  baseUrl: process.env.COSMOS_URL,
  token: process.env.COSMOS_TOKEN,
});

const IMAGE = process.env.DEPLOY_IMAGE || 'nginx:latest';
const serviceName = IMAGE.split(':')[0].split('/').pop();

// --- Deploy a compose service ---
// createService handles everything: pulling images, creating networks/volumes,
// starting containers, and setting up routes. No need to pull images separately.
console.log(`[1/2] Deploying ${IMAGE}...`);

const serviceDefinition = {
  services: {
    [serviceName]: {
      image: IMAGE,
      container_name: serviceName,
      restart: 'unless-stopped',
      labels: {},
      // Add your own configuration:
      // environment: ['NODE_ENV=production'],
      // volumes: [{ type: 'bind', source: '/host/data', target: '/data' }],
      // ports: ['8080:80'],
    },
  },
  // networks: {
  //   'my-net': { driver: 'bridge' },
  // },
  // volumes: {
  //   'my-data': { driver: 'local' },
  // },
};

await cosmos.docker.createService(serviceDefinition, (line) => {
  console.log(`  ${line}`);
});
console.log('  Done.');

// --- Verify ---
console.log(`[2/2] Verifying...`);
await new Promise((r) => setTimeout(r, 2000));

const containers = await cosmos.docker.list();
console.log(containers);
const deployed = containers.data.find(
  (c) => c.Image === IMAGE || c.Names?.some(n => n.includes(serviceName)),
);

if (deployed && deployed.State === 'running') {
  console.log(`  Container "${deployed.Names[0].replace(/^\//, '')}" is running.`);
  process.exit(0);
} else {
  console.error('  Container not found or not running.');
  process.exit(1);
}
