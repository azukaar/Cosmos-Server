#!/usr/bin/env node
// API Token Integration Test Script using cosmos-cloud-sdk
// Usage: node test-token.js <base_url> <admin_token> <readonly_token>
// Example: node test-token.js https://localhost cosmos_abc123 cosmos_xyz789
//
// Requires Node.js 18+ (built-in fetch). No external dependencies.
// Accepts self-signed certs.

process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

const { createClient } = require('./sdk/dist/index.js');

const [,, baseUrl, adminToken, readonlyToken] = process.argv;

if (!baseUrl || !adminToken || !readonlyToken) {
  console.error('Usage: node test-token.js <base_url> <admin_token> <readonly_token>');
  process.exit(2);
}

// --- Create SDK clients ---

const admin = createClient({ baseUrl, token: adminToken });
const readonly = createClient({ baseUrl, token: readonlyToken });

// Raw request helper (for no-token and invalid-token tests)
async function rawRequest(path, method, token, body) {
  const url = `${baseUrl}${path}`;
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  const opts = { method, headers };
  if (body !== undefined) opts.body = JSON.stringify(body);
  const res = await fetch(url, opts);
  let data = null;
  try { data = await res.json(); } catch { }
  return { status: res.status, data };
}

// --- Test framework ---

let passed = 0;
let failed = 0;

async function test(name, fn) {
  try {
    await fn();
    passed++;
    console.log(`[PASS] ${name}`);
  } catch (err) {
    failed++;
    console.log(`[FAIL] ${name} — ${err.message}`);
  }
}

function expectStatus(res, expected, label) {
  if (res.status !== expected) {
    throw new Error(`${label}: expected ${expected}, got ${res.status}`);
  }
}

// --- Test Groups ---

async function group1_sdkBasicCalls() {
  console.log('\n— Group 1: SDK Basic Calls —');

  await test('admin.getStatus()', async () => {
    const res = await admin.getStatus();
    if (res.status !== 'OK') throw new Error('Expected status OK, got ' + res.status);
  });

  await test('readonly.getStatus()', async () => {
    const res = await readonly.getStatus();
    if (res.status !== 'OK') throw new Error('Expected status OK, got ' + res.status);
  });

  await test('admin.isOnline()', async () => {
    const res = await admin.isOnline();
    if (res.status !== 'OK') throw new Error('Expected status OK');
  });
}

async function group2_sdkReadEndpoints() {
  console.log('\n— Group 2: SDK Read Endpoints —');

  await test('admin.config.get()', async () => {
    const res = await admin.config.get();
    if (res.status !== 'OK') throw new Error('Expected OK');
  });

  await test('readonly.config.get()', async () => {
    const res = await readonly.config.get();
    if (res.status !== 'OK') throw new Error('Expected OK');
  });

  await test('admin.docker.list()', async () => {
    const res = await admin.docker.list();
    if (res.status !== 'OK') throw new Error('Expected OK');
  });

  await test('readonly.docker.list()', async () => {
    const res = await readonly.docker.list();
    if (res.status !== 'OK') throw new Error('Expected OK');
  });

  await test('admin.users.list()', async () => {
    const res = await admin.users.list();
    if (res.status !== 'OK') throw new Error('Expected OK');
  });

  await test('readonly.users.list()', async () => {
    const res = await readonly.users.list();
    if (res.status !== 'OK') throw new Error('Expected OK');
  });

  await test('admin.apiTokens.list()', async () => {
    const res = await admin.apiTokens.list();
    if (res.status !== 'OK') throw new Error('Expected OK');
  });

  await test('readonly.apiTokens.list()', async () => {
    const res = await readonly.apiTokens.list();
    if (res.status !== 'OK') throw new Error('Expected OK');
  });
}

async function group3_sdkWriteEndpoints() {
  console.log('\n— Group 3: SDK Write Endpoints (admin vs read-only) —');

  // admin: create + delete probe token
  let probeCreated = false;
  try {
    await test('admin.apiTokens.create() - probe', async () => {
      const res = await admin.apiTokens.create({ name: '__test_probe__' });
      if (res.status !== 'OK') throw new Error('Expected OK, got ' + res.status);
      probeCreated = true;
    });

    await test('admin.apiTokens.remove() - cleanup probe', async () => {
      const res = await admin.apiTokens.remove('__test_probe__');
      if (res.status !== 'OK') throw new Error('Expected OK');
      probeCreated = false;
    });
  } finally {
    if (probeCreated) {
      try { await admin.apiTokens.remove('__test_probe__'); } catch { }
    }
  }

  // readonly: create should fail
  await test('readonly.apiTokens.create() - denied', async () => {
    try {
      await readonly.apiTokens.create({ name: '__test_probe__' });
      throw new Error('Should have thrown');
    } catch (e) {
      if (!e.code && !e.status) throw e;
      // Expected: error thrown (403)
    }
  });

  // readonly: delete should fail
  await test('readonly.apiTokens.remove() - denied', async () => {
    try {
      await readonly.apiTokens.remove('__test_probe__');
      throw new Error('Should have thrown');
    } catch (e) {
      if (!e.code && !e.status) throw e;
    }
  });

  // admin: idempotent config write-back
  await test('admin.config.set() - idempotent write-back', async () => {
    const getRes = await admin.config.get();
    const config = getRes.data;
    const putRes = await admin.config.set(config);
    if (putRes.status !== 'OK') throw new Error('Expected OK');
  });

  // readonly: config set denied
  await test('readonly.config.set() - denied', async () => {
    try {
      await readonly.config.set({});
      throw new Error('Should have thrown');
    } catch (e) {
      if (!e.code && !e.status) throw e;
    }
  });
}

async function group4_noToken() {
  console.log('\n— Group 4: No-Token Rejection —');

  for (const [path, label] of [
    ['/cosmos/api/config', 'config'],
    ['/cosmos/api/servapps', 'servapps'],
    ['/cosmos/api/users', 'users'],
  ]) {
    await test(`GET ${label} - no token rejected`, async () => {
      expectStatus(await rawRequest(path, 'GET', null), 401, 'no token');
    });
  }
}

async function group5_invalidToken() {
  console.log('\n— Group 5: Invalid Token —');

  await test('status - invalid token rejected', async () => {
    expectStatus(await rawRequest('/cosmos/api/status', 'GET', 'cosmos_invalidtoken'), 401, 'invalid token');
  });
}

// --- Run ---

async function main() {
  console.log(`Testing API tokens via cosmos-cloud-sdk against ${baseUrl}`);
  console.log(`Admin token: ${adminToken.slice(0, 12)}...`);
  console.log(`Readonly token: ${readonlyToken.slice(0, 12)}...`);

  await group1_sdkBasicCalls();
  await group2_sdkReadEndpoints();
  await group3_sdkWriteEndpoints();
  await group4_noToken();
  await group5_invalidToken();

  console.log(`\nResults: ${passed}/${passed + failed} passed`);
  if (failed > 0) {
    console.log(`${failed} test(s) FAILED`);
    process.exit(1);
  }
}

main().catch(err => {
  console.error('Fatal error:', err);
  process.exit(1);
});
