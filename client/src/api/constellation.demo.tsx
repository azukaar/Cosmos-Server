import wrap from './wrap';

function list() {
  return new Promise((resolve, reject) => {
    resolve({
      "data": [
        {
          "nickname": "admin",
          "deviceName": "phone",
          "publicKey": "-----BEGIN NEBULA X25519 PRIVATE KEY-----\naACf/...=\n-----END NEBULA X25519 PRIVATE KEY-----\n",
          "ip": "192.168.201.4/24",
          "isLighthouse": false,
          "isRelay": true,
          "publicHostname": "",
          "port": "4242",
          "blocked": false,
          "fingerprint": "..."
        },
        {
          "nickname": "admin",
          "deviceName": "laptop",
          "publicKey": "-----BEGIN NEBULA X25519 PRIVATE KEY-----\n78l4nDEB0+.../36YBQk7dkwg+.=\n-----END NEBULA X25519 PRIVATE KEY-----\n",
          "ip": "192.168.201.5/24",
          "isLighthouse": false,
          "isRelay": true,
          "publicHostname": "",
          "port": "4242",
          "blocked": false,
          "fingerprint": "..."
        },
        {
          "nickname": "Martha",
          "deviceName": "pink phone",
          "publicKey": "-----BEGIN NEBULA X25519 PRIVATE KEY-----\naACf/..=\n-----END NEBULA X25519 PRIVATE KEY-----\n",
          "ip": "192.168.201.6/24",
          "isLighthouse": false,
          "isRelay": true,
          "publicHostname": "",
          "port": "4242",
          "blocked": false,
          "fingerprint": "..."
        }
      ],
      "status": "OK"
    })
  });
}

function addDevice(device) {
  return new Promise((resolve, reject) => {
    resolve({
      "data": {
        "CA": "-----BEGIN NEBULA CERTIFICATE-----\....\n+dfE+ikL8jUh/n+C+....\....\nZon/Dw==\n-----END NEBULA CERTIFICATE-----\n",
        "Config": "constellation_api_key: ...\nconstellation_device_name: test\nconstellation_local_dns_overwrite: true\nconstellation_local_dns_overwrite_address: 192.168.201.1\nconstellation_public_hostname: \"\"\nfirewall:\n  conntrack:\n    default_timeout: 10m\n    tcp_timeout: 12m\n    udp_timeout: 3m\n  inbound:\n  - host: any\n    port: any\n    proto: any\n  inbound_action: drop\n  outbound:\n  - host: any\n    port: any\n    proto: any\n  outbound_action: drop\nlighthouse:\n  am_lighthouse: false\n  hosts:\n  - 192.168.201.1\n  interval: 60\nlisten:\n  host: 0.0.0.0\n  port: \"4242\"\nlogging:\n  format: text\n  level: info\npki:\n  blocklist: []\n  ca: |\n    -----BEGIN NEBULA CERTIFICATE-----\n    ...\n    +dfE+ikL8jUh/n+C+...\n    .\n    Zon/Dw==\n    -----END NEBULA CERTIFICATE-----\n  cert: |\n    -----BEGIN NEBULA CERTIFICATE-----\n    CmIKBHRlc3QSCoeSo4UMgP7//..\n    ...+QwZSiBxLdKhjkCH+.../..\n    ./hfL+....\n    ..==\n    -----END NEBULA CERTIFICATE-----\n  key: |\n    -----BEGIN NEBULA X25519 PRIVATE KEY-----\n    nS39dWX7uo1rhTvP2yl2XonGx3fWEkpk+43thNrMu7U=\n    -----END NEBULA X25519 PRIVATE KEY-----\npunchy:\n  punch: true\n  respond: true\nrelay:\n  am_relay: false\n  relays:\n  - 192.168.201.1\n  use_relays: true\nstatic_host_map:\n  192.168.201.1:\n  - vpn.domain.com:4242\ntun:\n  dev: nebula1\n  disabled: false\n  drop_local_broadcast: false\n  drop_multicast: false\n  mtu: 1300\n  routes: []\n  tx_queue: 500\n  unsafe_routes: []\n",
        "DeviceName": "test",
        "IP": "192.168.201.7/24",
        "IsLighthouse": false,
        "IsRelay": true,
        "LighthousesList": [],
        "Nickname": "admin",
        "Port": "4242",
        "PrivateKey": "-----BEGIN NEBULA CERTIFICATE-----\...//w8o3ZaFqQYwhdGFuAY6IGXmYRCr3z932Y....w\..==\n-----END NEBULA CERTIFICATE-----\n",
        "PublicHostname": "",
        "PublicKey": "-----BEGIN NEBULA X25519 PRIVATE KEY-----\nnS39dWX...hTvP......+43thNrMu7U=\n-----END NEBULA X25519 PRIVATE KEY-----\n"
      },
      "status": "OK"
    })
  });
}

function restart() {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
    })
  });
}


function reset() {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
    })
  });
}

function getConfig() {
  return new Promise((resolve, reject) => {
    resolve({
      "data": "pki:\n  ca: /config/ca.crt\n  cert: /config/cosmos.crt\n  key: /config/cosmos.key\n  blocklist: []\nstatic_host_map:\n  192.168.201.1:\n  - vpn.domain.com:4242\nlighthouse:\n  am_lighthouse: true\n  interval: 60\n  hosts: []\nlisten:\n  host: 0.0.0.0\n  port: 4242\npunchy:\n  punch: true\n  respond: true\nrelay:\n  am_relay: true\n  use_relays: true\n  relays: []\ntun:\n  disabled: false\n  dev: nebula1\n  drop_local_broadcast: false\n  drop_multicast: false\n  tx_queue: 500\n  mtu: 1300\n  routes: []\n  unsafe_routes: []\nlogging:\n  level: info\n  format: text\nfirewall:\n  outbound_action: drop\n  inbound_action: drop\n  conntrack:\n    tcp_timeout: 12m\n    udp_timeout: 3m\n    default_timeout: 10m\n  outbound:\n  - port: any\n    proto: any\n    host: any\n  inbound:\n  - port: any\n    proto: any\n    host: any\n",
      "status": "OK"
    })
  });
}

function getLogs() {
  return new Promise((resolve, reject) => {
    resolve({
      "data": "Some logs...",
      "status": "OK"
    })
  });
}

function connect(file) {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
    })
  });
}

function block(nickname, devicename, block) {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
    })
  });
}

function ping() {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
      "data": 1
    })
  });
}

export {
  list,
  addDevice,
  restart,
  getConfig,
  getLogs,
  reset,
  connect,
  block,
  ping,
};