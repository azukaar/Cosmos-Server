import wrap, {wrapRClone} from './wrap';

function _setData(data) {
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

function create(data) {
  return wrapRClone(fetch('/cosmos/rclone/config/create', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(_setData(data))
  }))
}

function list() {
  return wrapRClone(fetch('/cosmos/rclone/config/dump', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({}) // Empty body for POST request
  }))
}

// New function to update a provider
function update(data) {
  return wrapRClone(fetch('/cosmos/rclone/config/update', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(_setData(data))
  }))
}

// New function to ping a storage
function pingStorage(remoteName) {
  return wrapRClone(fetch('/cosmos/rclone/operations/about', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "fs": remoteName + ":/",
    })
  }))
}

// New function to ping a storage
function stats(remoteName) {
  return wrapRClone(fetch('/cosmos/rclone/vfs/stats', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      "fs": remoteName + ":"
    })
  }))
}

// New function to ping a storage
function coreStats(remoteName, remotePath) {
  return wrapRClone(fetch('/cosmos/rclone/core/stats', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: "{}"
  }))
}

function deleteRemote(remoteName) {
  return wrapRClone(fetch('/cosmos/rclone/config/delete', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      name: remoteName
    })
  }))
}

function restart() {
  return wrap(fetch('/cosmos/api/rclone-restart', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }))
}

export {
  create,
  list,
  deleteRemote,
  update,
  coreStats,
  stats,
  pingStorage,
  restart
};