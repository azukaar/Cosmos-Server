const hoursFromNow = (hours: number) => new Date(Date.now() + hours * 3600000).toISOString();
const ok = (data?: any) => Promise.resolve({ status: 'OK', data });

const STRIKE = 0;
const TEMP = 1;
const PERM = 2;

const demoClients = [
  {
    clientID: '203.0.113.42',
    status: 'perm',
    blockedUntil: '',
    history: [
      { id: 'demo-1', time: hoursFromNow(-30), banType: TEMP, node: 'cosmos', reason: { limit: 'requests', used: 1840, allowed: 1200, route: 'Jellyfin' } },
      { id: 'demo-2', time: hoursFromNow(-20), banType: TEMP, node: 'cosmos', reason: { limit: 'requests', used: 2210, allowed: 1200, route: 'Jellyfin' } },
      { id: 'demo-3', time: hoursFromNow(-8), banType: PERM, node: 'cosmos', reason: { limit: 'escalation', used: 0, allowed: 0, route: 'Jellyfin' } },
    ],
  },
  {
    clientID: '198.51.100.17',
    status: 'temp',
    blockedUntil: hoursFromNow(3),
    history: [
      { id: 'demo-4', time: hoursFromNow(-2), banType: STRIKE, node: 'cosmos', reason: { limit: 'time', used: 412, allowed: 300, route: 'Owncloud' } },
      { id: 'demo-5', time: hoursFromNow(-1), banType: TEMP, node: 'cosmos', reason: { limit: 'simultaneous', used: 64, allowed: 40, route: 'Owncloud' } },
    ],
  },
  {
    clientID: '192.0.2.88',
    status: 'strike',
    blockedUntil: hoursFromNow(0.25),
    history: [
      { id: 'demo-6', time: hoursFromNow(-0.5), banType: STRIKE, node: 'cosmos', reason: { limit: 'bytes', used: 5600000000, allowed: 5000000000, route: 'Navidrome' } },
    ],
  },
  {
    clientID: '2001:db8::7',
    status: 'clear',
    blockedUntil: '',
    history: [
      { id: 'demo-7', time: hoursFromNow(-48), banType: STRIKE, node: 'cosmos', reason: { limit: 'requests', used: 1310, allowed: 1200, route: 'Sonarr' } },
    ],
  },
];

function bans() {
  return ok(demoClients);
}

function unban(clientID: string) {
  return ok();
}

export {
  bans,
  unban,
};
