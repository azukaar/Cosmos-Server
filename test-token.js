#!/usr/bin/env node
// API Token Integration Test Script
// Usage: node test-token.js <base_url> <admin_token> <readonly_token>
// Example: node test-token.js https://localhost cosmos_abc123 cosmos_xyz789
//
// Requires Node.js 18+ (built-in fetch). No external dependencies.
// Accepts self-signed certs.

process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

const [,, baseUrl, adminToken, readonlyToken] = process.argv;

if (!baseUrl || !adminToken || !readonlyToken) {
  console.error('Usage: node test-token.js <base_url> <admin_token> <readonly_token>');
  process.exit(2);
}

// --- Helpers ---

let passed = 0;
let failed = 0;

async function request(path, method, token, body) {
  const url = `${baseUrl}${path}`;
  const headers = { 'Content-Type': 'application/json' };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  const opts = { method, headers };
  if (body !== undefined) {
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(url, opts);
  let data = null;
  try {
    data = await res.json();
  } catch {
    // response may not be JSON
  }
  return { status: res.status, data };
}

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

async function group1_basicAuth() {
  console.log('\n— Group 1: Basic Auth —');

  for (const [path, label] of [['/cosmos/api/status', 'status'], ['/cosmos/api/me', 'me']]) {
    await test(`${label} - admin token`, async () => {
      expectStatus(await request(path, 'GET', adminToken), 200, 'admin');
    });
    await test(`${label} - readonly token`, async () => {
      expectStatus(await request(path, 'GET', readonlyToken), 200, 'readonly');
    });
    await test(`${label} - no token rejected`, async () => {
      expectStatus(await request(path, 'GET', null), 401, 'no token');
    });
  }
}

async function group2_readEndpoints() {
  console.log('\n— Group 2: Read Endpoints —');

  const endpoints = [
    ['/cosmos/api/config', 'config'],
    ['/cosmos/api/servapps', 'servapps'],
    ['/cosmos/api/users', 'users'],
    ['/cosmos/api/metrics', 'metrics'],
    ['/cosmos/api/api-tokens', 'api-tokens'],
  ];

  for (const [path, label] of endpoints) {
    await test(`GET ${label} - admin token`, async () => {
      expectStatus(await request(path, 'GET', adminToken), 200, 'admin');
    });
    await test(`GET ${label} - readonly token`, async () => {
      expectStatus(await request(path, 'GET', readonlyToken), 200, 'readonly');
    });
    await test(`GET ${label} - no token rejected`, async () => {
      expectStatus(await request(path, 'GET', null), 401, 'no token');
    });
  }
}

async function group3_writeEndpoints() {
  console.log('\n— Group 3: Write Endpoints (admin vs read-only) —');

  // 3a. POST api-tokens (create probe) — admin should succeed, then clean up
  let probeCreated = false;
  try {
    await test('POST api-tokens - admin token (create probe)', async () => {
      const res = await request('/cosmos/api/api-tokens', 'POST', adminToken, { name: '__test_probe__' });
      expectStatus(res, 200, 'admin create');
      probeCreated = true;
    });

    await test('DELETE api-tokens - admin token (cleanup probe)', async () => {
      const res = await request('/cosmos/api/api-tokens', 'DELETE', adminToken, { name: '__test_probe__' });
      expectStatus(res, 200, 'admin delete');
      probeCreated = false;
    });
  } finally {
    // Safety: ensure probe is cleaned up even if DELETE test assertion failed
    if (probeCreated) {
      try {
        await request('/cosmos/api/api-tokens', 'DELETE', adminToken, { name: '__test_probe__' });
      } catch { /* best effort */ }
    }
  }

  // 3b. POST/DELETE api-tokens — readonly should be denied (403)
  await test('POST api-tokens - readonly token (denied)', async () => {
    expectStatus(await request('/cosmos/api/api-tokens', 'POST', readonlyToken, { name: '__test_probe__' }), 403, 'readonly create');
  });

  await test('DELETE api-tokens - readonly token (denied)', async () => {
    expectStatus(await request('/cosmos/api/api-tokens', 'DELETE', readonlyToken, { name: '__test_probe__' }), 403, 'readonly delete');
  });

  // 3c. PUT config — write same config back (admin succeeds, readonly denied)
  // NOTE: This triggers a soft restart on the server when admin succeeds
  await test('PUT config - admin token (idempotent write-back)', async () => {
    const getRes = await request('/cosmos/api/config', 'GET', adminToken);
    expectStatus(getRes, 200, 'get config');
    const config = getRes.data.data;
    const putRes = await request('/cosmos/api/config', 'PUT', adminToken, config);
    expectStatus(putRes, 200, 'admin put');
  });

  await test('PUT config - readonly token (denied)', async () => {
    expectStatus(await request('/cosmos/api/config', 'PUT', readonlyToken, {}), 403, 'readonly put');
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
      expectStatus(await request(path, 'GET', null), 401, 'no token');
    });
  }
}

async function group5_invalidToken() {
  console.log('\n— Group 5: Invalid Token —');

  await test('status - invalid token rejected', async () => {
    expectStatus(await request('/cosmos/api/status', 'GET', 'cosmos_invalidtoken'), 401, 'invalid token');
  });
}

// --- Run ---

async function main() {
  console.log(`Testing API tokens against ${baseUrl}`);
  console.log(`Admin token: ${adminToken.slice(0, 12)}...`);
  console.log(`Readonly token: ${readonlyToken.slice(0, 12)}...`);

  await group1_basicAuth();
  await group2_readEndpoints();
  await group3_writeEndpoints();
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
