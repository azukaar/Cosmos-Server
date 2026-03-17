import wrap, { wrapRClone, type ApiResponse, type ApiFetch } from './wrap';

function _setData(data: any): any {
  if(!data) {
    data = {};
  }

  if(!data.parameters) {
    data.parameters = {};
  }

  data.parameters.config_is_local = false;
  data.parameters.config_refresh_token = false;

  return data;
}

export default function createRcloneAPI(apiFetch: ApiFetch) {
  function save(): Promise<any> {
    return wrapRClone(apiFetch('/cosmos/rclone/config/save', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{}'
    }))
  }

  function create(data: any): Promise<any> {
    return wrapRClone(apiFetch('/cosmos/rclone/config/create', {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify(_setData(data))
    }))
  }

  function list(): Promise<any> {
    return wrapRClone(apiFetch('/cosmos/rclone/config/dump', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({})
    }))
  }

  function listRemotes(): Promise<any> {
    return wrapRClone(apiFetch('/cosmos/rclone/config/listremotes', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({})
    }))
  }

  function update(data: any): Promise<any> {
    return wrapRClone(apiFetch('/cosmos/rclone/config/update', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(_setData(data))
    }))
  }

  function pingStorage(remoteName: string): Promise<any> {
    return wrapRClone(apiFetch('/cosmos/rclone/operations/about', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        "fs": remoteName + ":/",
      })
    }))
  }

  function stats(remoteName: string): Promise<any> {
    return wrapRClone(apiFetch('/cosmos/rclone/vfs/stats', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        "fs": remoteName + ":"
      })
    }))
  }

  function coreStats(remoteName: string, remotePath?: string): Promise<any> {
    return wrapRClone(apiFetch('/cosmos/rclone/core/stats', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: "{}"
    }))
  }

  function deleteRemote(remoteName: string): Promise<any> {
    return wrapRClone(apiFetch('/cosmos/rclone/config/delete', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        name: remoteName
      })
    }))
  }

  function restart(): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/rclone-restart', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    }))
  }

  return {
    create,
    list,
    listRemotes,
    deleteRemote,
    update,
    coreStats,
    stats,
    pingStorage,
    restart
  };
}
