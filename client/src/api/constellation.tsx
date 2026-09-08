import wrap, { type ApiResponse, type ApiFetch } from './wrap';

export interface ConstellationDevice {
  Nickname: string;
  DeviceName: string;
  IP: string;
  IsLighthouse: boolean;
  IsRelay: boolean;
  Blocked: boolean;
  Fingerprint: string;
  PublicHostname: string;
  Port: string;
  [key: string]: any;
}

// `live` is heartbeat-derived: only the server knows whether the node is up.
export interface TagMatchNode {
  name: string;
  live: boolean;
  tags: string[];
}

export interface TagMatchResult {
  tags: string[];
  matched: TagMatchNode[];
  matchedCount: number;
  liveMatchedCount: number;
  totalNodes: number;
  liveNodes: number;
}

export default function createConstellationAPI(apiFetch: ApiFetch) {
  function list() {
    return wrap(apiFetch('/cosmos/api/constellation/devices', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function ping() {
    return wrap(apiFetch('/cosmos/api/constellation/ping', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function natsStatus() {
    return wrap(apiFetch('/cosmos/api/constellation/nats-status', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function getNextIP() {
    return wrap(apiFetch('/cosmos/api/constellation/get-next-ip', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function pingDevice(deviceId) {
    return wrap(apiFetch(`/cosmos/api/constellation/devices/${deviceId}/ping`, {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function addDevice(device) {
    return wrap(apiFetch('/cosmos/api/constellation/devices', {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify(device),
    }))
  }

  function resyncDevice(device) {
    return wrap(apiFetch('/cosmos/api/constellation/config-manual-sync', {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify(device),
    }))
  }

  function restart() {
    return wrap(apiFetch('/cosmos/api/constellation/restart', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      }
    }))
  }


  function reset() {
    return wrap(apiFetch('/cosmos/api/constellation/reset', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      }
    }))
  }

  // HA only: re-form the constellation around this manager after a lost quorum
  function forceReform() {
    return wrap(apiFetch('/cosmos/api/constellation/force-reform', {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      }
    }))
  }

  function getConfig() {
    return wrap(apiFetch('/cosmos/api/constellation/config', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      }
    }))
  }

  function getLogs() {
    return wrap(apiFetch('/cosmos/api/constellation/logs', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      }
    }))
  }

  function connect(file) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();

      reader.onload = () => {
        apiFetch('/cosmos/api/constellation/connect', {
          method: 'POST',
          headers: {
            'Content-Type': 'text/plain',
          },
          body: reader.result,
        })
        .then(response => {
          resolve(response);
        })
        .catch(error => {
          reject(error);
        });
      };

      reader.onerror = () => {
        reject(new Error('Failed to read the file.'));
      };

      reader.readAsText(file);
    });
  }

  function create(deviceName, isLighthouse, hostname, ipRange, natsReplicas) {
    return wrap(apiFetch(`/cosmos/api/constellation/create`, {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        deviceName,
        isLighthouse,
        hostname,
        ipRange,
        natsReplicas,
      }),
    }))
  }

  function block(nickname, devicename, block) {
    return wrap(apiFetch(`/cosmos/api/constellation/block`, {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        nickname, devicename, block
      }),
    }))
  }

  function tunnels() {
    return wrap(apiFetch('/cosmos/api/constellation/tunnels', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function editDevice(device) {
    return wrap(apiFetch('/cosmos/api/constellation/edit-device', {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify(device),
    }))
  }

  // --- DNS Entry CRUD ---
  function listDNSEntries(): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/constellation/dns', {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }))
  }

  function createDNSEntry(entry: { Type: string; Key: string; Value: string }): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/constellation/dns', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(entry),
    }))
  }

  function updateDNSEntry(key: string, entry: { Type: string; Key: string; Value: string }): Promise<ApiResponse> {
    return wrap(apiFetch(`/cosmos/api/constellation/dns/${encodeURIComponent(key)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(entry),
    }))
  }

  function deleteDNSEntry(key: string): Promise<ApiResponse> {
    return wrap(apiFetch(`/cosmos/api/constellation/dns/${encodeURIComponent(key)}`, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
    }))
  }

  // noError on the wrap: a transient 503 (JetStream reconnecting) must not pop a snackbar per keystroke.
  function tagNodes(tags: string[]): Promise<ApiResponse<TagMatchResult>> {
    const query = (tags || []).join(',');
    return wrap(apiFetch('/cosmos/api/constellation/tag-nodes?tags=' + encodeURIComponent(query), {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }), true)
  }

  return {
    list,
    addDevice,
    resyncDevice,
    restart,
    getConfig,
    getLogs,
    reset,
    forceReform,
    connect,
    block,
    ping,
    natsStatus,
    create,
    pingDevice,
    tunnels,
    editDevice,
    getNextIP,
    listDNSEntries,
    createDNSEntry,
    updateDNSEntry,
    deleteDNSEntry,
    tagNodes,
  };
}
