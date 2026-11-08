const inDays = (days: number) => new Date(Date.now() + days * 86400000).toISOString();
const ok = (data?: any) => Promise.resolve({ status: 'OK', data });

const demoTokens = {
  'homepage-widget': {
    name: 'homepage-widget',
    description: 'Read-only stats for the dashboard',
    owner: 'admin',
    tokenSuffix: 'a1b2',
    permissions: [1, 20, 30],
    restrictToConstellation: true,
    createdAt: inDays(-90),
  },
  'deploy-script': {
    name: 'deploy-script',
    description: 'Used by the deployment pipeline',
    owner: 'admin',
    tokenSuffix: '9f3c',
    permissions: [2, 21, 31],
    ipWhitelist: ['192.168.1.0/24'],
    restrictToConstellation: false,
    createdAt: inDays(-30),
    expiresAt: inDays(60),
  },
  'old-backup-job': {
    name: 'old-backup-job',
    description: 'Retired backup script',
    owner: 'Martha',
    tokenSuffix: '77de',
    permissions: [1, 20],
    restrictToConstellation: false,
    createdAt: inDays(-400),
    expiresAt: inDays(-35),
  },
};

function list() {
  return ok(demoTokens);
}

function create(request: any) {
  return ok({ token: 'this-is-a-demo-token', name: request.name });
}

function update(name: string, config: any) {
  return ok();
}

function remove(name: string) {
  return ok();
}

export {
  list,
  create,
  update,
  remove,
};
