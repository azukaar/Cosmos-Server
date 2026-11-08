import configDemo from './demo.config.json';

const http = configDemo.data.HTTPConfig as any;
const inDays = (days: number) => new Date(Date.now() + days * 86400000).toISOString();
const ok = (data?: any) => Promise.resolve({ status: 'OK', data });

const zoneDefaults = {
  derived: false,
  issuesCertificate: true,
  recordsSupported: false,
  TLSCert: '',
  UseWildcardCertificate: false,
  DNSChallengeProvider: '',
  DNSChallengeConfig: {},
  DNSChallengeResolvers: '',
  DNSChallengePropagationWait: 0,
  DisablePropagationChecks: false,
  ManageRecords: false,
  WildcardRecord: false,
};

const certsFor = (hosts: string[], cert: any) => Object.fromEntries(hosts.map((host) => [host, cert]));

const routeHosts: string[] = http.ProxyConfig.Routes.map((route: any) => route.Host).filter(Boolean);
const mainCertificate = {
  hosts: [http.Hostname, '*.' + http.Hostname],
  validUntil: inDays(63),
  issuedBy: "Let's Encrypt",
};

const wildcardHosts = ['yann-server.com', 'cloud.yann-server.com', 'vault.yann-server.com'];
const wildcardCertificate = {
  hosts: ['yann-server.com', '*.yann-server.com'],
  validUntil: inDays(41),
  issuedBy: "Let's Encrypt",
};

const labHosts = ['nas.lab.example.org'];
const pendingHosts = ['stream.media.example.net'];

const demoZones = [
  {
    ...zoneDefaults,
    Zone: http.Hostname,
    derived: true,
    HTTPSCertificateMode: http.HTTPSCertificateMode,
    DNSChallengeProvider: http.DNSChallengeProvider || '',
    hosts: routeHosts,
    certificate: mainCertificate,
    hostCertificates: certsFor(routeHosts, { source: 'zone', zone: http.Hostname, covered: true, ...mainCertificate }),
  },
  {
    ...zoneDefaults,
    Zone: 'yann-server.com',
    HTTPSCertificateMode: 'LETSENCRYPT',
    UseWildcardCertificate: true,
    DNSChallengeProvider: 'cloudflare',
    ManageRecords: true,
    WildcardRecord: true,
    recordsSupported: true,
    hosts: wildcardHosts,
    certificate: wildcardCertificate,
    hostCertificates: certsFor(wildcardHosts, { source: 'zone', zone: 'yann-server.com', covered: true, ...wildcardCertificate }),
    records: {
      supported: true,
      lastSync: inDays(-0.01),
      records: {
        'yann-server.com': ['199.199.199.199'],
        '*.yann-server.com': ['199.199.199.199'],
      },
    },
  },
  {
    ...zoneDefaults,
    Zone: 'lab.example.org',
    HTTPSCertificateMode: 'SELFSIGNED',
    issuesCertificate: false,
    hosts: labHosts,
    hostCertificates: certsFor(labHosts, { source: 'selfsigned', hosts: labHosts, validUntil: inDays(300), covered: true }),
  },
  {
    ...zoneDefaults,
    Zone: 'media.example.net',
    HTTPSCertificateMode: 'LETSENCRYPT',
    DNSChallengeProvider: 'ovh',
    ManageRecords: true,
    recordsSupported: true,
    hosts: pendingHosts,
    hostCertificates: certsFor(pendingHosts, { source: 'none', hosts: [], validUntil: '0001-01-01T00:00:00Z', covered: false }),
    records: {
      supported: true,
      lastSync: inDays(-0.2),
      lastError: 'ovh: authentication failed (403)',
      records: {},
    },
  },
];

function list() {
  return ok(demoZones);
}

function set(zone: string, config: any) {
  return ok();
}

function remove(zone: string) {
  return ok();
}

function addresses() {
  return ok([
    { value: '', address: '199.199.199.199' },
    { value: 'iface:eth0', address: '192.168.1.20' },
    { value: 'iface:nebula1', address: '192.168.201.1' },
  ]);
}

export {
  addresses,
  list,
  set,
  remove,
};
